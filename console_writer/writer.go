package console_writer

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	sg "github.ibm.com/ckwaldon/swiftlygo/slo"
)

const clearLine string = "\033[2K"
const upLine string = "\033[1A"

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
				fmt.Print("\r%s", upLine)
			}
			return
		default:
			out := fmt.Sprintf("\r%s%s%s", clearLine, loading[count], c.currentStage)

			if c.showStatus {
				out = getStats(c.status, out, first)
				first = false
			}

			fmt.Print(out)
			count = (count + 1) % len(loading)
		}
		time.Sleep(speed * time.Millisecond)
	}
}

func getStats(status *sg.Status, out string, first bool) string {
	progress := [11]string{">         ", "=>        ", "==>       ", "===>      ", "====>     ",
		"=====>    ", "======>   ", "=======>  ", "========> ", "=========>", "=========="}

	percent := status.PercentComplete()
	percentStr := fmt.Sprintf("%.0f%%", percent)
	if percent == 100 {
		percentStr += " Uploading manifest"
	}

	stats := fmt.Sprintf("\nSpeed %s |%s| %s", Cyan(fmt.Sprintf("%.2f MB/s", status.RateMBPS())), progress[int(percent/10.0)], percentStr)

	if first {
		stats = out + stats
	} else {
		stats = "\r" + clearLine + upLine + out + stats
	}

	return stats
}
