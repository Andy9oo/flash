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
	"flash/pkg/search"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find \"<query>\"",
	Short: "Search the index for a query",
	RunE: func(cmd *cobra.Command, args []string) error {
		index := index.Load(viper.GetString("indexpath"))

		engine := search.NewEngine(index)

		n, _ := cmd.Flags().GetInt("num_results")
		start := time.Now()

		results := engine.Search(args[0], n)

		fmt.Printf("Found %d results in %v\n", len(results), time.Since(start))
		for i, result := range results {
			path, _, _ := index.GetDocInfo(result.ID)
			fmt.Printf("%v. %v (%v)\n", i+1, path, result.Score)
		}

		return nil
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	findCmd.Flags().IntP("num_results", "n", 10, "The number of results that will be returned")
	rootCmd.AddCommand(findCmd)
}
