package main

import (
	"sync"
	"time"
)

type HeaterState struct {
	MinTemp           float64    `json:"min_temp"`
	MaxTemp           float64    `json:"max_temp"`
	EconoMode         bool       `json:"econo_mode"`
	Disabled          bool       `json:"disabled"`
	ForcedOnTimeLimit *time.Time `json:"forced_on_time_limit"`
	HeaterOn          bool       `json:"heater_on"`
	LastTempReading   float64    `json:"last_temp_reading"`
	mu                sync.Mutex
}

func NewHeaterState() *HeaterState {
	return &HeaterState{
		MinTemp:           70.99,
		MaxTemp:           72.1,
		EconoMode:         true,
		Disabled:          false,
		ForcedOnTimeLimit: nil,
		HeaterOn:          false,
		LastTempReading:   0.0,
	}
}

type HeaterCommand uint

const (
	Off HeaterCommand = iota
	On
	NoAction
)

func (hs *HeaterState) NextAction(localTime time.Time, temp float64) HeaterCommand {

	if hs.Disabled {
		return Off //no-op
	}
	if hs.ForcedOnTimeLimit != nil && localTime.Before(*hs.ForcedOnTimeLimit) {
		return On
	}

	if IsEconomyTime(localTime) && hs.EconoMode {
		return Off
	}
	if !IsEconomyTime(localTime) {
		hs.EconoMode = true
	}
	if temp > hs.MaxTemp {
		return Off
	}
	if temp < hs.MinTemp {
		return On
	}
	return NoAction
}

// IsEconomyTime returns true if it's daylight
func IsEconomyTime(t time.Time) bool {
	tz := "America/Los_Angeles"
	loc, _ := time.LoadLocation(tz)
	localTime := t.In(loc)
	if localTime.Hour() < 10 || localTime.Hour() >= 18 {
		return false
	}
	return true
}
