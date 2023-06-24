/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "sends geo data into Elasticsearch",
	Long:  `Provides a simple data generator to send locations into Elasticsearch.`,
	Run:   doRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func doRun(cmd *cobra.Command, args []string) {
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
		return
	}

	sim := &simulator{
		baseURL:  url,
		userName: user,
		password: password,
	}

	sim.execute()
}

type simulator struct {
	baseURL  string
	userName string
	password string
}

type location struct {
	Lat float64 `json:"lat,omitempty"`
	Lon float64 `json:"lon,omitempty"`
}

type entity struct {
	Timestamp   time.Time `json:"@timestamp,omitempty"`
	Name        string    `json:"name,omitempty"`
	Location    location  `json:"location,omitempty"`
	Tag         []string  `json:"tag,omitempty"`
	centerPoint location  `json:"-,omitempty"`
	action      action    `json:"-,omitempty"`
	mover       movement  `json:"-,omitempty"`
}

type action func(entity *entity) error

type movement interface {
	move(last location) (next location, when time.Time, err error)
}

type circular struct {
	centerPoint location
	radius      float64
	angle       float64
	theta       float64
}

func (c *circular) move(_ location) (next location, when time.Time, err error) {
	c.angle += c.theta
	next = location{
		Lat: c.centerPoint.Lat + math.Cos(c.angle)*c.radius,
		Lon: c.centerPoint.Lon + math.Sin(c.angle)*c.radius,
	}
	when = time.Now()
	return next, when, nil
}

func (e *entity) tick() error {
	next, when, err := e.mover.move(e.Location)
	if err != nil {
		return err
	}

	e.Location = next
	e.Timestamp = when

	// TODO: this api needs to change
	e.action(e)

	return nil
}

func New(name string, lat, lon float64, a action) *entity {
	thetaDelta := float64(15) / 180 * math.Pi
	mover := circular{
		centerPoint: location{
			Lat: lat,
			Lon: lon,
		},
		radius: float64(0.08),
		angle:  float64(0.0),
		theta:  thetaDelta,
	}

	e := &entity{
		Name: name,
		Tag:  []string{name, "common"},
	}

	e.action = a
	e.mover = &mover

	return e
}

func (s *simulator) execute() {

	assets := make([]*entity, 4)
	assets[0] = New("alpha", 38.804935386594465, -77.02314564908244, s.post)
	assets[1] = New("beta", 38.93466616893009, -77.12009876227299, s.post)
	assets[2] = New("charlie", 38.99534659899706, -77.04127967221889, s.post)
	assets[3] = New("delta", 38.893277091526166, -76.91043222058185, s.post)

	thetaDelta := float64(15) / 180 * math.Pi

	angle := thetaDelta
	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		tick(assets, angle)
		angle += thetaDelta
	}
}

func (s *simulator) post(e *entity) error {
	jBytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	buff := bytes.NewBuffer(jBytes)
	client := http.Client{Timeout: 5 * time.Second}

	url := fmt.Sprintf("%s/aircraft/_doc", s.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, buff)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(s.userName, s.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode > 300 {
		fmt.Printf("resp: %v\n", resp.Status)
		io.Copy(os.Stdout, resp.Body)
	}

	if err != nil {
		return err
	}

	return nil
}

func tick(assets []*entity, angle float64) {
	for _, e := range assets {
		e.tick()
	}
}
