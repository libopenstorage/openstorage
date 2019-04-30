package gcp

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/libopenstorage/openstorage/pkg/cloudops"
	"github.com/libopenstorage/openstorage/pkg/parser"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

const (
	clusterNameKey      = "cluster-name"
	clusterLocationKey  = "cluster-location"
	kubeLabelsKey       = "kube-labels"
	nodePoolKey         = "cloud.google.com/gke-nodepool"
	instanceTemplateKey = "instance-template"
)

type gcpOps struct {
	instInfo         *cloudops.InstanceInfo
	project          string
	computeService   *compute.Service
	containerService *container.Service
}

// NewGCPOps returns an instance of the GCP cloud ops implementation
func NewGCPOps() (cloudops.Ops, error) {
	ops := &gcpOps{
		instInfo: &cloudops.InstanceInfo{},
	}
	err := ops.populateBasicInfo()
	if err != nil {
		return nil, err
	}

	ops.instInfo.Region = ops.instInfo.Zone[:len(ops.instInfo.Zone)-2]

	ctx := context.Background()
	ops.computeService, err = compute.NewService(ctx, option.WithScopes(compute.ComputeScope))
	if err != nil {
		return nil, fmt.Errorf("unable to create Compute service: %v", err)
	}

	ops.containerService, err = container.NewService(ctx, option.WithScopes(compute.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("unable to create Container service: %v", err)
	}

	return ops, nil
}

func (g *gcpOps) InspectSelf() (*cloudops.InstanceInfo, error) {
	inst, err := g.computeService.Instances.Get(g.project, g.instInfo.Zone, g.instInfo.Name).Do()
	if err != nil {
		return nil, err
	}

	g.instInfo.Labels = inst.Labels
	return g.instInfo, nil
}

func (g *gcpOps) InspectSelfInstanceGroup() (*cloudops.InstanceGroupInfo, error) {
	inst, err := g.computeService.Instances.Get(g.project, g.instInfo.Zone, g.instInfo.Name).Do()
	if err != nil {
		return nil, err
	}

	meta := inst.Metadata
	if meta == nil {
		return nil, fmt.Errorf("instance doesn't have metadata set")
	}

	var (
		gkeClusterName   string
		instanceTemplate string
		clusterLocation  string
	)

	for _, item := range meta.Items {
		if item == nil {
			continue
		}

		if item.Key == clusterNameKey {
			if item.Value == nil {
				return nil, fmt.Errorf("instance has %s key in metadata but has invalid value", clusterNameKey)
			}

			gkeClusterName = *item.Value
		}

		if item.Key == instanceTemplateKey {
			if item.Value == nil {
				return nil, fmt.Errorf("instance has %s key in metadata but has invalid value", instanceTemplateKey)
			}

			instanceTemplate = *item.Value
		}

		if item.Key == clusterLocationKey {
			if item.Value == nil {
				return nil, fmt.Errorf("instance has %s key in metadata but has invalid value", clusterLocationKey)
			}

			clusterLocation = *item.Value
		}
	}

	if len(gkeClusterName) == 0 || len(instanceTemplate) == 0 || len(clusterLocation) == 0 {
		return nil, fmt.Errorf("API is currently only supported on the GKE platform.")
	}

	for _, item := range meta.Items {
		if item == nil {
			continue
		}

		if item.Key == kubeLabelsKey {
			if item.Value == nil {
				return nil, fmt.Errorf("instance has %s key in metadata but has invalid value", kubeLabelsKey)
			}

			kubeLabels, err := parser.LabelsFromString(*item.Value)
			if err != nil {
				return nil, err
			}

			for labelKey, labelValue := range kubeLabels {
				if labelKey == nodePoolKey {
					nodePoolPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", "portworx-eng", clusterLocation, gkeClusterName, labelValue)
					nodePool, err := g.containerService.Projects.Locations.Clusters.NodePools.Get(nodePoolPath).Do()
					//nodePool, err := g.containerService.Projects.Locations.Clusters.NodePools.Get(labelValue).Do()
					if err != nil {
						logrus.Errorf("failed to get node pool")
						return nil, err
					}

					zones := make([]string, 0)
					for _, igURL := range nodePool.InstanceGroupUrls {
						// e.g https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/instanceGroupManagers/gke-harsh-regional-asg-t-default-pool-a8750fe9-grp
						parts := strings.Split(igURL, "/")
						if len(parts) < 3 {
							return nil, fmt.Errorf("failed to parse zones for a node pool")
						}

						zones = append(zones, parts[len(parts)-3])
					}

					retval := &cloudops.InstanceGroupInfo{
						CloudObjectMeta: cloudops.CloudObjectMeta{
							Name:   nodePool.Name,
							Zone:   g.instInfo.Zone,
							Region: g.instInfo.Region,
						},
						Zones:              zones,
						AutoscalingEnabled: nodePool.Autoscaling.Enabled,
						Min:                nodePool.Autoscaling.MinNodeCount,
						Max:                nodePool.Autoscaling.MaxNodeCount,
					}

					if nodePool.Config != nil {
						retval.Labels = nodePool.Config.Labels
					}

					return retval, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("instance doesn't belong to a GKE node pool")
}

func (g *gcpOps) populateBasicInfo() error {
	var err error
	if metadata.OnGCE() {
		g.project, err = metadata.ProjectID()
		if err != nil {
			return err
		}

		g.instInfo.Zone, err = metadata.Zone()
		if err != nil {
			return err
		}

		g.instInfo.Name, err = metadata.InstanceName()
		if err != nil {
			return err
		}

		g.instInfo.ID, err = metadata.InstanceID()
		if err != nil {
			return err
		}

		return nil
	}

	// try env variables
	g.project, err = util.GetEnvValueStrict("GCE_INSTANCE_PROJECT")
	if err != nil {
		return err
	}

	g.instInfo.Zone, err = util.GetEnvValueStrict("GCE_INSTANCE_ZONE")
	if err != nil {
		return err
	}

	g.instInfo.Name, err = util.GetEnvValueStrict("GCE_INSTANCE_NAME")
	if err != nil {
		return err
	}

	// optional
	g.instInfo.ID, _ = util.GetEnvValueStrict("GCE_INSTANCE_ID")

	return nil
}
