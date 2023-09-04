// Copyright © 2018 Andrea Masi <eraclitux@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

const awsApisErrorFmt = "error on AWS request: %s"

// ServiceInfo stores state for a service.
type ServiceInfo struct {
	ARN     string
	TaskARN string
}

// ECSService implements cli.ecsProvider.
type ECSService struct {
	verbose bool
	client  *ecs.Client
}

// ServicePreviousVersion returns previous task version as ARN string.
func (es *ECSService) ServicePreviousVersion(serviceName, clusterName string) (string, error) {
	taskARN, err := es.getCurrentTask(serviceName, clusterName)
	if err != nil {
		return "", fmt.Errorf(awsApisErrorFmt, err)
	}
	tt := strings.Split(taskARN, ":")
	currentVersionStr := tt[len(tt)-1]
	currentVersion, err := strconv.Atoi(currentVersionStr)
	if err != nil {
		return "", err
	}
	previousVersion := currentVersion - 1
	if previousVersion <= 0 {
		return "", fmt.Errorf("impossible to rollback to version %d", previousVersion)
	}
	previousVersionStr := strconv.Itoa(previousVersion)
	return strings.Join(append(tt[:len(tt)-1], previousVersionStr), ":"), nil
}

// ServiceRollback updates a service to use a specific task version. If task is
// INACTIVE a new one is created with the old configuration.
func (es *ECSService) ServiceRollback(serviceName, clusterName, taskARN string) error {
	updateInput := &ecs.UpdateServiceInput{
		Cluster:        aws.String(clusterName),
		Service:        aws.String(serviceName),
		TaskDefinition: aws.String(taskARN),
	}
	_, err := es.client.UpdateService(context.TODO(), updateInput)
	if err != nil {
		if err.Error() != "TaskDefinition is inactive" {
			// return e
		}
	}
	// default:
	// 	return fmt.Errorf(awsApisErrorFmt, err)
	// }
	// At this point task is INACTIVE, register a new one with the same
	// configuration and update service with this.
	describeInput := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskARN),
	}
	out, err := es.client.DescribeTaskDefinition(context.TODO(), describeInput)
	if err != nil {
		return fmt.Errorf(awsApisErrorFmt, err)
	}
	taskDef := out.TaskDefinition
	registerInput := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    taskDef.ContainerDefinitions,
		Cpu:                     taskDef.Cpu,
		ExecutionRoleArn:        taskDef.ExecutionRoleArn,
		Family:                  taskDef.Family,
		Memory:                  taskDef.Memory,
		NetworkMode:             taskDef.NetworkMode,
		PlacementConstraints:    taskDef.PlacementConstraints,
		RequiresCompatibilities: taskDef.RequiresCompatibilities,
		TaskRoleArn:             taskDef.TaskRoleArn,
		Volumes:                 taskDef.Volumes,
	}
	registerOut, err := es.client.RegisterTaskDefinition(context.TODO(), registerInput)
	if err != nil {
		return fmt.Errorf(awsApisErrorFmt, err)
	}
	if es.verbose {
		fmt.Printf("%q new task definition registered with configuration from %q\n", *registerOut.TaskDefinition.TaskDefinitionArn, taskARN)
	}
	updateInput.TaskDefinition = registerOut.TaskDefinition.TaskDefinitionArn
	_, err = es.client.UpdateService(context.TODO(), updateInput)
	if err != nil {
		return fmt.Errorf(awsApisErrorFmt, err)
	}
	return nil
}

// ClusterRollback updates all services in an ECS cluster to their own previous
// task definition.
func (es *ECSService) ClusterRollback(clusterName string) error {
	serviceARNptrs, err := es.listServices(clusterName)
	if err != nil {
		return fmt.Errorf(awsApisErrorFmt, err)
	}
	servicesInfo := make([]ServiceInfo, 0, len(serviceARNptrs))
	for _, serviceARNptr := range serviceARNptrs {
		servicesInfo = append(servicesInfo, ServiceInfo{ARN: serviceARNptr, TaskARN: ""})
	}
	return es.rollbackServices(servicesInfo, clusterName)
}

