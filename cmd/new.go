// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new converter Project folder.",
	Long: `Create a new converter Project folder

	The created project has the following folder.
	 - input  Video file before conversion
	 - output Video file after conversion
	 - jobs   Conversion job
	 - preset Conversion format description file `,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		if name == "" {
			fmt.Println("error: ltc init [directory]")
			return
		}
		_, err := os.Stat(name)
		if err == nil {
			fmt.Println("already exists " + name + " directory")
			return
		}
		if err := os.Mkdir(name, 0777); err != nil {
			fmt.Println(err)
		}
		input := filepath.Join(name, "input")
		output := filepath.Join(name, "output")
		job := filepath.Join(name, "job")
		preset := filepath.Join(name, "preset")
		if err := os.Mkdir(input, 0777); err != nil {
			fmt.Println(err)
		}
		if err := os.Mkdir(output, 0777); err != nil {
			fmt.Println(err)
		}
		if err := os.Mkdir(job, 0777); err != nil {
			fmt.Println(err)
		}
		if err := os.Mkdir(preset, 0777); err != nil {
			fmt.Println(err)
		}

		createFile(job, "default.json")
		createFile(preset, "hls_high.json")
		createFile(preset, "hls_medium.json")
		createFile(preset, "hls_low.json")
	},
}

func createFile(path string, name string) {
	pathFile := filepath.Join(path, name)
	file, err := os.Create(pathFile)
	if err != nil {
		// Openエラー処理
	}
	defer file.Close()
	data, _ := Asset("default/" + name)
	file.Write(([]byte)(data))
}

var name string

func init() {
	RootCmd.AddCommand(newCmd)
	name = ""
	if len(os.Args) > 2 {
		name = os.Args[2]
	}
}
