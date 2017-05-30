// Copyright © 2017 scrpgil
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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start transcoding.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if jobname == "" {
			jobname = "default"
		}
		job.SetConfigName(jobname)
		job.SetConfigType("json")
		job.AddConfigPath("job")
		if err := job.ReadInConfig(); err != nil {
			panic(fmt.Errorf("read jobfile error: %s \n", err))
		}
		inputDir := job.GetString("Input.Directory")
		filename := job.GetString("Input.FileName")
		format := job.GetString("Input.Format")

		pName := job.GetString("Playlists.Name")
		pDir := job.GetString("Playlists.Directory")

		if filename == "*" {
			list, err := ioutil.ReadDir(inputDir)
			if err != nil {
				fmt.Println("%v", err)
				os.Exit(1)
			}
			for _, finfo := range list {
				if finfo.IsDir() || -1 == strings.Index(finfo.Name(), "."+format) {
					continue
				}
				src := inputDir + "/" + finfo.Name()
				absPath, err := filepath.Abs(src)
				if err != nil {
					fmt.Println(absPath, err)
				}
				pos := strings.LastIndex(finfo.Name(), ".")
				fileName := finfo.Name()[:pos]
				path := pDir + "/" + fileName + "/" + pName + ".m3u8"
				createPlayList(pDir+"/"+fileName, path)
				Outputs(src, fileName, path, job.Get("Outputs").([]interface{}))
			}
		} else {
			path := pDir + "/" + filename + "/" + pName + ".m3u8"
			createPlayList(pDir+"/"+filename, path)
			src := inputDir + "/" + filename + "." + format
			absPath, err := filepath.Abs(src)
			if err != nil {
				fmt.Println(absPath, err)
			}
			Outputs(src, filename, path, job.Get("Outputs").([]interface{}))
		}
		return
	},
}

func Outputs(src string, filename string, path string, outputs []interface{}) {
	wg := &sync.WaitGroup{}
	for _, out := range outputs {
		key := out.(map[string]interface{})["Key"].(string)
		fmt.Println("start ", key)
		outputDir := out.(map[string]interface{})["Directory"].(string)
		if err := os.MkdirAll(outputDir+"/"+filename, 0777); err != nil {
			fmt.Println(err)
		}
		preset := out.(map[string]interface{})["PresetFile"].(string)
		wg.Add(1)
		go runCommand(preset, src, outputDir+"/"+filename, key, path, wg)
	}
	wg.Wait()
}

func runCommand(preset string, src string, output string, key string, path string, wg *sync.WaitGroup) {
	fmt.Println(output)
	config := viper.New()
	config.SetConfigName(preset)
	config.SetConfigType("json")
	config.AddConfigPath("preset")
	if err := config.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read jobfile error: %s \n", err))
	}

	flags := []string{
		"-i", src,
	}
	if str := config.GetString("ProfileV"); str != "" {
		flags = append(flags, "-profile:v")
		flags = append(flags, str)
	}
	if str := config.GetString("Level"); str != "" {
		flags = append(flags, "-level")
		flags = append(flags, str)
	}
	if str := config.GetString("Size"); str != "" {
		flags = append(flags, "-s")
		flags = append(flags, str)
	}
	if str := config.GetString("Scale"); str != "" {
		flags = append(flags, "-vf")
		flags = append(flags, "scale="+str)
	}
	if str := config.GetString("Aspect"); str != "" {
		flags = append(flags, "-aspect")
		flags = append(flags, str)
	}
	if str := config.GetString("StartNumber"); str != "" {
		flags = append(flags, "-start_number")
		flags = append(flags, str)
	}
	if str := config.GetString("HlsTime"); str != "" {
		flags = append(flags, "-hls_time")
		flags = append(flags, str)
	}
	if str := config.GetString("HlsListSize"); str != "" {
		flags = append(flags, "-hls_list_size")
		flags = append(flags, str)
	}
	if str := config.GetString("Format"); str != "" {
		flags = append(flags, "-format")
		flags = append(flags, str)
	}
	var bandwidth string
	if str := config.GetString("Bandwidth"); str != "" {
		bandwidth = str
	}
	flags = append(flags, output+"/"+key+".m3u8")
	addPlayList(path, bandwidth, key+".m3u8")
	ffmpeg := exec.Command(ffmpegPath, flags...)

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		fmt.Println(err)
	}

	ffmpeg.Start()

	_, err = ioutil.ReadAll(stderr)
	if err != nil {
		fmt.Println(err)
	}

	ffmpeg.Wait()
	wg.Done()
	return
}

func createPlayList(dir string, path string) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		fmt.Println(err)
	}
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	output := `#EXTM3U
`
	file.Write(([]byte)(output))
}

func addPlayList(path string, bandwidth string, filename string) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 777)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	output := `#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=` + bandwidth + `,CODECS="avc1.4d001f,mp4a.40.2"
` + filename + `
`
	fmt.Println(output)
	writer := bufio.NewWriter(file)
	writer.WriteString(output)
	writer.Flush()
}

var ffmpegPath string
var jobname string
var job *viper.Viper

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&jobname, "job", "j", "", "Read to job setting file")
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println(err)
	}
	ffmpegPath = path
	job = viper.New()
}
