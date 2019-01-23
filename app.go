package main

import (
	"log"
	"net/http"
	"strconv"
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
	mpc := &MockPinCtrl{}
	msr := &MockSensorReader{}
	return App{
		pinCtrl:           mpc,
		temperatureReader: msr,
		heaterState:       NewHeaterState(GetLocalTime(), 71.0),
	}
}

func NewApp() App {
	sensorReader := SensorReader{}
	temp, _ := sensorReader.ReadTemperature()
	return App{
		pinCtrl:           &RaspberryPiPinCtrl{},
		temperatureReader: &sensorReader,
		heaterState:       NewHeaterState(GetLocalTime(), temp),
	}
}

func (app *App) Initialize() {
	app.pinCtrl.InitializePin()
}

func (app *App) TearDown() {
	app.pinCtrl.TearDown()
}

type State struct {
	MinTemp  float64 `json:"min_temp"`
	MaxTemp  float64 `json:"max_temp"`
	Temp     float64 `json:"temp"`
	Disabled bool    `json:"disabled"`
	HeaterOn bool    `json:"heater_on"`
}

func (app *App) GetState() State {
	return State{
		MinTemp:  app.heaterState.MinTemp,
		MaxTemp:  app.heaterState.MaxTemp,
		Temp:     app.heaterState.LastTempReading,
		Disabled: app.heaterState.Disabled,
		HeaterOn: app.pinCtrl.IsHeaterOn(),
	}
}

func (app *App) StartWebServer() {
	go func() {
		app.RefreshHeater()
		time.Sleep(10 * time.Second)
	}()
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
	}), middleware.Logger())
	e.GET("/status", func(c echo.Context) error {
		app.RefreshHeater()
		return c.JSON(http.StatusOK, app.GetState())
	})
	e.POST("/disable", func(c echo.Context) error {
		app.heaterState.Disable()
		app.RefreshHeater()
		return c.JSON(http.StatusOK, app.GetState())
	})
	e.POST("/enable", func(c echo.Context) error {
		app.heaterState.Enable()
		app.RefreshHeater()
		return c.JSON(http.StatusOK, app.GetState())
	})
	e.POST("/set_max_temp/:max_temp", func(c echo.Context) error {
		maxTempParam := c.Param("max_temp")
		maxTemp, err := strconv.ParseFloat(maxTempParam, 64)
		if err != nil {
			return c.JSON(http.StatusNotAcceptable, err)
		}
		app.heaterState.SetMaxTemp(maxTemp)
		if maxTemp <= app.heaterState.MinTemp {
			app.heaterState.SetMinTemp(maxTemp - 0.1)
		}
		app.heaterState.SetCustomRangeTimeOut(GetLocalTime().Add(1 * time.Hour))
		app.RefreshHeater()
		return c.JSON(http.StatusOK, app.GetState())
	})
	e.POST("/set_min_temp/:min_temp", func(c echo.Context) error {
		minTempParam := c.Param("min_temp")
		minTemp, err := strconv.ParseFloat(minTempParam, 64)
		if err != nil {
			return c.JSON(http.StatusNotAcceptable, err)
		}
		app.heaterState.SetMinTemp(minTemp)
		if minTemp >= app.heaterState.MaxTemp {
			app.heaterState.SetMaxTemp(minTemp + 0.1)
		}
		app.heaterState.SetCustomRangeTimeOut(GetLocalTime().Add(1 * time.Hour))
		app.RefreshHeater()
		return c.JSON(http.StatusOK, app.GetState())
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
		log.Println("error", err)
		app.pinCtrl.TurnHeaterOff()
	} else {
		switch app.heaterState.RefreshTimeAndTemp(localTime, temp) {
		case On:
			log.Println("turn heater on ", localTime, temp)
			app.pinCtrl.TurnHeaterOn()
		case Off:
			log.Println("turn heater off ", localTime, temp)
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
