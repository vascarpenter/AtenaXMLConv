/*
Package cmd command analyzer
Copyright © 2020 vascarpenter (gikoha5666@gmail.com)

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
	"fmt"
	"os"
	"strings"

	"github.com/eliben/gosax"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "AtenaXMLConv <filename.xml to convert>",
	Short: "Convert Atena 26 Contact XML into CSV",
	Long:  `Convert Atena 26 Contact XML into CSV`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires filename")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		analyze(args[0])
	},
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.AtenaXMLConv.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

		// Search config in home directory with name ".AtenaXMLConv" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".AtenaXMLConv")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// analyze : analyze user
func analyze(filename string) {
	columns := []string{"LastName", "FirstName", "furiLastName", "furiFirstName", "AddressCode", "FullAddress",
		"Suffix", "PhoneItem", "EmailItem", "Memo", "NamesOfFamily1", "X-Suffix1", "NamesOfFamily2", "X-Suffix2", "NamesOfFamily3", "X-Suffix3",
		"atxBaseYear", "X-NYCardHistory"}
	elemtype := ""
	exttype := ""
	oneline := make(map[string]string)

	for k := range columns {
		fmt.Printf("%s,", columns[k])
	}
	fmt.Printf("\n")

	// use sax;  fast simple xml parser
	scb := gosax.SaxCallbacks{
		StartElement: func(name string, attrs []string) {
			exttype = ""
			elemtype = name
			if name == "FirstName" || name == "LastName" {
				exttype = attrs[1] // ふりがな
			} else if name == "ExtensionItem" {
				exttype = attrs[3] // X-NYCardHistoryなど
			} else if name == "ContactXML" || name == "ContactXMLItem" ||
				name == "PersonName" || name == "PersonNameItem" || name == "ImageItem" ||
				name == "Address" || name == "AddressItem" || name == "Extension" || name == "Email" {
				elemtype = "NFSW"
			} else {

			}

		},

		EndElement: func(name string) {
			elemtype = ""
			if name == "ContactXMLItem" { // 一人分終了時
				// columns に登録されたコラムだけ CSVとして出力
				for k := range columns {
					fmt.Printf("%s,", oneline[columns[k]])
				}
				fmt.Printf("\n")
				oneline = make(map[string]string)
			}
		},

		Characters: func(contents string) {
			cont := strings.TrimSpace(contents)
			if elemtype == "NFSW" {

			} else if elemtype == "ExtensionItem" {
				if exttype == "NamesOfFamily" {
					// 複数回出現ありうる
					for i := 1; i < 10; i++ {
						str := fmt.Sprintf("NamesOfFamily%d", i)
						_, ok := oneline[str]
						if ok {
							// exist; next
						} else {
							// not exist
							oneline[str] = cont
							break
						}
					}
				} else {
					oneline[exttype] = cont
				}
			} else if elemtype == "FirstName" || elemtype == "LastName" {
				oneline[elemtype] = cont
				oneline["furi"+elemtype] = exttype
			} else if len(cont) != 0 {
				oneline[elemtype] = cont
			}
		},
	}

	err := gosax.ParseFile(filename, scb)
	if err != nil {
		panic(err)
	}

}
