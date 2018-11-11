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

package cli

import "github.com/eraclitux/ecsundo/internal/platform/aws"

// ecsProvider models an interface on AWS ECS service apis.
type ecsProvider interface {
	// ServicePreviousVersion returns previous task version as ARN string.
	ServicePreviousVersion(serviceName, clusterName string) (string, error)
	// ServiceRollback updates a service to use a specific task version.
	ServiceRollback(serviceName, clusterName, taskARN string) error
	// ClusterRollback updates all services in a given cluster.
	ClusterRollback(clusterName string) error
	// ClusterSnapshot returns current task versions for all services.
	ClusterSnapshot(clusterName string) ([]aws.ServiceInfo, error)
	// ClusterRestore restores all services to specific versions.
	ClusterRestore(serviceSnapshots []aws.ServiceInfo, clusterName string) error
}
