/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "geosim",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: doRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("url", "U", "", "Elasticsearch URL")
	rootCmd.PersistentFlags().StringP("kibana-url", "K", "", "Kibana URL")
	rootCmd.PersistentFlags().StringP("user", "u", "", "Elasticsearch user name")
	rootCmd.PersistentFlags().StringP("password", "p", "", "Elasticsearch password")

	viper.SetConfigName("geosim-config") // name of config file (without extension)
	viper.SetConfigType("yaml")          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/geosim/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.geosim") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	viper.BindPFlags(rootCmd.PersistentFlags())

	viper.SetEnvPrefix("geosim")
	viper.AutomaticEnv()
}
