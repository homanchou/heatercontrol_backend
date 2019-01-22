package main

import (
	"fmt"
	"log"

	"github.com/stianeikeland/go-rpio"
)

type PinCtrl interface {
	InitializePin()
	TurnHeaterOn()
	TurnHeaterOff()
	TearDown()
}

type RaspberryPiPinCtrl struct {
	pin      rpio.Pin
	HeaterOn bool
}

func (rppc *RaspberryPiPinCtrl) InitializePin() {
	err := rpio.Open()
	if err != nil {
		log.Println("Not a raspberry pi?")
		log.Fatal(err)
	}
	rppc.pin = rpio.Pin(18)
	rppc.pin.Output()
	rppc.pin.Low()
	log.Println("Raspberry Pi Pin Initialized")
}

func (rppc *RaspberryPiPinCtrl) TurnHeaterOn() {
	rppc.pin.High()
	rppc.HeaterOn = true
}

func (rppc *RaspberryPiPinCtrl) TurnHeaterOff() {
	rppc.pin.Low()
	rppc.HeaterOn = false
}

func (rppc *RaspberryPiPinCtrl) TearDown() {
	rppc.TurnHeaterOff()
	rpio.Close()
	fmt.Println("shutting down")
}

type MockPinCtrl struct {
	HeaterOn bool
}

func (pcm *MockPinCtrl) InitializePin() {
	log.Println("Mock initialize gpio pin")
}
func (pcm *MockPinCtrl) TurnHeaterOn() {
	pcm.HeaterOn = true
	log.Println("Mock turn heater on")
}
func (pcm *MockPinCtrl) TurnHeaterOff() {
	pcm.HeaterOn = true
	log.Println("Mock turn heater off")
}
func (pcm *MockPinCtrl) TearDown() {
	pcm.TurnHeaterOff()
	log.Println("Mock tear down")
}
