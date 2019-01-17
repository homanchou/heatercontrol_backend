package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	app := NewMockApp()
	app.Initialize()

	go func() {
		<-sigs
		app.TearDown()
		os.Exit(0)
	}()

	app.StartWebServer()

}
