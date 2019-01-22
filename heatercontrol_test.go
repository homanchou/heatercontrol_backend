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
	myMaxTemp := 99.9
	myMinTemp := 45.0
	// Allow user to control a MaxTemp and a MinTemp.

	// Allow unit to poll for current Temp every 10 seconds

	Context("Allow unit to transition max and min temp when going from daylight to evening and vice versa", func() {
		It("does not reset temps when staying daytime to daytime", func() {
			hs := HeaterState{
				MaxTemp:       DefaultDaytimeMaxTemp,
				MinTemp:       DefaultDaytimeMinTemp,
				LastUpdatedAt: dayTime,
			}
			newTemp := 72.0
			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.MaxTemp).To(Equal(DefaultDaytimeMaxTemp))
			Expect(hs.MinTemp).To(Equal(DefaultDaytimeMinTemp))

		})
		It("resets temps when transition night to day", func() {
			hs := HeaterState{
				MaxTemp:       DefaultNighttimeMaxTemp,
				MinTemp:       DefaultNighttimeMinTemp,
				LastUpdatedAt: nightTime,
			}
			newTemp := 72.0

			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.MaxTemp).To(Equal(DefaultDaytimeMaxTemp))
			Expect(hs.MinTemp).To(Equal(DefaultDaytimeMinTemp))

		})

	})
	Context("Allow unit to return maxTemp and minTemp back to regularly scheduled defaults 1 hour after user adjusted", func() {
		It("keeps min and max if timer not run out even if transitioning from night to day", func() {

			timerEnd := dayTime.Add(1 * time.Minute)
			hs := HeaterState{
				MaxTemp:              myMaxTemp,
				MinTemp:              myMinTemp,
				LastUpdatedAt:        nightTime,
				CustomRangeTimeLimit: &timerEnd,
			}
			newTemp := 72.0
			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.MaxTemp).To(Equal(myMaxTemp))
			Expect(hs.MinTemp).To(Equal(myMinTemp))

		})
		It("goes back to day if timer ends in the day", func() {

			timerEnd := dayTime.Add(-1 * time.Minute)
			hs := HeaterState{
				MaxTemp:              myMaxTemp,
				MinTemp:              myMinTemp,
				LastUpdatedAt:        dayTime,
				CustomRangeTimeLimit: &timerEnd,
			}
			newTemp := 72.0

			hs.RefreshTimeAndTemp(dayTime, newTemp)
			Expect(hs.MaxTemp).To(Equal(DefaultDaytimeMaxTemp))
			Expect(hs.MinTemp).To(Equal(DefaultDaytimeMinTemp))
		})
		It("goes back to night if timer ends in the night", func() {

			timerEnd := nightTime.Add(-1 * time.Minute)
			hs := HeaterState{
				MaxTemp:              myMaxTemp,
				MinTemp:              myMinTemp,
				LastUpdatedAt:        nightTime,
				CustomRangeTimeLimit: &timerEnd,
			}
			newTemp := 72.0

			hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(hs.MaxTemp).To(Equal(DefaultNighttimeMaxTemp))
			Expect(hs.MinTemp).To(Equal(DefaultNighttimeMinTemp))
		})
	})

	Context("Allow unit to turn heater On or Off based on Temp", func() {
		It("returns heater off command if current temp hotter than max", func() {

			hs := HeaterState{
				MaxTemp:       myMaxTemp,
				MinTemp:       myMinTemp,
				LastUpdatedAt: nightTime,
			}
			newTemp := 100.0
			command := hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(command).To(Equal(Off))
		})
		It("returns heater on command if current temp too cold", func() {

			hs := HeaterState{
				MaxTemp:       myMaxTemp,
				MinTemp:       myMinTemp,
				LastUpdatedAt: nightTime,
			}
			newTemp := 30.0
			command := hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(command).To(Equal(On))
		})
		It("does nothing if in range", func() {

			hs := HeaterState{
				MaxTemp:       myMaxTemp,
				MinTemp:       myMinTemp,
				LastUpdatedAt: nightTime,
			}
			newTemp := 50.0
			command := hs.RefreshTimeAndTemp(nightTime, newTemp)
			Expect(command).To(Equal(NoAction))
		})
	})
	Context("Allow user to control Disable and Enable.", func() {
		It("disables heater", func() {
			hs := HeaterState{
				MaxTemp:       myMaxTemp,
				MinTemp:       myMinTemp,
				LastUpdatedAt: nightTime,
			}
			hs.Disable()
			Expect(hs.Disabled).To(Equal(true))

			command := hs.RefreshTimeAndTemp(nightTime, myMinTemp-0.5)
			Expect(command).To(Equal(Off))
		})
		It("enables heater", func() {
			hs := HeaterState{
				MaxTemp:       myMaxTemp,
				MinTemp:       myMinTemp,
				LastUpdatedAt: nightTime,
			}
			hs.Enable()
			Expect(hs.Disabled).To(Equal(false))
		})
	})

})
