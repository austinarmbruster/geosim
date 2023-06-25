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

// TODO split the rest of this into another package

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
	Timestamp time.Time `json:"@timestamp,omitempty"`
	Name      string    `json:"name,omitempty"`
	Location  location  `json:"location,omitempty"`
	Tag       []string  `json:"tag,omitempty"`
	action    action
	mover     movement
}

type action func(entity *entity) error

func noopAction(_ *entity) error {
	return nil
}

type movement interface {
	move(last location) (next location, when time.Time, err error)
}

type noopMover struct{}

func (n noopMover) move(last location) (next location, when time.Time, err error) {
	return last, time.Now(), nil
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

	e.action(e)

	return nil
}

type entityOption func(*entity)

func WithBasicCircleMover(lat, lon float64) entityOption {
	return func(e *entity) {
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

		e.mover = &mover
	}
}

func WithSetPathMover(locations []location, offsets []time.Duration) entityOption {
	return func(e *entity) {
		e.mover = NewSetPathMover(locations, offsets)
	}
}

type setPathMover struct {
	locations  []location
	timeOffets []time.Duration
	i          int
}

func NewSetPathMover(locations []location, timeOffets []time.Duration) *setPathMover {
	m := &setPathMover{
		locations:  locations,
		timeOffets: timeOffets,
		i:          -1,
	}
	return m
}

func (m *setPathMover) move(last location) (next location, when time.Time, err error) {
	m.i++
	if m.i == math.MaxInt {
		m.i = 0
	}
	next = m.loc(last)
	when = m.time()

	return next, when, nil
}

func (m *setPathMover) loc(last location) location {
	if len(m.locations) == 0 {
		return last
	}

	return m.locations[m.i%len(m.locations)]
}

func (m *setPathMover) time() time.Time {
	if len(m.timeOffets) == 0 {
		return time.Now()
	}

	return time.Now().Add(m.timeOffets[m.i%len(m.timeOffets)])
}

func WithTag(tag string) entityOption {
	return func(e *entity) {
		e.Tag = append(e.Tag, tag)
	}
}

func WithAction(a action) entityOption {
	return func(e *entity) {
		e.action = a
	}
}

func New(name string, opts ...entityOption) *entity {

	e := &entity{
		Name:   name,
		Tag:    make([]string, 0, 2),
		mover:  noopMover{},
		action: noopAction,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (s *simulator) execute() {

	assets := make([]*entity, 0, 7)
	assets = append(assets,
		New("delta",
			WithTag("delta"),
			WithTag("common"),
			WithBasicCircleMover(38.893277091526166, -76.91043222058185),
			WithAction(s.post),
		))
	assets = append(assets,
		New("alpha",
			WithTag("alpha"),
			WithTag("common"),
			WithBasicCircleMover(38.804935386594465, -77.02314564908244),
			WithAction(s.post),
		))
	assets = append(assets,
		New("beta",
			WithTag("beta"),
			WithTag("common"),
			WithBasicCircleMover(38.93466616893009, -77.12009876227299),
			WithAction(s.post),
		))
	assets = append(assets,
		New("charlie",
			WithTag("charlie"),
			WithTag("common"),
			WithBasicCircleMover(38.99534659899706, -77.04127967221889),
			WithAction(s.post),
		))
	assets = append(assets,
		New("delta",
			WithTag("delta"),
			WithTag("common"),
			WithBasicCircleMover(38.893277091526166, -76.91043222058185),
			WithAction(s.post),
		))

	a := location{
		Lat: 38.81451,
		Lon: -77.16904,
	}
	b := location{
		Lat: 38.89283,
		Lon: -77.03948,
	}
	c := location{
		Lat: 38.97388,
		Lon: -76.89814,
	}

	tickDuration := 3 * time.Second
	repeatCount := 7

	assets = append(assets,
		New("abc",
			WithTag("abc"),
			WithTag("order-check"),
			WithTag("in-order"),
			WithSetPathMover(
				repeatedLocs([]location{a, b, c}, repeatCount),
				[]time.Duration{0 * time.Second}),
			WithAction(s.post),
		))

	assets = append(assets,
		New("cab",
			WithTag("cab"),
			WithTag("order-check"),
			WithTag("last-first"),
			WithSetPathMover(
				repeatedLocs([]location{c, a, b}, repeatCount),
				repeatedTimes(3, repeatCount, []int{0, -2, -1}, tickDuration)),
			WithAction(s.post),
		))

	assets = append(assets,
		New("cba",
			WithTag("cba"),
			WithTag("order-check"),
			WithTag("reverse-order"),
			WithSetPathMover(
				repeatedLocs([]location{a, b, c}, repeatCount),
				repeatedTimes(3, repeatCount, []int{0, -1, -2}, tickDuration)),
			WithAction(s.post),
		))

	ticker := time.NewTicker(tickDuration)
	for range ticker.C {
		tick(assets)
	}
}

func repeatedLocs(locs []location, cnt int) []location {
	rtnVal := make([]location, 0, len(locs)*cnt)

	for _, l := range locs {
		for i := 0; i < cnt; i++ {
			rtnVal = append(rtnVal, l)
		}
	}

	return rtnVal
}

func repeatedTimes(numPositions int, repeatCount int, orderOffsetFromFirst []int, tickDuration time.Duration) []time.Duration {
	rtnVal := make([]time.Duration, 0, numPositions*repeatCount)
	for p := 0; p < numPositions; p++ {
		for i := 0; i < repeatCount; i++ {
			d := time.Duration((orderOffsetFromFirst[p]-p)*repeatCount*int(tickDuration.Milliseconds())) * time.Millisecond
			rtnVal = append(rtnVal, d)
		}
	}
	return rtnVal
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

func tick(assets []*entity) {
	for _, e := range assets {
		e.tick()
	}
}
