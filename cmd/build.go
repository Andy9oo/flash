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
	"errors"
	"flash/pkg/index"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build <directory_path>",
	Short: "Builds the index for a given directory",
	Run: func(cmd *cobra.Command, args []string) {
		indexpath := viper.GetString("indexpath")

		// Delete current index
		os.RemoveAll(indexpath)

		// Build new index
		path, _ := filepath.Abs(args[0])
		index.Build(viper.GetString("indexpath"), path)
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("The directory to be indexed must be given")
		}

		path, _ := filepath.Abs(args[0])

		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return errors.New("Path given is not a directory")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
