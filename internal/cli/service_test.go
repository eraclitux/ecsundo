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

import (
	"testing"

	"github.com/eraclitux/ecsundo/internal/mock"
	"github.com/spf13/cobra"
)

func TestServiceRollback(t *testing.T) {
	clusterName := "my-cluster-under-test-a"
	serviceName := "my-service-under-test"
	ecsService := &mock.ECSService{}
	rootCmd.SetArgs([]string{"service", "-c", clusterName, serviceName})
	serviceCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.RunE = makeServiceRunE(ecsService)
		return nil
	}
	err := rootCmd.Execute()
	if err != nil {
		t.Log("running Execute():", err)
		t.FailNow()
	}
	if ecsService.ClusterName != clusterName {
		t.Fatal("wrong clusterName")
	}
	if ecsService.ServiceName != serviceName {
		t.Fatal("wrong serviceName")
	}
}
