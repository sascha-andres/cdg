/*
Copyright © 2020 Sascha Andres <sascha.andres@outlook.com>

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
	"fmt"
	"github.com/apoorvam/goterminal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Build cache",
	Long:  `Use command to create a cache file with your git repositories inside`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, cacheFile := getAndValidate()

		var (
			directories = make([]string, 0)
			err         error
		)

		// initialize channels
		c := make(chan string, 2)
		d := make(chan bool)
		defer close(c)
		defer close(d)

		fmt.Printf("Scanning %s for git repositories", rootPath)

		// walk through directory
		go func() {
			err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
				if info == nil {
					return nil
				}
				if info.IsDir() && info.Name() == ".git" {
					p := strings.TrimSuffix(path, info.Name())
					c <- p
				}
				return nil
			})
			d <- true
		}()

		writer := goterminal.New(os.Stdout)
		for {
			exit := false
			select {
			case gitDirectory := <-c:
				writer.Clear()
				directories = append(directories, gitDirectory)
				fmt.Fprintf(writer, "found: %s\n", gitDirectory)
				writer.Print()
			case <-d:
				exit = true
			}
			if exit {
				break
			}
		}

		if err != nil {
			os.Exit(1)
		}
		writer.Clear()

		var sb strings.Builder

		for _, dir := range directories {
			_, _ = sb.WriteString(fmt.Sprintf("%s\n", dir))
		}

		err = ioutil.WriteFile(cacheFile, []byte(sb.String()), 0600)
		if err != nil {
			os.Exit(1)
		}

		fmt.Fprintf(writer, "wrote %d directories to cache", len(directories))
		writer.Print()

		os.Exit(0)
	},
}

func getAndValidate() (string, string) {
	var (
		rootPath  string
		cacheFile string
		err       error
	)

	rootPath, err = must("root-path")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cacheFile, err = must("cache-file")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return rootPath, cacheFile
}

func init() {
	rootCmd.AddCommand(cacheCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	cacheCmd.Flags().StringP("root-path", "p", ".", "Path to scan for git repositories")
	cacheCmd.Flags().StringP("cache-file", "c", "", "File to write cache to")

	//cacheCmd.MarkFlagRequired("path")

	_ = viper.BindPFlag("root-path", cacheCmd.Flags().Lookup("root-path"))
}

func must(argument string) (string, error) {
	value := viper.GetString(argument)
	if "" == value {
		return "", fmt.Errorf("%s not provided", argument)
	}
	return value, nil
}
