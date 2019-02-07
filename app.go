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
		heaterState:       NewHeaterState(GetLocalTime(), 71.234234324230),
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
	DesiredTemp float64 `json:"desired_temp"`
	Temp        float64 `json:"temp"`
	Enabled     bool    `json:"enabled"`
	HeaterOn    bool    `json:"heater_on"`
}

func (app *App) GetState() State {
	return State{
		DesiredTemp: app.heaterState.DesiredTemp,
		Temp:        app.heaterState.LastTempReading,
		Enabled:     app.heaterState.Enabled,
		HeaterOn:    app.pinCtrl.IsHeaterOn(),
	}
}

func (app *App) StartWebServer() {
	go func() {
		for {
			app.RefreshHeater(true)
			time.Sleep(10 * time.Second)
		}
	}()
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
	}), middleware.Logger())
	e.Static("/", "dist")

	e.GET("/status", func(c echo.Context) error {
		app.RefreshHeater(false)
		return c.JSON(http.StatusOK, app.GetState())
	})
	e.POST("/disable", func(c echo.Context) error {
		app.heaterState.Disable()
		app.RefreshHeater(false)
		return c.JSON(http.StatusOK, app.GetState())
	})
	e.POST("/enable", func(c echo.Context) error {
		app.heaterState.Enable()
		app.RefreshHeater(false)
		return c.JSON(http.StatusOK, app.GetState())
	})
	e.POST("/set_desired_temp/:desired_temp", func(c echo.Context) error {
		desiredTempParam := c.Param("desired_temp")
		desiredTemp, err := strconv.ParseFloat(desiredTempParam, 64)
		if err != nil {
			return c.JSON(http.StatusNotAcceptable, err)
		}
		app.heaterState.SetDesiredTemp(desiredTemp)
		app.heaterState.SetCustomRangeTimeOut(GetLocalTime().Add(2 * time.Hour))
		app.RefreshHeater(false)
		return c.JSON(http.StatusOK, app.GetState())
	})

	if err := e.Start(":5000"); err != nil {
		e.Logger.Info("Error starting server")
	}
}

func (app *App) GetTemp(getNewTemp bool) (float64, error) {
	if getNewTemp {
		return app.temperatureReader.ReadTemperature()
	}
	return app.heaterState.LastTempReading, nil
}

func (app *App) RefreshHeater(getNewTemp bool) {
	app.heaterState.mu.Lock()
	localTime := GetLocalTime()
	temp, err := app.GetTemp(getNewTemp)
	if err != nil {
		log.Println("error", err)
		app.pinCtrl.TurnHeaterOff()
	} else {
		switch app.heaterState.RefreshTimeAndTemp(localTime, temp) {
		case On:
			log.Println("turn heater on ", localTime, temp, app.heaterState.DesiredTemp)
			app.pinCtrl.TurnHeaterOn()
		case Off:
			log.Println("turn heater off ", localTime, temp, app.heaterState.DesiredTemp)
			app.pinCtrl.TurnHeaterOff()
		case NoAction:
			log.Println("don't do anything", localTime, temp, app.heaterState.DesiredTemp)
		}
	}
	app.heaterState.mu.Unlock()
}

func GetLocalTime() time.Time {
	tz := "America/Los_Angeles"
	loc, _ := time.LoadLocation(tz)
	return time.Now().In(loc)
}
