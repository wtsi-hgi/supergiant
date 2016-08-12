package core

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/pkg/models"
	"github.com/supergiant/supergiant/pkg/util"
)

// TODO
var globalK8SHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func (c *Core) K8S(m *models.Kube) guber.Client {
	return guber.NewClient(m.Config.MasterPublicIP, m.Username, m.Password, globalK8SHTTPClient)
}

//------------------------------------------------------------------------------

var AWSMasterAMIs = map[string]string{
	"ap-northeast-1": "ami-907fa690",
	"ap-southeast-1": "ami-b4a79de6",
	"eu-central-1":   "ami-e8635bf5",
	"eu-west-1":      "ami-0fd0ae78",
	"sa-east-1":      "ami-f9f675e4",
	"us-east-1":      "ami-f57b8f9e",
	"us-west-1":      "ami-87b643c3",
	"cn-north-1":     "ami-3abf2203",
	"ap-southeast-2": "ami-1bb9c221",
	"us-west-2":      "ami-33566d03",
}

type Kubes struct {
	Collection
}

func (c *Kubes) Create(m *models.Kube) error {

	// Set some defaults
	if m.Config != nil {
		if len(m.Config.InstanceTypes) == 0 {
			m.Config.InstanceTypes = []string{
				"m4.large",
				"m4.xlarge",
				"m4.2xlarge",
				"m4.4xlarge",
			}
		}
		if m.Username == "" && m.Password == "" {
			m.Username = util.RandomString(16)
			m.Password = util.RandomString(8)
		}
	}

	if err := c.Collection.Create(m); err != nil {
		return err
	}
	provision := &Action{
		Status: &models.ActionStatus{
			Description: "provisioning",
			MaxRetries:  20,
		},
		core:       c.core,
		resourceID: m.UUID,
		model:      m,
		fn: func(a *Action) error {
			return c.provision(m, a)
		},
	}
	return provision.Async()
}

func (c *Kubes) Delete(id *int64, m *models.Kube) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "deleting",
			MaxRetries:  5,
		},
		core:           c.core,
		scope:          c.core.DB.Preload("CloudAccount").Preload("Entrypoints").Preload("Volumes.Kube.CloudAccount").Preload("Apps").Preload("Nodes.Kube.CloudAccount"),
		model:          m,
		id:             id,
		cancelExisting: true,
		fn: func(_ *Action) error {
			for _, entrypoint := range m.Entrypoints {
				if err := c.core.Entrypoints.Delete(entrypoint.ID, entrypoint).Now(); err != nil {
					return err
				}
			}
			// Delete nodes first to get rid of any potential hanging volumes
			for _, node := range m.Nodes {
				if err := c.core.Nodes.Delete(node.ID, node).Now(); err != nil {
					return err
				}
			}
			for _, app := range m.Apps {
				if err := c.core.Apps.Delete(app.ID, app).Now(); err != nil {
					return err
				}
			}
			if err := c.teardown(m); err != nil {
				return err
			}
			return c.Collection.Delete(id, m)
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Kubes) servers(m *models.Kube) (instances []*ec2.Instance, err error) {
	return c.filteredServers(m, map[string][]string{
		"tag:aws:autoscaling:groupName": m.AutoScalingGroupNames(),
	})
}

func (c *Kubes) server(m *models.Kube, id string) (*ec2.Instance, error) {
	instances, err := c.filteredServers(m, map[string][]string{
		"instance-id": []string{id},
	})
	if err != nil {
		return nil, err
	}
	return instances[0], nil
}

func (c *Kubes) filteredServers(m *models.Kube, filters map[string][]string) (instances []*ec2.Instance, err error) {
	filters["instance-state-name"] = []string{
		"pending",
		"running",
	}
	var ec2Filters []*ec2.Filter
	for key, vals := range filters {
		ec2Filter := &ec2.Filter{
			Name: aws.String(key),
		}
		for _, val := range vals {
			ec2Filter.Values = append(ec2Filter.Values, aws.String(val))
		}
		ec2Filters = append(ec2Filters, ec2Filter)
	}
	input := &ec2.DescribeInstancesInput{
		Filters: ec2Filters,
	}
	resp, err := c.core.CloudAccounts.ec2(m.CloudAccount, m.Config.Region).DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

func (c *Kubes) autoScalingGroup(m *models.Kube, instanceType string) (*autoscaling.Group, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(m.AutoScalingGroupName(instanceType))},
	}
	resp, err := c.core.CloudAccounts.autoscaling(m.CloudAccount, m.Config.Region).DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, err
	}
	return resp.AutoScalingGroups[0], nil
}

