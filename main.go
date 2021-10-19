package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

type location struct {
	Lat float64 `json:"lat,omitempty"`
	Lon float64 `json:"lon,omitempty"`
}

type entity struct {
	Timestamp   time.Time `json:"@timestamp,omitempty"`
	Name        string    `json:"name,omitempty"`
	Location    location  `json:"location,omitempty"`
	centerPoint location  `json:"-"`
	action action `json:"-"`
	mover movement `json:"-"`
}

type action func(entity *entity) error

type movement interface{
	move(last location) (next location, when time.Time, err error)
}

type circular struct {
	centerPoint location
	radius float64
	angle float64
}

func (c circular) move(_ location) (next location, when time.Time, err error) {
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

func New(name string, lat, lon float64) *entity {
	mover := circular{
		centerPoint: location {
			Lat: lat,
			Lon: lon,
		},
		radius: float64(0.08),
		angle: float64(0),
	}

	e := &entity{
		Name: name,
	}

	e.action = post
	e.mover = mover

	return e
}

func rotate(e *entity, radius, angle float64) {
	e.Location.Lat = e.centerPoint.Lat + math.Cos(angle)*radius
	e.Location.Lon = e.centerPoint.Lon + math.Sin(angle)*radius
	e.Timestamp = time.Now()
}

func print(e *entity) {
	jBytes, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jBytes))
}

func main() {

	assets := make([]*entity, 4)
	assets[0] = New("alpha", 38.804935386594465, -77.02314564908244)
	assets[1] = New("beta", 38.93466616893009, -77.12009876227299)
	assets[2] = New("charlie", 38.99534659899706, -77.04127967221889)
	assets[3] = New("delta", 38.893277091526166, -76.91043222058185)

	thetaDelta := float64(15) / 180 * math.Pi

	angle := thetaDelta
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			tick(assets,angle)
			angle += thetaDelta
		}
	}
}

func post(e *entity) error {
	jBytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	buff := bytes.NewBuffer(jBytes)
	resp, err := http.Post("https://elastic:zE2AARvtp5GLLRJsc6WEvZ61@geo.es.eastus2.azure.elastic-cloud.com:9243/file_points/_doc", "application/json", buff)
	if resp != nil {
		defer resp.Body.Close()
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
