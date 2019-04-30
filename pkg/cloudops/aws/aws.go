package aws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/libopenstorage/openstorage/pkg/cloudops"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/sirupsen/logrus"
)

// ErrAWSEnvNotAvailable is the error type when aws credentials are not set
var ErrAWSEnvNotAvailable = fmt.Errorf("AWS credentials are not set in environment")

type awsOps struct {
	instInfo    *cloudops.InstanceInfo
	ec2         *ec2.EC2
	autoscaling *autoscaling.AutoScaling
}

// NewAWSps returns an instance of the AWS cloud ops implementation
func NewAWSOps() (cloudops.Ops, error) {
	ops := &awsOps{
		instInfo: &cloudops.InstanceInfo{},
	}
	err := ops.populateBasicInfo()
	if err != nil {
		return nil, err
	}

	ops.instInfo.Region = ops.instInfo.Zone[:len(ops.instInfo.Zone)-1]

	ops.ec2 = ec2.New(
		session.New(
			&aws.Config{
				Region:      &ops.instInfo.Region,
				Credentials: credentials.NewEnvCredentials(),
			},
		),
	)

	ops.autoscaling = autoscaling.New(
		session.New(
			&aws.Config{
				Region:      &ops.instInfo.Region,
				Credentials: credentials.NewEnvCredentials(),
			},
		),
	)

	return ops, nil
}

func (a *awsOps) InspectSelf() (*cloudops.InstanceInfo, error) {
	inst, err := DescribeInstanceByID(a.ec2, a.instInfo.ID)
	if err != nil {
		return nil, err
	}

	a.instInfo.Labels = labelsFromTags(inst.Tags)
	return a.instInfo, nil
}

func (a *awsOps) InspectSelfInstanceGroup() (*cloudops.InstanceGroupInfo, error) {
	selfInfo, err := a.InspectSelf()
	if err != nil {
		return nil, err
	}

	for tag, value := range selfInfo.Labels {
		// https://docs.aws.amazon.com/autoscaling/ec2/userguide/autoscaling-tagging.html#tag-lifecycle
		if tag == "aws:autoscaling:groupName" {
			input := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{
					aws.String(value),
				},
			}

			result, err := a.autoscaling.DescribeAutoScalingGroups(input)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				} else {
					return nil, err
				}
			}

			if len(result.AutoScalingGroups) != 1 {
				return nil, fmt.Errorf("DescribeAutoScalingGroups (%v) returned %v groups, expect 1",
					value, len(result.AutoScalingGroups))
			}

			group := result.AutoScalingGroups[0]
			if group.MinSize == nil || group.MaxSize == nil {
				return nil, fmt.Errorf("Autoscaling group: %s does not have min or max set", value)
			}

			zones := make([]string, 0)
			for _, z := range group.AvailabilityZones {
				zones = append(zones, *z)
			}

			retval := &cloudops.InstanceGroupInfo{
				CloudObjectMeta: cloudops.CloudObjectMeta{
					Name:   *group.AutoScalingGroupName,
					Zone:   a.instInfo.Zone,
					Region: a.instInfo.Region,
					Labels: labelsFromTags(group.Tags),
				},
				Zones:              zones,
				AutoscalingEnabled: true,
				Min:                *group.MinSize,
				Max:                *group.MaxSize,
			}

			return retval, nil
		}
	}

	return nil, fmt.Errorf("instance doesn't belong to an instance group")
}

func (a *awsOps) populateBasicInfo() error {
	var err error
	a.instInfo.Zone, a.instInfo.ID, err = getInfoFromMetadata()
	if err != nil {
		logrus.Infof("Failed to query ec2 metadata info. Instance may not be running on ec2.")
	} else {
		return nil
	}

	// try env variables
	a.instInfo.Zone, err = util.GetEnvValueStrict("AWS_ZONE")
	if err != nil {
		return err
	}

	a.instInfo.ID, err = util.GetEnvValueStrict("AWS_INSTANCE_ID")
	if err != nil {
		return err
	}

	if _, err := credentials.NewEnvCredentials().Get(); err != nil {
		return ErrAWSEnvNotAvailable
	}

	return nil

}

func getInfoFromMetadata() (string, string, error) {
	zone, err := metadata("placement/availability-zone")
	if err != nil {
		return "", "", err
	}

	instanceID, err := metadata("instance-id")
	if err != nil {
		return "", "", err
	}

	return zone, instanceID, nil
}

// instanceData retrieves instance data specified by key.
func instanceData(key string) (string, error) {
	client := http.Client{Timeout: time.Second * 3}
	url := "http://169.254.169.254/latest/" + key
	res, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err = fmt.Errorf("Code %d returned for url %s", res.StatusCode, url)
		return "", fmt.Errorf("Error querying AWS metadata for key %s: %v", key, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Error querying AWS metadata for key %s: %v", key, err)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("Failed to retrieve AWS metadata for key %s: %v", key, err)
	}
	return string(body), nil
}

func metadata(key string) (string, error) {
	return instanceData("meta-data/" + key)
}

// DescribeInstanceByID describes the given instance by instance ID
func DescribeInstanceByID(service *ec2.EC2, id string) (*ec2.Instance, error) {
	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&id},
	}
	out, err := service.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	if len(out.Reservations) != 1 {
		return nil, fmt.Errorf("DescribeInstances(%v) returned %v reservations, expect 1",
			id, len(out.Reservations))
	}
	if len(out.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("DescribeInstances(%v) returned %v Reservations, expect 1",
			id, len(out.Reservations[0].Instances))
	}
	return out.Reservations[0].Instances[0], nil
}

func labelsFromTags(input interface{}) map[string]string {
	labels := make(map[string]string)
	ec2Tags, ok := input.([]*ec2.Tag)
	if ok {
		for _, tag := range ec2Tags {
			if tag == nil {
				continue
			}

			if tag.Key == nil || tag.Value == nil {
				continue
			}

			labels[*tag.Key] = *tag.Value
		}

		return labels
	}

	autoscalingTags, ok := input.([]*autoscaling.TagDescription)
	if ok {
		for _, tag := range autoscalingTags {
			if tag == nil {
				continue
			}

			if tag.Key == nil || tag.Value == nil {
				continue
			}

			labels[*tag.Key] = *tag.Value
		}

		return labels
	}

	return labels
}
