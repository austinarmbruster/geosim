/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long:  `Capture the configuration`,
	Run:   doConfig,
}

func validConfig(cmd *cobra.Command) bool {
	user := viper.GetString("user")
	password := viper.GetString("password")
	url := viper.GetString("url")

	err := checkRequired(map[string]string{
		"user name": user,
		"password":  password,
		"url":       url,
	})

	if err != nil {
		cmd.PrintErrf("Missing configuration: %v\n", err)
		cmd.Usage()
		return false
	}
	return true
}

func doConfig(cmd *cobra.Command, args []string) {
	if !validConfig(cmd) {
		return
	}

	skip := map[string]bool{
		"overwrite": true,
		"force":     true,
		"config":    true,
	}

	toWrite := viper.New()
	toWrite.SetConfigName("geosim")        // name of config file (without extension)
	toWrite.SetConfigType("yaml")          // REQUIRED if the config file does not have the extension in the name
	toWrite.AddConfigPath("/etc/geosim/")  // path to look for the config file in
	toWrite.AddConfigPath("$HOME/.geosim") // call multiple times to add many search paths
	toWrite.AddConfigPath(".")             // optionally look for config in the working directory

	for _, k := range viper.AllKeys() {
		if skip[k] {
			continue
		}
		toWrite.Set(k, viper.Get(k))
	}

	var err error
	outputFileName := viper.GetString("config")

	if outputFileName == "" {
		err = toWrite.SafeWriteConfig()
		if err != nil && viper.GetBool("overwrite") {
			err = toWrite.WriteConfig()
		}
	} else {
		err = toWrite.SafeWriteConfigAs(outputFileName)
		if err != nil && viper.GetBool("overwrite") {
			err = toWrite.WriteConfigAs(outputFileName)
		}
	}

	if err != nil {
		fmt.Println(err)
	}
}

func checkRequired(inputs map[string]string) error {
	for k, v := range inputs {
		if v == "" {
			return fmt.Errorf("missing a required field: %v", k)
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringP("config", "c", "", "config file name")
	configCmd.Flags().BoolP("overwrite", "o", false, "overwrite the config file")

	viper.BindPFlags(configCmd.Flags())
}
