package main

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type App struct {
	pinCtrl           PinCtrl
	temperatureReader TemperatureReader
	heaterState       *HeaterState
}

func NewMockApp() App {
	return App{
		pinCtrl:           &MockPinCtrl{},
		temperatureReader: &MockSensorReader{},
		heaterState:       &HeaterState{},
	}
}

func NewApp() App {
	return App{
		pinCtrl:           &RaspberryPiPinCtrl{},
		temperatureReader: &SensorReader{},
		heaterState:       &HeaterState{},
	}
}

func (app *App) Initialize() {
	app.pinCtrl.InitializePin()
}

func (app *App) TearDown() {
	app.pinCtrl.TearDown()
}

func (app *App) StartWebServer() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
	}), middleware.Logger())
	e.GET("/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, app.heaterState)
	})
	e.POST("/force_on", func(c echo.Context) error {
		oneHourLater := GetLocalTime().Add(time.Hour * 1)
		app.heaterState.ForcedOnTimeLimit = &oneHourLater
		return c.JSON(http.StatusOK, app.heaterState)
	})
	e.DELETE("/force_on", func(c echo.Context) error {
		app.heaterState.ForcedOnTimeLimit = nil
		return c.JSON(http.StatusOK, app.heaterState)
	})
	e.POST("/economy_mode", func(c echo.Context) error {
		app.heaterState.EconoMode = true
		return c.JSON(http.StatusOK, app.heaterState)
	})
	e.DELETE("/economy_mode", func(c echo.Context) error {
		app.heaterState.EconoMode = false
		return c.JSON(http.StatusOK, app.heaterState)
	})
	e.POST("/disable", func(c echo.Context) error {
		app.heaterState.Disabled = true
		return c.JSON(http.StatusOK, app.heaterState)
	})
	e.DELETE("/disable", func(c echo.Context) error {
		app.heaterState.Disabled = false
		return c.JSON(http.StatusOK, app.heaterState)
	})

	if err := e.Start(":5000"); err != nil {
		e.Logger.Info("Error starting server")
	}
}

func (app *App) RefreshHeater() {
	app.heaterState.mu.Lock()
	localTime := GetLocalTime()
	temp, err := app.temperatureReader.ReadTemperature()
	if err != nil {
		log.Println("error reading temp", err)
		app.heaterState.HeaterOn = false
		app.pinCtrl.TurnHeaterOff()
	} else {
		app.heaterState.LastTempReading = temp
		switch app.heaterState.NextAction(localTime, temp) {
		case On:
			log.Println("turn heater on ", localTime, temp)
			app.heaterState.HeaterOn = true
			app.pinCtrl.TurnHeaterOn()
		case Off:
			log.Println("turn heater off ", localTime, temp)
			app.heaterState.HeaterOn = false
			app.pinCtrl.TurnHeaterOff()
		case NoAction:
			log.Println("don't do anything", localTime, temp)
		}
	}
	app.heaterState.mu.Unlock()
}

func GetLocalTime() time.Time {
	tz := "America/Los_Angeles"
	loc, _ := time.LoadLocation(tz)
	return time.Now().In(loc)
}

/*

// want an http REST client that is syncronous, updates the
// the heater state immediately then returns the new state

// want a continuous async process that updates the heater every 10 seconds

var pin18 rpio.Pin

type HeaterState struct {
	MinTemp           float64    `json:"min_temp"`
	MaxTemp           float64    `json:"max_temp"`
	EconoMode         bool       `json:"econo_mode"`
	Disabled          bool       `json:"disabled"`
	ForcedOnTimeLimit *time.Time `json:"forced_on_time_limit"`
	HeaterOn          bool       `json:"heater_on"`
	LastTempReading   float64    `json:"last_temp_reading"`
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


// you can just lock the heater while you make updates
hs := NewHeaterState()
	// set up a single observer and changer of heater state
	c := make(chan struct{})
	go func(chanPoll chan struct{}) {
		for {
			<-chanPoll
			hs.Refresh(pinCtrl)
		}
	}(c)
	// ask for state refresh every 10 seconds
	go func(chanPoll chan struct{}) {
		for {
			chanPoll <- struct{}{}
			time.Sleep(10 * time.Second)
		}
	}(c)

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
	}), middleware.Logger())
	e.GET("/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, hs)
	})
	e.POST("/force_on", func(c echo.Context) error {
		oneHourLater := GetLocalTime().Add(time.Hour * 1)
		hs.ForcedOnTimeLimit = &oneHourLater
		return c.JSON(http.StatusOK, hs)
	})
	e.DELETE("/force_on", func(c echo.Context) error {
		hs.ForcedOnTimeLimit = nil
		return c.JSON(http.StatusOK, hs)
	})
	e.POST("/economy_mode", func(c echo.Context) error {
		hs.EconoMode = true
		return c.JSON(http.StatusOK, hs)
	})
	e.DELETE("/economy_mode", func(c echo.Context) error {
		hs.EconoMode = false
		return c.JSON(http.StatusOK, hs)
	})
	e.POST("/disable", func(c echo.Context) error {
		hs.Disabled = true
		return c.JSON(http.StatusOK, hs)
	})
	e.DELETE("/disable", func(c echo.Context) error {
		hs.Disabled = false
		return c.JSON(http.StatusOK, hs)
	})

	if err := e.Start(":5000"); err != nil {
		e.Logger.Info("Error starting server")
	}

*/
