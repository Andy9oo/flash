/*
Copyright © 2020 Andrew Cullis acullis68@gmail.com

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
	"flash/pkg/index"
	"flash/pkg/monitordaemon"
	"fmt"
	"os"
	"os/user"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var fileIndex *index.Index
var daemon *monitordaemon.MonitorDaemon

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flash",
	Short: "flash is a full-text desktop search engine",
	Long:  "Flash is a full-text desktop search engine, designed to help users find their files. Using preprocessing techniques, flash creates an index, allowing searching without having to crawl the filesystem, substantially reducing search times.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, loadDatastructures)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flash.json)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	username := usr.Username
	if username == "root" {
		username = os.Getenv("SUDO_USER")
	}
	home := "/home/" + username

	viper.SetDefault("indexpath", home+"/.flash")
	viper.SetDefault("dirs", []string{})

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile(home + "/.flash.json")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func loadDatastructures() {
	fileIndex = index.Load(viper.GetString("indexpath"))
	daemon = monitordaemon.Init(viper.GetStringSlice("dirs"), fileIndex)
}
