/*
Copyright © 2020 Ricardo Hernandez <richerve@gmail.com>

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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states/statefile"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type Migration struct {
	From string `json:"from"`
	To   string `json:"to,omitempty"`
}

type terramoveConfig struct {
	Migrations []Migration `json:"migrations"`
}

var config terramoveConfig
var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terramove",
	Short: "terramove will help you in importing resources to a new terraform state",
	Run:   generateTerraformImportsRun,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terramove.yaml)")
	// rootCmd.MarkPersistentFlagRequired("config")
	rootCmd.Flags().StringP("state-file", "s", "", "Terraform state file path")
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

		// Search config in home directory with name ".terramove" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".terramove")

	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// If a config file is found, read it into terramoveConfig
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
}

// InstanceAttributes is a convenient type to extract only the id attribute when unmarshalling
type InstanceAttributes struct {
	Id string
}

func getInstanceAttributes(attrs []byte) (InstanceAttributes, error) {
	iattrs := InstanceAttributes{}

	err := json.Unmarshal(attrs, &iattrs)
	if err != nil {
		return iattrs, err
	}

	return iattrs, nil

}

func generateTerraformImportsRun(cmd *cobra.Command, args []string) {

	stateFilePath, err := cmd.Flags().GetString("state-file")
	if err != nil {
		fmt.Printf("Error reading state-file flag, %v", err)
	}

	stateFile, err := os.Open(stateFilePath)
	if err != nil {
		panic("Error opening the state file")
	}

	tfstate, err := statefile.Read(stateFile)
	if err != nil {
		fmt.Printf("Error reading the file as a terraform state file, %v", err)
	}

	// Iterate over the "from" keys to get the ids on the source state
	for _, migration := range config.Migrations {
		// ParseAbsREsourceInstanceSrt returns a tfdias.Diagnostics type of value that we don't need
		// so we drop that
		absResource, _ := addrs.ParseAbsResourceInstanceStr(migration.From)

		currentResource := tfstate.State.ResourceInstance(absResource).Current

		iattrs, err := getInstanceAttributes(currentResource.AttrsJSON)
		if err != nil {
			fmt.Printf("Error getting instance attributes, %v", err)
		}

		// By default the `to` field is set to `from` as the first can be omitted
		to := migration.From

		if migration.To != "" {
			to = migration.To
		}

		fmt.Printf("terraform import %q %q\n", to, iattrs.Id)
	}
}
