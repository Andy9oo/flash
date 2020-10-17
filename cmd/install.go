/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"context"
	"flash/pkg/monitordaemon"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/go-tika/tika"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Performs all setup required for flash to run",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Installing")

		fmt.Println("Creating flash dir")
		os.MkdirAll(viper.GetString("flashhome"), 0775)

		fmt.Println("Downloading parser")
		tikapath := viper.GetString("tikapath")
		_, err := os.Stat(tikapath)
		if err != nil {
			err := tika.DownloadServer(context.Background(), "1.21", tikapath)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Println("Installing daemon")
		err = exec.Command("/bin/sh", "-c", "sudo flash daemon install").Run()
		if err != nil {
			log.Fatal(err)
		}
		d := monitordaemon.Init()
		d.Start()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
