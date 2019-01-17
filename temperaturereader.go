package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
)

type TemperatureReader interface {
	ReadTemperature() (temp float64, err error)
}

type SensorReader struct {
}

func (s *SensorReader) ReadTemperature() (temp float64, err error) {
	resp, err := http.Get("http://tsensor:5000/")
	if err != nil {
		return temp, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return temp, err
	}
	return strconv.ParseFloat(string(body), 64)
}

type MockSensorReader struct{}

func (ms *MockSensorReader) ReadTemperature() (temp float64, err error) {
	return 75.99, nil
}
