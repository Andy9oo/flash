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
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a file or directory from the watch list",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Are you sure you want to remove %v [y/n]: ", args[0])
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			fmt.Println("Removing...")
			client, err := rpc.DialHTTP("tcp", "localhost:1234")
			if err != nil {
				log.Fatal(err)
			}

			var success bool
			err = client.Call("Handler.Remove", path, &success)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
