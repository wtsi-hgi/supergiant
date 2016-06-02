package core

import (
	"io/ioutil"
	"net/http"
)

func checkForAWSMeta(c *Core) {
	if c.AwsAZ == "" {
		az, err := getAWSAZ()
		if err != nil {
			Log.Error(err)
		} else {
			c.AwsAZ = az
			c.AwsRegion = az[:len(az)-1]
			Log.Info("AWS Availability Zone detected", c.AwsAZ)
		}
	}
	if c.AwsSubnetID == "" {
		sub, err := getAWSSubnetID()
		if err != nil {
			Log.Error(err)
		} else {
			c.AwsSubnetID = sub
			Log.Info("AWS Security Group detected", c.AwsSubnetID)
		}
	}
}

// This file contains simple functions to discover metadata from the cloud provider.

func getAWSAZ() (string, error) {
	return getAwsMeta("placement/availability-zone")
}

func getAWSSubnetID() (string, error) {
	mac, err := getAwsMeta("network/interfaces/macs")
	if err != nil {
		return "", err
	}
	return getAwsMeta("network/interfaces/macs/" + mac + "/subnet-id")
}

func getAwsMeta(path string) (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/" + path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
