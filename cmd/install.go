/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/savaki/jq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: doInstall,
}

type reqBasics struct {
	label       string
	method      string
	url         string
	contentType string
	body        io.Reader
}

func doInstall(cmd *cobra.Command, args []string) {
	if !validConfig(cmd) {
		return
	}

	err := doLoad()
	if err != nil {
		log.Fatal(err)
	}

	err = doRuleEnable()
	if err != nil {
		log.Fatal(err)
	}

}

func doRuleEnable() error {
	client := http.Client{Timeout: 5 * time.Second}

	rawData, err := getRules(&client)
	if err != nil {
		return fmt.Errorf("error getting rules: %w", err)
	}

	for i := 0; i < 4; i++ {
		err := enableRule(&client, rawData, i)
		if err != nil {
			return fmt.Errorf("error enabling rule %v: %w", i, err)
		}
	}

	return nil
}

func getRules(client *http.Client) ([]byte, error) {
	url := fmt.Sprintf("%s/api/alerting/rules/_find", viper.GetString("kibana-url"))
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	userName := viper.GetString("user")
	password := viper.GetString("password")
	req.SetBasicAuth(userName, password)
	req.Header.Set("kbn-xsrf", "true")

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	buff := &bytes.Buffer{}
	_, err = io.Copy(buff, resp.Body)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func enableRule(client *http.Client, rawData []byte, ruleNumber int) error {
	op, err := jq.Parse(fmt.Sprintf(".data.[%v].id", ruleNumber))
	if err != nil {
		return err
	}

	data, err := op.Apply(rawData)
	if err != nil {
		return err
	}
	id := strings.Replace(string(data), "\"", "", -1)
	fmt.Println(id)

	url := fmt.Sprintf("%s/api/alerting/rule/%s/_enable", viper.GetString("kibana-url"), id)
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		return err
	}

	userName := viper.GetString("user")
	password := viper.GetString("password")
	req.SetBasicAuth(userName, password)
	req.Header.Set("kbn-xsrf", "true")

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	}
	io.Copy(os.Stdout, resp.Body)

	// TODO confirm that the rule was enabled
	return nil
}

func doLoad() error {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	fw, err := w.CreateFormFile("file", "export.ndjson")
	if err != nil {
		panic(err)
	}
	fw.Write([]byte(savedObjects))
	w.Close()

	setup := []reqBasics{
		{
			label:       "alert_history-index",
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/alert_history", viper.GetString("url")),
			contentType: "application/json",
			body:        strings.NewReader(alertHistoryIdxConfig),
		},
		{
			label:       "aircraft-index",
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/aircraft", viper.GetString("url")),
			contentType: "application/json",
			body:        strings.NewReader(aircraftIdxConfig),
		},
		{
			label:       "dc_area-index",
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/dc_area", viper.GetString("url")),
			contentType: "application/json",
			body:        strings.NewReader(dcAreaIdxConfig),
		},
		{
			label:       "dc-doc",
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/dc_area/_doc", viper.GetString("url")),
			contentType: "application/json",
			body:        strings.NewReader(dcDoc),
		},
		{
			label:       "south_of_dc-doc",
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/dc_area/_doc", viper.GetString("url")),
			contentType: "application/json",
			body:        strings.NewReader(southOfDCDoc),
		},
		{
			label:       "saved_objects-kibana",
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/api/saved_objects/_import?overwrite=true", viper.GetString("kibana-url")),
			contentType: w.FormDataContentType(),
			body:        b,
		},
	}

	userName := viper.GetString("user")
	password := viper.GetString("password")
	client := http.Client{Timeout: 5 * time.Second}

	for _, v := range setup {
		req, err := http.NewRequest(v.method, v.url, v.body)
		if err != nil {
			log.Fatalf("failed to create a part of the config (%s): %v", v.label, err)
		}

		req.SetBasicAuth(userName, password)
		req.Header.Set("Content-Type", v.contentType)
		req.Header.Set("kbn-xsrf", "true")

		resp, err := client.Do(req)
		if resp != nil {
			defer resp.Body.Close()
		}

		if err != nil {
			return fmt.Errorf("failed to create a part of the config (%s): %v", v.label, err)
		}

		io.Copy(os.Stderr, resp.Body)
		if resp.StatusCode >= 400 {
			log.Fatalf("\n\nHTTP error from the server: [url:%s] [code:%v]: %v",
				v.url, resp.StatusCode, resp.Status)
		}

		log.Printf("\nSuccessfully installed %s\n", v.label)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}

//go:embed install-files/aircraft-index.json
var aircraftIdxConfig string

//go:embed install-files/dc_area-index.json
var dcAreaIdxConfig string

//go:embed install-files/dc-doc.json
var dcDoc string

//go:embed install-files/south_of_dc-doc.json
var southOfDCDoc string

//go:embed install-files/alert_history-index.json
var alertHistoryIdxConfig string

//go:embed install-files/export.ndjson
var savedObjects string
