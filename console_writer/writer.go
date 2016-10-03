package console_writer

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

const speed time.Duration = 200

type consoleWriter struct {
	quit         chan int
	currentStage string
}

func NewConsoleWriter() *consoleWriter {
	return &consoleWriter{
		quit:         make(chan int),
		currentStage: "Getting started",
	}
}

func (c *consoleWriter) Quit() {
	c.quit <- 0
}

func (c *consoleWriter) SetCurrentStage(currentStage string) {
	c.currentStage = currentStage
}

func (c *consoleWriter) Write() {
	loading := [6]string{" *    ", "  *   ", "   *  ", "    * ", "   *  ", "  *   "}
	count := 0

	for {
		select {
		case <-c.quit:
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("\r%s                                     \n", green("OK"))
			return
		default:
			fmt.Printf("\r%s%s", loading[count], c.currentStage)
			count = (count + 1) % len(loading)
		}
		time.Sleep(speed * time.Millisecond)
	}
}
