package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo"
	rpio "github.com/stianeikeland/go-rpio"
)

/*

Controls the heater with a GPIO Pin that powers a 3V relay

Auto mode - checks the temp in Haylie's room every 10 seconds
and keeps the temp between minTemp and maxTemp

EnergySaving mode during the day - does not regulate temp during the day
since people are dressed and moving about (night is after 7pm and before 10am)

EnergySaving mode maybe disabled until it's night time again

On - keeps heater on for 1 hour, then returns to program in progress

On (1 hr expire) > Auto (1 cycle expire) > EnergySavingAuto (default)

http server

index shows the current temp in Haylie's room
whether the heater is currently on or not
what mode the heater is in: On (time left) | Auto | EnergySavingAuto

Allow a switch between On | Auto | EnergySavingAuto

On should last a total of only 1 hour max.
Auto should return to EnergySavingAuto

*/

var pin18 rpio.Pin

type HeaterState struct {
	MinTemp           float64
	MaxTemp           float64
	EconoMode         bool
	ForcedOff         bool
	ForcedOnTimeLimit *time.Time
	HeaterOn          bool
}

func NewHeaterState() *HeaterState {
	return &HeaterState{
		MinTemp:           70.99,
		MaxTemp:           72.1,
		EconoMode:         true,
		ForcedOff:         false,
		ForcedOnTimeLimit: nil,
		HeaterOn:          false,
	}
}

type HeaterCommand uint

const (
	Off HeaterCommand = iota
	On
	NoAction
)

func (hs *HeaterState) CheckWhatToDo(pc PinCtrl) {
	for {
		localTime := GetLocalTime()
		temp, err := readTemp()
		if err != nil {
			hs.HeaterOn = false
			pc.TurnHeaterOff()
			time.Sleep(60 * time.Second)
		} else {
			switch hs.NextAction(localTime, temp) {
			case On:
				log.Println("turn heater on ", localTime, temp)
				hs.HeaterOn = true
				pc.TurnHeaterOn()
			case Off:
				log.Println("turn heater off ", localTime, temp)
				hs.HeaterOn = false
				pc.TurnHeaterOff()
			case NoAction:
				log.Println("don't do anything", localTime, temp)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func (hs *HeaterState) NextAction(localTime time.Time, temp float64) HeaterCommand {

	if hs.ForcedOff {
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
func GetLocalTime() time.Time {
	tz := "America/Los_Angeles"
	loc, _ := time.LoadLocation(tz)
	return time.Now().In(loc)
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

func readTemp() (temp float64, err error) {
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

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	pinCtrl := &PinCtrlMock{}
	pinCtrl.InitializePin()

	go func(pc PinCtrl) {
		<-sigs
		pc.TearDown()
		os.Exit(0)
	}(pinCtrl)

	hs := NewHeaterState()
	go hs.CheckWhatToDo(pinCtrl)

	e := echo.New()
	e.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("Heater On ? %v", hs.HeaterOn))
	})

	if err := e.Start(":5000"); err != nil {
		e.Logger.Info("Error starting server")
	}

}
