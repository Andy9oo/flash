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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists documents added to the index",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.DialHTTP("tcp", "localhost:1234")
		if err != nil {
			log.Fatal(err)
		}

		var results monitordaemon.DirList
		err = client.Call("Handler.List", "", &results)
		if err != nil {
			log.Fatal(err)
		}

		for i := range results.Dirs {
			fmt.Printf("%d: %v\n", i+1, results.Dirs[i])
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