// ClusterSnapshot returns current task versions for all services.
func (es *ECSService) ClusterSnapshot(clusterName string) ([]ServiceInfo, error) {
	serviceARNptrs, err := es.listServices(clusterName)
	if err != nil {
		return nil, fmt.Errorf(awsApisErrorFmt, err)
	}
	errCh := make(chan error)
	resultsCh := make(chan ServiceInfo)
	for _, serviceARNptr := range serviceARNptrs {
		go func(serviceARN string) {
			client := NewECSClient(es.verbose)
			taskARN, err := client.getCurrentTask(serviceARN, clusterName)
			if err != nil {
				errCh <- fmt.Errorf("%q: %s", nameFromARN(serviceARN), err)
				return
			}
			resultsCh <- ServiceInfo{ARN: serviceARN, TaskARN: taskARN}
		}(serviceARNptr)
	}
	l := len(serviceARNptrs)
	servicesInfo := make([]ServiceInfo, 0, l)
	var lastErr error
	for i := 0; i < l; i++ {
		select {
		case result := <-resultsCh:
			servicesInfo = append(servicesInfo, result)
		case lastErr = <-errCh:
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf(awsApisErrorFmt, lastErr)
	}
	return servicesInfo, nil
}

// ClusterRestore restores all services to specific versions.
func (es *ECSService) ClusterRestore(serviceSnapshots []ServiceInfo, clusterName string) error {
	return es.rollbackServices(serviceSnapshots, clusterName)
}

func (es *ECSService) listServices(clusterName string) ([]string, error) {
	listInput := &ecs.ListServicesInput{
		Cluster: aws.String(clusterName),
	}
	listOut, err := es.client.ListServices(context.TODO(), listInput)
	if err != nil {
		return nil, err
	}
	serviceARNs := listOut.ServiceArns
	for listOut.NextToken != nil {
		listInput.NextToken = listOut.NextToken
		listOut, err = es.client.ListServices(context.TODO(), listInput)
		if err != nil {
			return nil, err
		}
		serviceARNs = append(serviceARNs, listOut.ServiceArns...)
	}
	return serviceARNs, nil
}

func (es *ECSService) getCurrentTask(serviceName, clusterName string) (string, error) {
	input := &ecs.DescribeServicesInput{
		Cluster: aws.String(clusterName),
		Services: []string{
			serviceName,
		},
	}
	result, err := es.client.DescribeServices(context.TODO(), input)
	if err != nil {
		return "", err
	}
	var taskARNPtr *string
	if len(result.Services) > 0 {
		taskARNPtr = result.Services[0].TaskDefinition
	}
	if taskARNPtr == nil {
		return "", fmt.Errorf("empty task definition for %s", serviceName)
	}
	return *taskARNPtr, nil
}

// rollbackServices rollbacks all services to the versions specified.
// If the task version supplied is empty it will attempt to rollback to previous version.
func (es *ECSService) rollbackServices(servicesInfo []ServiceInfo, clusterName string) error {
	doneCh := make(chan struct{})
	failedCh := make(chan string)
	for _, service := range servicesInfo {
		go func(service ServiceInfo) {
			defer func() {
				doneCh <- struct{}{}
			}()
			sendError := func(err error) {
				failedCh <- fmt.Sprintf("%q: %s", nameFromARN(service.ARN), err)
			}
			client := NewECSClient(es.verbose)
			if service.TaskARN == "" {
				taskARN, err := client.ServicePreviousVersion(service.ARN, clusterName)
				if err != nil {
					sendError(err)
					return
				}
				service.TaskARN = taskARN
			}
			if es.verbose {
				fmt.Printf("rolling back %q to %s\n", nameFromARN(service.ARN), nameFromARN(service.TaskARN))
			}
			err := client.ServiceRollback(service.ARN, clusterName, service.TaskARN)
			if err != nil {
				sendError(err)
				return
			}
		}(service)
	}
	l := len(servicesInfo)
	failedServiceARNs := make([]string, 0, l)
	i := 0
	for i < l {
		select {
		case <-doneCh:
			i++
		case serviceARN := <-failedCh:
			failedServiceARNs = append(failedServiceARNs, serviceARN)
		}
	}
	if len(failedServiceARNs) > 0 {
		return fmt.Errorf(
			"rollback failed on these services:\n%s",
			strings.Join(failedServiceARNs, "\n"),
		)
	}
	return nil
}

// NewECSClient returns an implementation of cmd.ecsService.
func NewECSClient(verbose bool) *ECSService {
	if os.Getenv("AWS_REGION") == "" {
		region, err := getRegion()
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to get AWS region from metadata")
		} else {
			os.Setenv("AWS_REGION", region)
		}
	}
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err.Error())
	}
	// create a new context from the previous ctx with a timeout, e.g. 5 seconds

	return &ECSService{
		verbose: verbose,
		client:  ecs.NewFromConfig(cfg),
	}
}
