package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHeatercontrol(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Heatercontrol Suite")
}
