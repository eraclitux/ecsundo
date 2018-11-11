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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/eraclitux/ecsundo/internal/platform/aws"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	separatorConst = ";"
	filePathFormat = "%s/.%s.ecsundo"
)

func makeClusterRunE(ecs ecsProvider) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clusterName := viper.GetString("cluster")
		if clusterName == "" {
			if len(args) <= 0 {
				return errors.New("cluster name is mandatory")
			}
		}
		if len(args) > 0 {
			clusterName = args[0]
		}
		err := ecs.ClusterRollback(clusterName)
		if err != nil {
			return fmt.Errorf("error for %q: %s", clusterName, err)
		}
		return nil
	}
}

func makeSnapshotRunE(ecs ecsProvider) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clusterName := viper.GetString("cluster")
		if clusterName == "" {
			if len(args) <= 0 {
				return errors.New("cluster name is mandatory")
			}
		}
		if len(args) > 0 {
			clusterName = args[0]
		}
		filePath, err := cmd.Flags().GetString("snapshot-path")
		if err != nil {
			return err
		}
		if filePath == "" {
			home, err := homedir.Dir()
			if err != nil {
				return err
			}
			filePath = fmt.Sprintf(filePathFormat, home, clusterName)
		}
		serviceVersions, err := ecs.ClusterSnapshot(clusterName)
		if err != nil {
			return fmt.Errorf("error for %q: %s", clusterName, err)
		}
		var snapshotData bytes.Buffer
		for _, serviceInfo := range serviceVersions {
			fmt.Fprintf(&snapshotData, "%s%s%s\n", serviceInfo.ARN, separatorConst, serviceInfo.TaskARN)
		}
		return ioutil.WriteFile(filePath, snapshotData.Bytes(), 0640)
	}
}

func makeRestoreRunE(ecs ecsProvider) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clusterName := viper.GetString("cluster")
		if clusterName == "" {
			if len(args) <= 0 {
				return errors.New("cluster name is mandatory")
			}
		}
		if len(args) > 0 {
			clusterName = args[0]
		}
		filePath, err := cmd.Flags().GetString("snapshot-path")
		if err != nil {
			return err
		}
		if filePath == "" {
			home, err := homedir.Dir()
			if err != nil {
				return err
			}
			filePath = fmt.Sprintf(filePathFormat, home, clusterName)
		}
		bb, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer(bb)
		servicesInfo := make([]aws.ServiceInfo, 0)
		for buf.Len() > 0 {
			line, err := buf.ReadString(byte('\n'))
			if err != nil {
				return err
			}
			ss := strings.Split(line, separatorConst)
			if len(ss) < 2 {
				return errors.New("invalid snapshot format")
			}
			servicesInfo = append(
				servicesInfo,
				aws.ServiceInfo{
					ARN:     strings.TrimRight(ss[0], "\n"),
					TaskARN: strings.TrimRight(ss[1], "\n"),
				},
			)
		}
		err = ecs.ClusterRestore(servicesInfo, clusterName)
		if err != nil {
			return fmt.Errorf("error for %q: %s", clusterName, err)
		}
		return nil
	}
}

// clusterCmd represents the cluster command.
var clusterCmd = &cobra.Command{
	Use:   "cluster [flags] <cluster-name>",
	Short: "Rollback all services in a given cluster",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This hook helps to inject runtime parameters to ecsProvider.
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		cmd.RunE = makeClusterRunE(aws.NewECSClient(verbose))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Overridden by PersistentPreRun.
		return nil
	},
}

// snapshotCmd represents the snapshot subcommand.
var snapshotCmd = &cobra.Command{
	Use:   "snapshot [flags] <cluster-name>",
	Short: "Save current task versions for all services",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This hook helps to inject runtime parameters to ecsProvider.
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		cmd.RunE = makeSnapshotRunE(aws.NewECSClient(verbose))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Overridden by PersistentPreRun.
		return nil
	},
}

// snapshotCmd represents the snapshot subcommand.
var restoreCmd = &cobra.Command{
	Use:   "restore [flags] <cluster-name>",
	Short: "Restore all services to the versions from the snapshot",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This hook helps to inject runtime parameters to ecsProvider.
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		cmd.RunE = makeRestoreRunE(aws.NewECSClient(verbose))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Overridden by PersistentPreRun.
		return nil
	},
}

func init() {
	snapshotCmd.Flags().StringP("snapshot-path", "s", "", "Path to snapshot file (default $HOME/.<cluster-name>.ecsundo)")
	restoreCmd.Flags().StringP("snapshot-path", "s", "", "Path to snapshot file (default $HOME/.<cluster-name>.ecsundo)")
	clusterCmd.AddCommand(snapshotCmd)
	clusterCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(clusterCmd)
}
