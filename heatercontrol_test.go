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

	coldTemp := 69.0
	warmTemp := 75.0
	var hs *HeaterState
	BeforeEach(func() {
		hs = NewHeaterState()
	})
	Context("night time", func() {
		It("should turn heater on if cold", func() {
			Expect(hs.NextAction(nightTime, coldTemp)).To(Equal(On))
		})
		It("should turn heater off if warm", func() {
			Expect(hs.NextAction(nightTime, warmTemp)).To(Equal(Off))
		})
	})
	Context("day time", func() {
		It("should keep heater off even if cold to save resources", func() {
			Expect(hs.NextAction(dayTime, coldTemp)).To(Equal(Off))
		})
		It("should keep heater off if warm", func() {
			Expect(hs.NextAction(dayTime, warmTemp)).To(Equal(Off))
		})
		It("should allow heater on if econo turned off and it's cold", func() {
			hs.EconoMode = false
			// hs.TurnOffEconoMode()
			Expect(hs.NextAction(dayTime, coldTemp)).To(Equal(On))
		})
	})
	Context("Econo Mode Resumes", func() {
		It("doesn't resume during the day", func() {
			hs.EconoMode = false
			hs.NextAction(dayTime, warmTemp)
			Expect(hs.EconoMode).To(Equal(false))
		})
		It("resumes at night", func() {
			hs.EconoMode = false
			hs.NextAction(nightTime, warmTemp)
			Expect(hs.EconoMode).To(Equal(true))
			hs.EconoMode = false
			hs.NextAction(nightTime, coldTemp)
			Expect(hs.EconoMode).To(Equal(true))

		})
	})
	Context("force heater on", func() {
		It("Keeps heater on even during the day and it's not cold", func() {
			oneHourFromNow := dayTime.Add(time.Hour * 1)
			hs.ForcedOnTimeLimit = &oneHourFromNow
			Expect(hs.NextAction(dayTime, warmTemp)).To(Equal(On))
		})
		It("Resumes regular behavior after time is up", func() {
			oneHourAgo := dayTime.Add(time.Hour * -1)
			hs.ForcedOnTimeLimit = &oneHourAgo
			Expect(hs.NextAction(dayTime, warmTemp)).To(Equal(Off))
		})

	})
	Context("heater forced off", func() {
		It("always gives back heater off", func() {
			hs.ForcedOff = true
			Expect(hs.NextAction(nightTime, coldTemp)).To(Equal(Off))
		})
	})

})
