package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "gitlab.com/homanchou/heatercontrol"
)

var _ = Describe("Heatercontrol", func() {
	nightTime, _ := time.Parse(
		time.RFC3339,
		"2019-01-12T22:04:32-08:00")
	dayTime, _ := time.Parse(
		time.RFC3339,
		"2019-01-12T11:04:32-08:00")
	myDesiredTemp := 90.0
	// Allow user to control a MaxTemp and a MinTemp.

	// Allow unit to poll for current Temp every 10 seconds

	Context("Allow unit to resume daylight to evening and vice versa", func() {
		It("does not reset temps when staying daytime to daytime", func() {
			newTemp := 72.0
			hs := NewHeaterState(dayTime, newTemp)
			hs.DesiredTemp = DefaultDaytimeTemp
			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.DesiredTemp).To(Equal(DefaultDaytimeTemp))

		})
		It("resets temps when transition night to day", func() {
			newTemp := 72.0
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = DefaultNighttimeTemp

			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.DesiredTemp).To(Equal(DefaultDaytimeTemp))

		})

	})
	Context("Allow unit to return maxTemp and minTemp back to regularly scheduled defaults 1 hour after user adjusted", func() {
		It("keeps min and max if timer not run out even if transitioning from night to day", func() {

			timerEnd := dayTime.Add(1 * time.Minute)
			newTemp := 72.0
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = myDesiredTemp
			hs.CustomRangeTimeLimit = &timerEnd

			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.DesiredTemp).To(Equal(myDesiredTemp))

		})
		It("goes back to day if timer ends in the day", func() {

			timerEnd := dayTime.Add(-1 * time.Minute)
			newTemp := 72.0
			hs := NewHeaterState(dayTime, newTemp)
			hs.DesiredTemp = myDesiredTemp
			hs.CustomRangeTimeLimit = &timerEnd

			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.DesiredTemp).To(Equal(DefaultDaytimeTemp))
		})
		It("goes back to night if timer ends in the night", func() {

			timerEnd := nightTime.Add(-1 * time.Minute)
			newTemp := 72.0
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = myDesiredTemp
			hs.CustomRangeTimeLimit = &timerEnd

			hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(hs.DesiredTemp).To(Equal(DefaultNighttimeTemp))
		})
	})

	Context("Allow unit to turn heater On or Off based on Temp", func() {
		It("returns heater off command if current temp hotter than desired", func() {

			newTemp := 100.0
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = myDesiredTemp
			command := hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(command).To(Equal(Off))
		})
		It("returns heater on command if current temp too cold", func() {
			newTemp := 30.0
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = myDesiredTemp

			command := hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(command).To(Equal(On))
		})
		It("does nothing if in range", func() {

			newTemp := myDesiredTemp
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = myDesiredTemp

			newTemp = myDesiredTemp + 0.1
			command := hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(command).To(Equal(NoAction))
			newTemp = myDesiredTemp - 0.1
			command = hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(command).To(Equal(NoAction))
		})
	})
	Context("Allow user to control Disable and Enable.", func() {
		It("disables heater", func() {
			newTemp := 10.0
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = myDesiredTemp
			hs.Disable()
			Expect(hs.Enabled).To(Equal(false))

			command := hs.RefreshTimeAndTemp(nightTime, 10.0)
			Expect(command).To(Equal(Off))
		})
		It("enables heater", func() {
			newTemp := 30.0
			hs := NewHeaterState(nightTime, newTemp)
			hs.DesiredTemp = myDesiredTemp
			hs.Enable()
			Expect(hs.Enabled).To(Equal(true))
		})
	})

})