type provisioner struct {
	core  *Core
	kube  *models.Kube
	steps []*step
}

type step struct {
	desc string
	fn   func() error
}

func (p *provisioner) addStep(desc string, fn func() error) {
	p.steps = append(p.steps, &step{desc, fn})
}

func (p *provisioner) run() error {
	for _, step := range p.steps {
		p.core.Log.Infof("Running step of Kube provisioner: %s", step.desc)
		if err := step.fn(); err != nil {
			return err
		}
		if err := p.core.DB.Save(p.kube); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

// is it NOT Not Found
func isErrAndNotAWSNotFound(err error) bool {
	return err != nil && !regexp.MustCompile(`([Nn]ot *[Ff]ound|404)`).MatchString(err.Error())
}

func createIAMRole(iamS *iam.IAM, name string, policy string) error {
	getInput := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}
	_, err := iamS.GetRole(getInput)
	if err == nil {
		return nil
	} else if isErrAndNotAWSNotFound(err) {
		return err
	}
	input := &iam.CreateRoleInput{
		RoleName: aws.String(name),
		Path:     aws.String("/"),
		AssumeRolePolicyDocument: aws.String(policy),
	}
	_, err = iamS.CreateRole(input)
	return err
}

func createIAMRolePolicy(iamS *iam.IAM, name string, policy string) error {
	getInput := &iam.GetRolePolicyInput{
		RoleName:   aws.String(name),
		PolicyName: aws.String(name),
	}
	_, err := iamS.GetRolePolicy(getInput)
	if err == nil {
		return nil
	} else if isErrAndNotAWSNotFound(err) {
		return err
	}

	putRoleInput := &iam.PutRolePolicyInput{
		RoleName:       aws.String(name),
		PolicyName:     aws.String(name),
		PolicyDocument: aws.String(policy),
	}
	_, err = iamS.PutRolePolicy(putRoleInput)
	return err
}

func createIAMInstanceProfile(iamS *iam.IAM, name string) error {
	getInput := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}
	_, err := iamS.GetInstanceProfile(getInput)
	if err == nil {
		return nil
	} else if isErrAndNotAWSNotFound(err) {
		return err
	}

	input := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(name),
		Path:                aws.String("/"),
	}
	_, err = iamS.CreateInstanceProfile(input)
	if err != nil {
		return err
	}

	addInput := &iam.AddRoleToInstanceProfileInput{
		RoleName:            aws.String(name),
		InstanceProfileName: aws.String(name),
	}
	_, err = iamS.AddRoleToInstanceProfile(addInput)
	return err
}

func tagAWSResource(ec2S *ec2.EC2, idstr string, tags map[string]string) error {
	var ec2Tags []*ec2.Tag
	for key, val := range tags {
		ec2Tags = append(ec2Tags, &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		})
	}
	input := &ec2.CreateTagsInput{
		Resources: []*string{aws.String(idstr)},
		Tags:      ec2Tags,
	}
	_, err := ec2S.CreateTags(input)
	return err
}

//------------------------------------------------------------------------------

