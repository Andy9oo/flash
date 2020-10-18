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
	"flash/pkg/monitordaemon"
	"fmt"
	"log"
	"net/rpc"

	"github.com/spf13/cobra"
)

// blacklistCmd represents the blacklist command
var blacklistCmd = &cobra.Command{
	Use:   "blacklist",
	Short: "Blacklists all files which match a given regex",
}

var blacklistAddCmd = &cobra.Command{
	Use:   "add \"<regex>\"",
	Short: "Blacklists all files which match a given regex",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.DialHTTP("tcp", "localhost:1234")
		if err != nil {
			log.Fatal(err)
		}

		var success bool
		err = client.Call("Handler.BlacklistAdd", args[0], &success)
		if err != nil {
			log.Fatal(err)
		}
	},
	Args: cobra.ExactArgs(1),
}

var blacklistRemoveCmd = &cobra.Command{
	Use:   "remove \"<regex>\"",
	Short: "Removes the given regex from the blacklist",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.DialHTTP("tcp", "localhost:1234")
		if err != nil {
			log.Fatal(err)
		}

		var success bool
		err = client.Call("Handler.BlacklistRemove", args[0], &success)
		if err != nil {
			log.Fatal(err)
		}
	},
	Args: cobra.ExactArgs(1),
}

var blacklistListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all blacklisted patterns",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.DialHTTP("tcp", "localhost:1234")
		if err != nil {
			log.Fatal(err)
		}

		var results monitordaemon.BlacklistPatterns
		err = client.Call("Handler.BlacklistGet", "", &results)
		if err != nil {
			log.Fatal(err)
		}

		for i := range results.Patterns {
			fmt.Printf("%d: %v\n", i+1, results.Patterns[i])
		}
	},
}

func init() {
	blacklistCmd.AddCommand(blacklistAddCmd)
	blacklistCmd.AddCommand(blacklistRemoveCmd)
	blacklistCmd.AddCommand(blacklistListCmd)
	rootCmd.AddCommand(blacklistCmd)
}
