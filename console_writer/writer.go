package console_writer

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	sg "github.ibm.com/ckwaldon/swiftlygo/slo"
)

// speed is the time (in milliseconds) between console writes
const speed time.Duration = 200

// color output wrappers
var Cyan (func(string, ...interface{}) string) = color.New(color.FgCyan, color.Bold).SprintfFunc()
var White (func(string, ...interface{}) string) = color.New(color.FgHiWhite, color.Bold).SprintfFunc()
var Green (func(string, ...interface{}) string) = color.New(color.FgGreen, color.Bold).SprintfFunc()
var Red (func(string, ...interface{}) string) = color.New(color.FgRed, color.Bold).SprintfFunc()

// ConsoleWriter asynchronously prints the current state to the console
type ConsoleWriter struct {
	quit         chan int
	currentStage string
	status       *sg.Status
	showStatus   bool
}

// NewConsoleWriter creates a new ConsoleWriter
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		quit:         make(chan int),
		currentStage: "Getting started",
		status:       nil,
		showStatus:   false,
	}
}

// Quit sends a kill signal to this ConsoleWriter
func (c *ConsoleWriter) Quit() {
	c.quit <- 0
}

// SetCurrentStage sets the current state
func (c *ConsoleWriter) SetCurrentStage(currentStage string) {
	c.currentStage = currentStage
}

// SetStatus gives the writer the uploader's status, if available
func (c *ConsoleWriter) SetStatus(status *sg.Status) {
	c.status = status
}

// ShowStatus tells the writer that an upload status is available
func (c *ConsoleWriter) ShowStatus() {
	c.showStatus = true
}

// Write begins printing output
func (c *ConsoleWriter) Write() {
	loading := [6]string{" *    ", "  *   ", "   *  ", "    * ", "   *  ", "  *   "}
	count := 0
	first := true

	for {
		select {
		case <-c.quit:
			if c.showStatus {
				fmt.Print("\r\033[1A")
			}
			return
		default:
			out := fmt.Sprintf("\r\033[2K%s%s", loading[count], c.currentStage)

			if c.showStatus {
				if first {
					out += fmt.Sprintf("\nSpeed: %f MB/s, %f%% complete", c.status.RateMBPS(), c.status.PercentComplete())
					first = false
				} else {
					out = "\r\033[2K\033[1A" + out + fmt.Sprintf("\nSpeed: %.2f MB/s, %.2f%% complete", c.status.RateMBPS(), c.status.PercentComplete())
				}
			}

			fmt.Print(out)
			count = (count + 1) % len(loading)
		}
		time.Sleep(speed * time.Millisecond)
	}
}
