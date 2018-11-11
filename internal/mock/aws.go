package mock

import "github.com/eraclitux/ecsundo/internal/platform/aws"

type ECSService struct {
	ServiceName string
	ClusterName string
	Version     string
}

func (ecs *ECSService) ServicePreviousVersion(serviceName, clusterName string) (string, error) {
	ecs.ServiceName = serviceName
	ecs.ClusterName = clusterName
	return "", nil
}

func (ecs *ECSService) ServiceRollback(serviceName, clusterName, version string) error {
	ecs.Version = version
	return nil
}

func (ecs *ECSService) ClusterRollback(clusterName string) error {
	ecs.ClusterName = clusterName
	return nil
}

func (ecs *ECSService) ClusterSnapshot(clusterName string) ([]aws.ServiceInfo, error) {
	return nil, nil
}

func (ecs *ECSService) ClusterRestore(serviceSnapshots []aws.ServiceInfo, clusterName string) error {
	return nil
}
