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
	"errors"

	"github.com/eraclitux/ecsundo/internal/platform/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func makeServiceRunE(ecs ecsProvider) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clusterName := viper.GetString("cluster")
		if clusterName == "" {
			return errors.New("cluster cannot be empty")
		}
		if len(args) <= 0 {
			return errors.New("service name is mandatory")
		}
		serviceName := args[0]
		desiredVersion, err := ecs.ServicePreviousVersion(serviceName, clusterName)
		if err != nil {
			return err
		}
		return ecs.ServiceRollback(serviceName, clusterName, desiredVersion)
	}
}

// serviceCmd represents the service command.
var serviceCmd = &cobra.Command{
	Use:   "service [flags] <service-name>",
	Short: "Rollback an ECS service by name",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This hook helps to inject runtime parameters to the ecsProvider.
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		cmd.RunE = makeServiceRunE(aws.NewECSClient(verbose))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Overridden by PersistentPreRun.
		return nil
	},
}

func init() {
	serviceCmd.PersistentFlags().StringP("cluster", "c", "", "The ECS cluster name when the service run")
	viper.BindPFlag("cluster", serviceCmd.PersistentFlags().Lookup("cluster"))
	rootCmd.AddCommand(serviceCmd)
}