func (c *Kubes) provision(m *models.Kube, action *Action) error {
	iamS := c.core.CloudAccounts.iam(m.CloudAccount, m.Config.Region)
	ec2S := c.core.CloudAccounts.ec2(m.CloudAccount, m.Config.Region)
	autoscalingS := c.core.CloudAccounts.autoscaling(m.CloudAccount, m.Config.Region)
	p := &provisioner{core: c.core, kube: m}

	p.addStep("preparing IAM Role kubernetes-master", func() error {
		policy := `{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Action": "sts:AssumeRole",
          "Principal": {"AWS": "*"},
          "Effect": "Allow",
          "Sid": ""
        }
      ]
    }`
		return createIAMRole(iamS, "kubernetes-master", policy)
	})

	p.addStep("preparing IAM Role Policy kubernetes-master", func() error {
		policy := `{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": ["ec2:*"],
          "Resource": ["*"]
        },
        {
          "Effect": "Allow",
          "Action": ["elasticloadbalancing:*"],
          "Resource": ["*"]
        },
        {
          "Effect": "Allow",
          "Action": "s3:*",
          "Resource": [
            "arn:aws:s3:::kubernetes-*"
          ]
        }
      ]
    }`
		return createIAMRolePolicy(iamS, "kubernetes-master", policy)
	})

	p.addStep("preparing IAM Instance Profile kubernetes-minion", func() error {
		return createIAMInstanceProfile(iamS, "kubernetes-minion")
	})

	p.addStep("preparing IAM Role kubernetes-minion", func() error {
		policy := `{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Action": "sts:AssumeRole",
          "Principal": {"AWS": "*"},
          "Effect": "Allow",
          "Sid": ""
        }
      ]
    }`
		return createIAMRole(iamS, "kubernetes-minion", policy)
	})

	p.addStep("preparing IAM Role Policy kubernetes-minion", func() error {
		policy := `{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": "s3:*",
          "Resource": [
            "arn:aws:s3:::kubernetes-*"
          ]
        },
        {
          "Effect": "Allow",
          "Action": "ec2:Describe*",
          "Resource": "*"
        },
        {
          "Effect": "Allow",
          "Action": "ec2:AttachVolume",
          "Resource": "*"
        },
        {
          "Effect": "Allow",
          "Action": "ec2:DetachVolume",
          "Resource": "*"
        }
      ]
    }`
		return createIAMRolePolicy(iamS, "kubernetes-minion", policy)
	})

	p.addStep("preparing IAM Instance Profile kubernetes-minion", func() error {
		return createIAMInstanceProfile(iamS, "kubernetes-minion")
	})

	p.addStep("creating SSH Key Pair", func() error {
		if m.Config.PrivateKey != "" {
			return nil
		}
		input := &ec2.CreateKeyPairInput{
			KeyName: aws.String(m.Name + "-key"),
		}
		resp, err := ec2S.CreateKeyPair(input)
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate") {

				delInput := &ec2.DeleteKeyPairInput{
					KeyName: aws.String(m.Name + "-key"),
				}
				_, _ = ec2S.DeleteKeyPair(delInput)
				return errors.New("KeyPair existed, but key material was not captured. Deleted KeyPair... will retry")

			}
			return err
		}
		m.Config.PrivateKey = *resp.KeyMaterial
		return nil
	})

	p.addStep("creating VPC", func() error {
		if m.Config.VPCID != "" {
			return nil
		}
		input := &ec2.CreateVpcInput{
			CidrBlock: aws.String(m.Config.VPCIPRange),
		}
		resp, err := ec2S.CreateVpc(input)
		if err != nil {
			return err
		}
		m.Config.VPCID = *resp.Vpc.VpcId
		return nil
	})

	p.addStep("tagging VPC", func() error {
		return tagAWSResource(ec2S, m.Config.VPCID, map[string]string{
			"KubernetesCluster": m.Name,
			"Name":              m.Name + "-vpc",
		})
	})

	p.addStep("enabling VPC DNS", func() error {
		input := &ec2.ModifyVpcAttributeInput{
			VpcId:              aws.String(m.Config.VPCID),
			EnableDnsHostnames: &ec2.AttributeBooleanValue{Value: aws.Bool(true)},
		}
		_, err := ec2S.ModifyVpcAttribute(input)
		return err
	})

	// Create Internet Gateway

	p.addStep("creating Internet Gateway", func() error {
		if m.Config.InternetGatewayID != "" {
			return nil
		}
		resp, err := ec2S.CreateInternetGateway(new(ec2.CreateInternetGatewayInput))
		if err != nil {
			return err
		}
		m.Config.InternetGatewayID = *resp.InternetGateway.InternetGatewayId
		return nil
	})

	p.addStep("tagging Internet Gateway", func() error {
		return tagAWSResource(ec2S, m.Config.InternetGatewayID, map[string]string{
			"KubernetesCluster": m.Name,
			"Name":              m.Name + "-ig",
		})
	})

	p.addStep("attaching Internet Gateway to VPC", func() error {
		input := &ec2.AttachInternetGatewayInput{
			VpcId:             aws.String(m.Config.VPCID),
			InternetGatewayId: aws.String(m.Config.InternetGatewayID),
		}
		if _, err := ec2S.AttachInternetGateway(input); err != nil && !strings.Contains(err.Error(), "already attached") {
			return err
		}
		return nil
	})

	// Create Subnet

	p.addStep("creating Subnet", func() error {
		if m.Config.PublicSubnetID != "" {
			return nil
		}
		input := &ec2.CreateSubnetInput{
			VpcId:            aws.String(m.Config.VPCID),
			CidrBlock:        aws.String(m.Config.PublicSubnetIPRange),
			AvailabilityZone: aws.String(m.Config.AvailabilityZone),
		}
		resp, err := ec2S.CreateSubnet(input)
		if err != nil {
			return err
		}
		m.Config.PublicSubnetID = *resp.Subnet.SubnetId
		return nil
	})

	p.addStep("tagging Subnet", func() error {
		return tagAWSResource(ec2S, m.Config.PublicSubnetID, map[string]string{
			"KubernetesCluster": m.Name,
			"Name":              m.Name + "-psub",
		})
	})

	// Route Table

	p.addStep("creating Route Table", func() error {
		if m.Config.RouteTableID != "" {
			return nil
		}
		input := &ec2.CreateRouteTableInput{
			VpcId: aws.String(m.Config.VPCID),
		}
		resp, err := ec2S.CreateRouteTable(input)
		if err != nil {
			return err
		}
		m.Config.RouteTableID = *resp.RouteTable.RouteTableId
		return nil
	})

	p.addStep("tagging Route Table", func() error {
		return tagAWSResource(ec2S, m.Config.RouteTableID, map[string]string{
			"KubernetesCluster": m.Name,
			"Name":              m.Name + "-rt",
		})
	})

	p.addStep("associating Route Table with Subnet", func() error {
		if m.Config.RouteTableSubnetAssociationID != "" {
			return nil
		}
		input := &ec2.AssociateRouteTableInput{
			RouteTableId: aws.String(m.Config.RouteTableID),
			SubnetId:     aws.String(m.Config.PublicSubnetID),
		}
		resp, err := ec2S.AssociateRouteTable(input)
		if err != nil {
			return err
		}
		m.Config.RouteTableSubnetAssociationID = *resp.AssociationId
		return nil
	})

	p.addStep("creating Route for Internet Gateway", func() error {
		input := &ec2.CreateRouteInput{
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
			RouteTableId:         aws.String(m.Config.RouteTableID),
			GatewayId:            aws.String(m.Config.InternetGatewayID),
		}
		if _, err := ec2S.CreateRoute(input); err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
			return err
		}
		return nil
	})

	// Create Security Groups

	p.addStep("creating ELB Security Group", func() error {
		if m.Config.ELBSecurityGroupID != "" {
			return nil
		}
		input := &ec2.CreateSecurityGroupInput{
			GroupName:   aws.String(m.Name + "_elb_sg"),
			Description: aws.String("Allow any external port through to internal 30-40k range"),
			VpcId:       aws.String(m.Config.VPCID),
		}
		resp, err := ec2S.CreateSecurityGroup(input)
		if err != nil {
			return err
		}
		m.Config.ELBSecurityGroupID = *resp.GroupId
		return nil
	})

	p.addStep("tagging ELB Security Group", func() error {
		return tagAWSResource(ec2S, m.Config.ELBSecurityGroupID, map[string]string{
			"KubernetesCluster": m.Name,
		})
	})

	p.addStep("creating ELB Security Group ingress rules", func() error {
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(m.Config.ELBSecurityGroupID),
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort:   aws.Int64(0),
					ToPort:     aws.Int64(0),
					IpProtocol: aws.String("-1"),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},
			},
		}
		if _, err := ec2S.AuthorizeSecurityGroupIngress(input); err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
			return err
		}
		return nil
	})

	p.addStep("creating ELB Security Group egress rules", func() error {
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(m.Config.ELBSecurityGroupID),
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort:   aws.Int64(30000),
					ToPort:     aws.Int64(40000),
					IpProtocol: aws.String("tcp"),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},
				{
					FromPort:   aws.Int64(10250),
					ToPort:     aws.Int64(10250),
					IpProtocol: aws.String("tcp"),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},
			},
		}
		if _, err := ec2S.AuthorizeSecurityGroupIngress(input); err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
			return err
		}
		return nil
	})

	p.addStep("creating Node Security Group", func() error {
		if m.Config.NodeSecurityGroupID != "" {
			return nil
		}
		input := &ec2.CreateSecurityGroupInput{
			GroupName:   aws.String(m.Name + "_sg"),
			Description: aws.String("Allow any traffic to 443 and 22, but only traffic from ELB for 10250 and 30k-40k"),
			VpcId:       aws.String(m.Config.VPCID),
		}
		resp, err := ec2S.CreateSecurityGroup(input)
		if err != nil {
			return err
		}
		m.Config.NodeSecurityGroupID = *resp.GroupId
		return nil
	})

	p.addStep("tagging Node Security Group", func() error {
		return tagAWSResource(ec2S, m.Config.NodeSecurityGroupID, map[string]string{
			"KubernetesCluster": m.Name,
		})
	})

	p.addStep("creating Node Security Group ingress rules", func() error {
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(m.Config.NodeSecurityGroupID),
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort:   aws.Int64(0),
					ToPort:     aws.Int64(0),
					IpProtocol: aws.String("-1"),
					UserIdGroupPairs: []*ec2.UserIdGroupPair{
						{
							GroupId: aws.String(m.Config.NodeSecurityGroupID), // ?? TODO is this correct? -- https://github.com/supergiant/terraform-assets/blob/master/aws/1.1.7/security_groups.tf#L39
						},
					},
				},
				{
					FromPort:   aws.Int64(22),
					ToPort:     aws.Int64(22),
					IpProtocol: aws.String("tcp"),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},
				{
					FromPort:   aws.Int64(443),
					ToPort:     aws.Int64(443),
					IpProtocol: aws.String("tcp"),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},
				{
					FromPort:   aws.Int64(30000),
					ToPort:     aws.Int64(40000),
					IpProtocol: aws.String("tcp"),
					UserIdGroupPairs: []*ec2.UserIdGroupPair{
						{
							GroupId: aws.String(m.Config.ELBSecurityGroupID),
						},
					},
				},
				{
					FromPort:   aws.Int64(10250),
					ToPort:     aws.Int64(10250),
					IpProtocol: aws.String("tcp"),
					UserIdGroupPairs: []*ec2.UserIdGroupPair{
						{
							GroupId: aws.String(m.Config.ELBSecurityGroupID),
						},
					},
				},
			},
		}
		if _, err := ec2S.AuthorizeSecurityGroupIngress(input); err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
			return err
		}
		return nil
	})

	p.addStep("creating Node Security Group egress rules", func() error {
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(m.Config.NodeSecurityGroupID),
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort:   aws.Int64(0),
					ToPort:     aws.Int64(0),
					IpProtocol: aws.String("-1"),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},
			},
		}
		if _, err := ec2S.AuthorizeSecurityGroupIngress(input); err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
			return err
		}
		return nil
	})

	// Master Instance

	p.addStep("creating Server for Kubernetes master", func() error {
		if m.Config.MasterID != "" {
			return nil
		}

		userdataTemplate, err := ioutil.ReadFile("config/master_userdata.txt")
		if err != nil {
			return err
		}
		template, err := template.New("master_template").Parse(string(userdataTemplate))
		if err != nil {
			return err
		}
		var userdata bytes.Buffer
		if err = template.Execute(&userdata, m); err != nil {
			return err
		}
		encodedUserdata := base64.StdEncoding.EncodeToString(userdata.Bytes())

		input := &ec2.RunInstancesInput{
			MinCount:     aws.Int64(1),
			MaxCount:     aws.Int64(1),
			ImageId:      aws.String(AWSMasterAMIs[m.Config.Region]),
			InstanceType: aws.String(m.Config.MasterInstanceType), // NOTE this **should** be the smallest
			KeyName:      aws.String(m.Name + "-key"),
			NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
				{
					DeviceIndex:              aws.Int64(0),
					AssociatePublicIpAddress: aws.Bool(true),
					DeleteOnTermination:      aws.Bool(true),
					Groups: []*string{
						aws.String(m.Config.NodeSecurityGroupID),
					},
					SubnetId:         aws.String(m.Config.PublicSubnetID),
					PrivateIpAddress: aws.String(m.Config.MasterPrivateIP),
				},
			},
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Name: aws.String("kubernetes-master"),
			},
			BlockDeviceMappings: []*ec2.BlockDeviceMapping{
				{
					DeviceName: aws.String("/dev/xvdb"),
					Ebs: &ec2.EbsBlockDevice{
						DeleteOnTermination: aws.Bool(true),
						VolumeType:          aws.String("gp2"),
						VolumeSize:          aws.Int64(20),
					},
				},
			},
			UserData: aws.String(encodedUserdata),
		}
		resp, err := ec2S.RunInstances(input)
		if err != nil {
			return err
		}

		instance := resp.Instances[0]

		m.Config.MasterID = *instance.InstanceId
		return nil
	})

	p.addStep("tagging Kubernetes master", func() error {
		return tagAWSResource(ec2S, m.Config.MasterID, map[string]string{
			"KubernetesCluster": m.Name,
			"Name":              m.Name + "-master",
			"Role":              m.Name + "-master",
		})
	})

	// Wait for server to be ready

	p.addStep("waiting for Kubernetes master to launch", func() error {
		input := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(m.Config.MasterID),
			},
		}

		return action.cancellableWaitFor("Kubernetes master launch", 5*time.Minute, 3*time.Second, func() (bool, error) {
			resp, err := ec2S.DescribeInstances(input)
			if err != nil {
				return false, err
			}

			instance := resp.Reservations[0].Instances[0]

			// Save IP when ready
			if m.Config.MasterPublicIP == "" {
				if ip := instance.PublicIpAddress; ip != nil {
					m.Config.MasterPublicIP = *ip
					if err := c.core.DB.Save(m); err != nil {
						return false, err
					}
				}
			}

			return *instance.State.Name == "running", nil
		})
	})

	// Create route for master

	p.addStep("creating Route for Kubernetes master", func() error {
		input := &ec2.CreateRouteInput{
			DestinationCidrBlock: aws.String("10.246.0.0/24"),
			RouteTableId:         aws.String(m.Config.RouteTableID),
			InstanceId:           aws.String(m.Config.MasterID),
		}
		_, err := ec2S.CreateRoute(input)
		return err
	})

	// Create autoscaling groups

	p.addStep("creating Kubernetes minion AutoScaling Groups", func() error {
		for i, instanceType := range m.Config.InstanceTypes {
			launchConfigName := m.AutoScalingGroupName(instanceType)

			userdataTemplate, err := ioutil.ReadFile("config/minion_userdata.txt")
			if err != nil {
				return err
			}
			template, err := template.New("minion_template").Parse(string(userdataTemplate))
			if err != nil {
				return err
			}
			var userdata bytes.Buffer
			if err = template.Execute(&userdata, m); err != nil {
				return err
			}
			encodedUserdata := base64.StdEncoding.EncodeToString(userdata.Bytes())

			lcInput := &autoscaling.CreateLaunchConfigurationInput{
				LaunchConfigurationName:  aws.String(launchConfigName),
				InstanceType:             aws.String(instanceType),
				ImageId:                  aws.String(AWSMasterAMIs[m.Config.Region]),
				AssociatePublicIpAddress: aws.Bool(true),
				EbsOptimized:             aws.Bool(true),
				KeyName:                  aws.String(m.Name + "-key"),
				SecurityGroups: []*string{
					aws.String(m.Config.NodeSecurityGroupID),
				},
				IamInstanceProfile: aws.String("kubernetes-minion"),
				BlockDeviceMappings: []*autoscaling.BlockDeviceMapping{

					// root device ? TODO (do we need the one after this?)
					{
						DeviceName: aws.String("/dev/sda1"),
						Ebs: &autoscaling.Ebs{
							VolumeType:          aws.String("gp2"),
							VolumeSize:          aws.Int64(80),
							DeleteOnTermination: aws.Bool(true),
						},
					},

					&autoscaling.BlockDeviceMapping{
						DeviceName: aws.String("/dev/xvdb"),
						Ebs: &autoscaling.Ebs{
							VolumeType:          aws.String("gp2"),
							VolumeSize:          aws.Int64(80),
							DeleteOnTermination: aws.Bool(true),
						},
					},
				},
				UserData: aws.String(encodedUserdata),
			}
			if _, err = autoscalingS.CreateLaunchConfiguration(lcInput); err != nil && !strings.Contains(err.Error(), "already exists") {
				return err
			}

			desiredCapacity := 0
			if i == 0 { // NOTE !! should (should) be the smallest
				desiredCapacity = 1
			}

			input := &autoscaling.CreateAutoScalingGroupInput{
				AutoScalingGroupName:    aws.String(launchConfigName),
				LaunchConfigurationName: aws.String(launchConfigName),
				VPCZoneIdentifier:       aws.String(m.Config.PublicSubnetID),
				HealthCheckGracePeriod:  aws.Int64(100),
				HealthCheckType:         aws.String("EC2"),
				MinSize:                 aws.Int64(0),
				MaxSize:                 aws.Int64(int64(desiredCapacity)),
				DesiredCapacity:         aws.Int64(int64(desiredCapacity)),
				Tags: []*autoscaling.Tag{
					{
						Key:               aws.String("KubernetesCluster"),
						Value:             aws.String(m.Name),
						PropagateAtLaunch: aws.Bool(true),
					},
					{
						Key:               aws.String("Name"),
						Value:             aws.String(m.Name + "-minion"),
						PropagateAtLaunch: aws.Bool(true),
					},
					{
						Key:               aws.String("Role"),
						Value:             aws.String(m.Name + "-minion"),
						PropagateAtLaunch: aws.Bool(true),
					},
				},
			}
			if _, err = autoscalingS.CreateAutoScalingGroup(input); err != nil && !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
		return nil
	})

	p.addStep("waiting for Kubernetes", func() error {
		return action.cancellableWaitFor("Kubernetes API and first minion", 20*time.Minute, time.Second, func() (bool, error) {
			nodes, err := c.core.K8S(m).Nodes().List()
			if err != nil {
				return false, nil
			}
			return len(nodes.Items) > 0, nil
		})
	})

	if err := p.run(); err != nil {
		return err
	}

	return c.core.DB.Model(m).Update("ready", true).Error
}

