package main

import (
	"log"
	"sync"
	"time"
)

const DefaultDaytimeMinTemp = 60.0
const DefaultDaytimeMaxTemp = 65.0
const DefaultNighttimeMinTemp = 70.99
const DefaultNighttimeMaxTemp = 72.1

type HeaterState struct {
	MinTemp              float64    `json:"min_temp"`
	MaxTemp              float64    `json:"max_temp"`
	Disabled             bool       `json:"disabled"`
	CustomRangeTimeLimit *time.Time `json:"custom_range_time_limit"`
	// HeaterOn             bool       `json:"heater_on"`
	LastTempReading float64   `json:"last_temp_reading"`
	LastUpdatedAt   time.Time `json:"last_updated_at`
	// pinCtrl              PinCtrl
	// tempReader           TemperatureReader
	mu sync.Mutex
}

func NewHeaterState() *HeaterState {
	return &HeaterState{
		MinTemp:              DefaultDaytimeMinTemp,
		MaxTemp:              DefaultDaytimeMaxTemp,
		Disabled:             false,
		CustomRangeTimeLimit: nil,
		// HeaterOn:             false,
		LastTempReading: 73.3,
		LastUpdatedAt:   time.Now(),
		// pinCtrl:              pc,
		// tempReader:           tr,
	}
}

func (hs *HeaterState) RefreshTimeAndTemp(newTime time.Time, newTemp float64) HeaterCommand {
	if hs.CustomRangeTimeLimit != nil {
		if newTime.After(*hs.CustomRangeTimeLimit) {
			hs.CustomRangeTimeLimit = nil // clear the timer
			if IsDayTime(newTime) {
				hs.MinTemp = DefaultDaytimeMinTemp
				hs.MaxTemp = DefaultDaytimeMaxTemp
			} else {
				hs.MinTemp = DefaultNighttimeMinTemp
				hs.MaxTemp = DefaultNighttimeMaxTemp
			}
		}
	} else
	// going from daytime to night time
	if IsDayTime(hs.LastUpdatedAt) && !IsDayTime(newTime) {
		log.Println("daytime to nighttime", hs.LastUpdatedAt, newTime)
		hs.MinTemp = DefaultNighttimeMinTemp
		hs.MaxTemp = DefaultNighttimeMaxTemp
	} else if !IsDayTime(hs.LastUpdatedAt) && IsDayTime(newTime) {
		//night time to daytime
		log.Println("nighttime to daytime", hs.LastUpdatedAt, newTime)
		hs.MinTemp = DefaultDaytimeMinTemp
		hs.MaxTemp = DefaultDaytimeMaxTemp
	}
	hs.LastUpdatedAt = newTime
	hs.LastTempReading = newTemp
	if hs.Disabled {
		return Off
	}
	if newTemp > hs.MaxTemp {
		return Off
	}
	if newTemp < hs.MinTemp {
		return On
	}
	return NoAction
}

func (hs *HeaterState) SetMaxTemp(temp float64) {
	hs.MaxTemp = temp

}

func (hs *HeaterState) Disable() {
	hs.Disabled = true
}

func (hs *HeaterState) Enable() {
	hs.Disabled = false

}

type HeaterCommand uint

const (
	Off HeaterCommand = iota
	On
	NoAction
)

// func (hs *HeaterState) StartRefreshRoutine() {
// 	// go func() {
// 	// 	for {
// 	// 		// app.RefreshHeater()
// 	// 		time.Sleep(10 * time.Second)
// 	// 	}
// 	// }()
// }

// func (hs *HeaterState) NextAction(localTime time.Time, temp float64) HeaterCommand {

// 	if hs.Disabled {
// 		return Off //no-op
// 	}
// 	if temp > hs.MaxTemp {
// 		return Off
// 	}
// 	if temp < hs.MinTemp {
// 		return On
// 	}
// 	return NoAction
// }

func IsDayTime(t time.Time) bool {
	tz := "America/Los_Angeles"
	loc, _ := time.LoadLocation(tz)
	localTime := t.In(loc)
	if localTime.Hour() < 10 || localTime.Hour() >= 18 {
		return false
	}
	return true
}
