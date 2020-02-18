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
	"bufio"
	"fmt"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cdg",
	Short: "Tool to switch git folders",
	Long:  `Tool to provide helper functions to switch rapidly between git folders`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err        error
			cacheFile  string
			lines      []string
			currentDir string
		)

		currentDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		cacheFile, err = must("cache-file")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		lines, err = linesInFile(cacheFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		idx, err := fuzzyfinder.Find(
			lines,
			func(i int) string {
				return lines[i]
			})

		if idx <= 0 {
			fmt.Println(currentDir)
			os.Exit(0)
		}

		fmt.Println(lines[idx])
	},
}

func linesInFile(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	result := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}
	return result, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cdg.yaml)")

	cacheCmd.PersistentFlags().StringP("cache-file", "c", "", "File to write cache to")
	cacheCmd.MarkPersistentFlagRequired("cache-file")

	_ = viper.BindPFlag("cache-file", cacheCmd.Flags().Lookup("cache-file"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cdg" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cdg")
	}

	viper.AutomaticEnv() // read in environment variables that match

	_ = viper.ReadInConfig()
}