//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------

func (c *Kubes) teardown(m *models.Kube) error {
	ec2S := c.core.CloudAccounts.ec2(m.CloudAccount, m.Config.Region)
	autoscalingS := c.core.CloudAccounts.autoscaling(m.CloudAccount, m.Config.Region)
	p := &provisioner{core: c.core, kube: m}

	p.addStep("deleting minion Instances", func() error {
		for _, instanceType := range m.Config.InstanceTypes {
			launchConfigName := m.AutoScalingGroupName(instanceType)

			asInput := &autoscaling.DeleteAutoScalingGroupInput{
				AutoScalingGroupName: aws.String(launchConfigName),
				ForceDelete:          aws.Bool(true), // this should delete the instances
			}
			_, err := autoscalingS.DeleteAutoScalingGroup(asInput)
			if isErrAndNotAWSNotFound(err) {
				return err
			}

			lcInput := &autoscaling.DeleteLaunchConfigurationInput{
				LaunchConfigurationName: aws.String(launchConfigName),
			}
			_, err = autoscalingS.DeleteLaunchConfiguration(lcInput)
			if isErrAndNotAWSNotFound(err) {
				return err
			}
		}
		return nil
	})

	p.addStep("deleting master", func() error {
		if m.Config.MasterID == "" {
			return nil
		}

		input := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(m.Config.MasterID),
			},
		}
		if _, err := ec2S.TerminateInstances(input); isErrAndNotAWSNotFound(err) {
			return err
		}

		// Wait for termination
		descinput := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(m.Config.MasterID),
			},
		}
		waitErr := util.WaitFor("Kubernetes master termination", 5*time.Minute, 3*time.Second, func() (bool, error) { // TODO --------- use server() method
			resp, err := ec2S.DescribeInstances(descinput)
			if err != nil {
				return false, err
			}
			if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
				return true, nil
			}
			instance := resp.Reservations[0].Instances[0]
			return *instance.State.Name == "terminated", nil
		})
		// Done waiting
		if waitErr != nil {
			return waitErr
		}

		m.Config.MasterID = ""
		return nil
	})

	p.addStep("disassociating Route Table from Subnet", func() error {
		if m.Config.RouteTableSubnetAssociationID == "" {
			return nil
		}
		input := &ec2.DisassociateRouteTableInput{
			AssociationId: aws.String(m.Config.RouteTableSubnetAssociationID),
		}
		if _, err := ec2S.DisassociateRouteTable(input); isErrAndNotAWSNotFound(err) {
			return err
		}
		m.Config.RouteTableSubnetAssociationID = ""
		return nil
	})

	p.addStep("deleting Internet Gateway", func() error {
		if m.Config.InternetGatewayID == "" {
			return nil
		}
		diginput := &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(m.Config.InternetGatewayID),
			VpcId:             aws.String(m.Config.VPCID),
		}

		// NOTE we do this (maybe we should just describe, not spam detach) because
		// we can't wait directly on minions to terminate (we can, but I'm lazy rn)
		waitErr := util.WaitFor("Internet Gateway to detach", 5*time.Minute, 5*time.Second, func() (bool, error) {
			if _, err := ec2S.DetachInternetGateway(diginput); err != nil && !strings.Contains(err.Error(), "not attached") {

				p.core.Log.Warn(err.Error())

				return false, nil
			}
			return true, nil
		})
		if waitErr != nil {
			return waitErr
		}

		input := &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(m.Config.InternetGatewayID),
		}
		if _, err := ec2S.DeleteInternetGateway(input); isErrAndNotAWSNotFound(err) {
			return err
		}
		m.Config.InternetGatewayID = ""
		return nil
	})

	p.addStep("deleting Route Table", func() error {
		if m.Config.RouteTableID == "" {
			return nil
		}
		input := &ec2.DeleteRouteTableInput{
			RouteTableId: aws.String(m.Config.RouteTableID),
		}
		if _, err := ec2S.DeleteRouteTable(input); isErrAndNotAWSNotFound(err) {
			return err
		}
		m.Config.RouteTableID = ""
		return nil
	})

	p.addStep("deleting public Subnet", func() error {
		if m.Config.PublicSubnetID == "" {
			return nil
		}
		input := &ec2.DeleteSubnetInput{
			SubnetId: aws.String(m.Config.PublicSubnetID),
		}

		waitErr := util.WaitFor("Public Subnet to delete", 2*time.Minute, 5*time.Second, func() (bool, error) {
			if _, err := ec2S.DeleteSubnet(input); isErrAndNotAWSNotFound(err) {
				return false, nil
			}
			return true, nil
		})
		if waitErr != nil {
			return waitErr
		}

		m.Config.PublicSubnetID = ""
		return nil
	})

	p.addStep("deleting Node Security Group", func() error {
		if m.Config.NodeSecurityGroupID == "" {
			return nil
		}
		input := &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(m.Config.NodeSecurityGroupID),
		}
		if _, err := ec2S.DeleteSecurityGroup(input); isErrAndNotAWSNotFound(err) {
			return err
		}
		m.Config.NodeSecurityGroupID = ""
		return nil
	})

	p.addStep("deleting ELB Security Group", func() error {
		if m.Config.ELBSecurityGroupID == "" {
			return nil
		}
		input := &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(m.Config.ELBSecurityGroupID),
		}
		if _, err := ec2S.DeleteSecurityGroup(input); isErrAndNotAWSNotFound(err) {
			return err
		}
		m.Config.ELBSecurityGroupID = ""
		return nil
	})

	p.addStep("deleting VPC", func() error {
		if m.Config.VPCID == "" {
			return nil
		}
		input := &ec2.DeleteVpcInput{
			VpcId: aws.String(m.Config.VPCID),
		}
		if _, err := ec2S.DeleteVpc(input); isErrAndNotAWSNotFound(err) {
			return err
		}
		m.Config.VPCID = ""
		return nil
	})

	p.addStep("deleting SSH Key Pair", func() error {
		input := &ec2.DeleteKeyPairInput{
			KeyName: aws.String(m.Name + "-key"),
		}
		if _, err := ec2S.DeleteKeyPair(input); isErrAndNotAWSNotFound(err) {
			return err
		}
		return nil
	})

	return p.run()
}
