package main

import (
	"log"
	"sync"
	"time"
)

const DefaultDaytimeTemp = 60.0
const DefaultNighttimeTemp = 71.8
const buffer = 0.2

type HeaterState struct {
	DesiredTemp          float64    `json:"desired_temp"`
	Enabled              bool       `json:"enabled"`
	CustomRangeTimeLimit *time.Time `json:"custom_range_time_limit"`
	LastTempReading      float64    `json:"last_temp_reading"`
	LastUpdatedAt        time.Time  `json:"last_updated_at`
	mu                   sync.Mutex
}

func NewHeaterState(localTime time.Time, temp float64) *HeaterState {
	newHeater := HeaterState{
		Enabled:              true,
		CustomRangeTimeLimit: nil,
		LastTempReading:      temp,
		LastUpdatedAt:        localTime,
	}
	newHeater.ResetDefaultTemp(localTime)
	return &newHeater
}

func (hs *HeaterState) ResetDefaultTemp(localTime time.Time) {
	if IsDayTime(localTime) {
		hs.DesiredTemp = DefaultDaytimeTemp
	} else {
		hs.DesiredTemp = DefaultNighttimeTemp
	}
}

func (hs *HeaterState) RefreshTimeAndTemp(newTime time.Time, newTemp float64) HeaterCommand {
	if hs.CustomRangeTimeLimit != nil {
		if newTime.After(*hs.CustomRangeTimeLimit) {
			hs.CustomRangeTimeLimit = nil // clear the timer
			hs.ResetDefaultTemp(newTime)
		}
	} else
	// going from daytime to night time
	if IsDayTime(hs.LastUpdatedAt) && !IsDayTime(newTime) {
		log.Println("daytime to nighttime", hs.LastUpdatedAt, newTime)
		hs.ResetDefaultTemp(newTime)
	} else if !IsDayTime(hs.LastUpdatedAt) && IsDayTime(newTime) {
		//night time to daytime
		log.Println("nighttime to daytime", hs.LastUpdatedAt, newTime)
		hs.ResetDefaultTemp(newTime)
	}
	hs.LastUpdatedAt = newTime
	hs.LastTempReading = newTemp

	if !hs.Enabled {
		return Off
	}
	if newTemp >= hs.DesiredTemp+buffer {
		return Off
	}
	if newTemp <= hs.DesiredTemp-buffer {
		return On
	}
	return NoAction
}

func (hs *HeaterState) SetDesiredTemp(temp float64) {
	hs.DesiredTemp = temp
}

func (hs *HeaterState) SetCustomRangeTimeOut(expireTime time.Time) {
	hs.CustomRangeTimeLimit = &expireTime
}

func (hs *HeaterState) Disable() {
	hs.Enabled = false
}

func (hs *HeaterState) Enable() {
	hs.Enabled = true

}

type HeaterCommand uint

const (
	Off HeaterCommand = iota
	On
	NoAction
)

func IsDayTime(t time.Time) bool {
	tz := "America/Los_Angeles"
	loc, _ := time.LoadLocation(tz)
	localTime := t.In(loc)
	if localTime.Hour() < 10 || localTime.Hour() >= 18 {
		return false
	}
	return true
}
