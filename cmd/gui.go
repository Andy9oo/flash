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
	"flash/pkg/gui"

	"github.com/spf13/cobra"
)

// guiCmd represents the gui command
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Opens a graphical search box",
	Run: func(cmd *cobra.Command, args []string) {
		gui.Show()
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
