package console_writer

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

const speed time.Duration = 200

var Cyan (func(string, ...interface{}) string) = color.New(color.FgCyan, color.Bold).SprintfFunc()
var White (func(string, ...interface{}) string) = color.New(color.FgWhite, color.Bold).SprintfFunc()
var Green (func(string, ...interface{}) string) = color.New(color.FgGreen, color.Bold).SprintfFunc()
var Red (func(string, ...interface{}) string) = color.New(color.FgRed, color.Bold).SprintfFunc()

type ConsoleWriter struct {
	quit         chan int
	currentStage string
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		quit:         make(chan int),
		currentStage: "Getting started                   ",
	}
}

func (c *ConsoleWriter) Quit() {
	c.quit <- 0
}

func (c *ConsoleWriter) SetCurrentStage(currentStage string) {
	c.currentStage = currentStage
}

func (c *ConsoleWriter) Write() {
	loading := [6]string{" *    ", "  *   ", "   *  ", "    * ", "   *  ", "  *   "}
	count := 0

	for {
		select {
		case <-c.quit:
			return
		default:
			fmt.Printf("\r%s%s", loading[count], c.currentStage)
			count = (count + 1) % len(loading)
		}
		time.Sleep(speed * time.Millisecond)
	}
}
