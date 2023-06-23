/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: doRemove,
}

func doRemove(cmd *cobra.Command, args []string) {
	if !validConfig(cmd) {
		return
	}

	remove := []reqBasics{
		{
			label:  "dc_area-index",
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/dc_area", viper.GetString("url")),
			body:   http.NoBody,
		},
		{
			label:  "aircraft-index",
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/aircraft", viper.GetString("url")),
			body:   http.NoBody,
		},
		{
			label:  "alert_history",
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/alert_history", viper.GetString("url")),
			body:   http.NoBody,
		},
	}

	userName := viper.GetString("user")
	password := viper.GetString("password")
	client := http.Client{Timeout: 5 * time.Second}

	for _, v := range remove {
		req, err := http.NewRequest(v.method, v.url, v.body)
		if err != nil {
			log.Fatalf("failed to remove a part of the config (%s): %v", v.label, err)
		}

		req.SetBasicAuth(userName, password)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if resp != nil {
			defer resp.Body.Close()
		}

		if err != nil {
			log.Fatalf("failed to remove a part of the config (%s): %v", v.label, err)
		}

		if resp.StatusCode >= 400 {
			io.Copy(os.Stderr, resp.Body)
			if viper.GetBool("force") {
				continue
			}
			log.Fatalf("HTTP error from the server: [url:%s] [code:%v]: %v",
				v.url, resp.StatusCode, resp.Status)
		}

		log.Printf("Successfully removed %s\n", v.label)
	}
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	removeCmd.Flags().BoolP("force", "f", false, "ignore errors on removal")

	viper.BindPFlags(removeCmd.Flags())
}
