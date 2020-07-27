/*
Copyright © 2020 Andrew Cullis <acullis68@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"flash/pkg/monitordaemon"
	"fmt"

	"github.com/spf13/cobra"
)

// daemonCmd represents the init command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Used to control the file monitor daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := monitordaemon.Get()
		d.Watch()
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the file monitor daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := monitordaemon.Get()
		d.Start()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the file monitor daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := monitordaemon.Get()
		d.Stop()
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs the file monitor daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := monitordaemon.Get()
		_, err := d.Install()
		if err != nil {
			fmt.Println(err)
		}
		d.Start()
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes the file monitor daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := monitordaemon.Get()
		d.Stop()
		_, err := d.Remove()
		if err != nil {
			fmt.Println(err)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Returns the status of the file monitor daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := monitordaemon.Get()
		status, err := d.Status()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(status)
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run starts the file monitor daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := monitordaemon.Get()
		d.Watch()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(startCmd)
	daemonCmd.AddCommand(stopCmd)
	daemonCmd.AddCommand(installCmd)
	daemonCmd.AddCommand(removeCmd)
	daemonCmd.AddCommand(statusCmd)
	daemonCmd.AddCommand(runCmd)
}

