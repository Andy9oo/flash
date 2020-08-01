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
	"flash/pkg/monitordaemon"
	"fmt"
	"log"
	"net/rpc"

	"github.com/spf13/cobra"
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find \"<query>\"",
	Short: "Search the index for a query",
	Run: func(cmd *cobra.Command, args []string) {
		n, _ := cmd.Flags().GetInt("num_results")
		query := args[0]

		client, err := rpc.DialHTTP("tcp", "localhost:1234")
		if err != nil {
			log.Fatal("Connection error: ", err)
		}

		var results monitordaemon.Results
		err = client.Call("Handler.Search", monitordaemon.Query{Str: query, N: n}, &results)
		if err != nil {
			log.Fatal(err)
		}

		for i, path := range results.Paths {
			fmt.Printf("%d: %v\n", i+1, path)
		}
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	findCmd.Flags().IntP("num_results", "n", 10, "The number of results that will be returned")
	rootCmd.AddCommand(findCmd)
}
