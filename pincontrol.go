package main

import (
	"fmt"
	"log"

	rpio "github.com/stianeikeland/go-rpio"
)

type PinCtrl interface {
	InitializePin()
	TurnHeaterOn()
	TurnHeaterOff()
	TearDown()
}

type PinCtrlRaspberryPi struct{}

func (pcrp *PinCtrlRaspberryPi) InitializePin() {
	err := rpio.Open()
	if err != nil {
		log.Println("Not a raspberry pi?")
		log.Fatal(err)
	}
	pin18 = rpio.Pin(18)
	pin18.Output()
	pin18.Low()
	log.Println("Pin Initialized")
}

func (pcrp *PinCtrlRaspberryPi) TurnHeaterOn() {
	pin18.High()
}

func (pcrp *PinCtrlRaspberryPi) TurnHeaterOff() {
	pin18.Low()
}

func (pcrp *PinCtrlRaspberryPi) TearDown() {
	pin18.Low()
	rpio.Close()
	fmt.Println("shutting down")
}

type PinCtrlMock struct{}

func (pcm *PinCtrlMock) InitializePin() {
	log.Println("Mock initialize gpio pin")
}
func (pcm *PinCtrlMock) TurnHeaterOn() {
	log.Println("Mock turn heater on")
}
func (pcm *PinCtrlMock) TurnHeaterOff() {
	log.Println("Mock turn heater off")
}
func (pcm *PinCtrlMock) TearDown() {
	log.Println("Mock tear down")
}
