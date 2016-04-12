package core

import (
	"errors"
	"io/ioutil"
	"net/http"
)

func checkForAWSMeta() {
	// Discover AWS ENV
	if AwsRegion == "" {
		region, err := getAWSRegion()
		if err != nil {
			Log.Error(err)
		} else {
			AwsRegion = region
			Log.Info("AWS Region Detected,", AwsRegion)
		}
	}
	if AwsAZ == "" {
		az, err := getAWSAZ()
		if err != nil {
			Log.Error(err)
		} else {
			AwsAZ = az
			Log.Info("AWS AZ Detected,", AwsAZ)
		}
	}
	if AwsSgID == "" {
		sg, err := getAWSSecurityGroupID()
		if err != nil {
			Log.Error(err)
		} else {
			AwsSgID = sg
			Log.Info("AWS Security Group Detected,", AwsSgID)
		}
	}
	if AwsSubnetID == "" {
		sub, err := getAWSSubnetID()
		if err != nil {
			Log.Info("ERROR:", err)
		} else {
			AwsSubnetID = sub
			Log.Info("INFO: AWS Security Group Detected,", AwsSubnetID)
		}
	}
}

// This file contains simple functions to discover metadata from the cloud provider.
func getMacs() (string, error) {
	macr, err := http.Get("http://169.254.169.254/latest/meta-data/network/interfaces/macs/")
	if err != nil {
		return "", err
	}
	defer macr.Body.Close()

	maclen, err := macr.Body.Read(nil)
	if err != nil {
		return "", err
	}
	if maclen > 1 {
		return "", errors.New("More than one mac address detected. Cannot determine Securtiy Group. Please specify manually.")
	}

	macbyte, err := ioutil.ReadAll(macr.Body)
	if err != nil {
		return "", err
	}
	mac := string(macbyte)
	return mac, nil
}
func getAWSRegion() (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/placement/availability-zone")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	az := string(body)
	region := az[:len(az)-1]

	return region, nil
}

func getAWSAZ() (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/placement/availability-zone")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	az := string(body)
	return az, nil
}

func getAWSSecurityGroupID() (string, error) {
	mac, err := getMacs()
	if err != nil {
		return "", err
	}
	secgr, err := http.Get("http://169.254.169.254/latest/meta-data/network/interfaces/macs/" + mac + "/security-group-ids")
	if err != nil {
		return "", err
	}
	defer secgr.Body.Close()

	sglen, err := secgr.Body.Read(nil)
	if err != nil {
		return "", err
	}
	if sglen > 1 {
		return "", errors.New("More than one security group detected. Cannot determine security group. Please specify manually.")
	}

	sgidbyte, err := ioutil.ReadAll(secgr.Body)
	if err != nil {
		return "", err
	}

	sgid := string(sgidbyte)
	return sgid, nil
}

func getAWSSubnetID() (string, error) {
	mac, err := getMacs()
	if err != nil {
		return "", err
	}
	secgr, err := http.Get("http://169.254.169.254/latest/meta-data/network/interfaces/macs/" + mac + "/subnet-id")
	if err != nil {
		return "", err
	}
	defer secgr.Body.Close()

	sgidbyte, err := ioutil.ReadAll(secgr.Body)
	if err != nil {
		return "", err
	}

	subid := string(sgidbyte)
	return subid, nil
}
