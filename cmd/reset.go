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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Removes all files from the index",
	Run: func(cmd *cobra.Command, args []string) {
		indexpath := viper.GetString("indexpath")
		os.RemoveAll(indexpath)
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}