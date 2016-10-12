package console_writer

import (
	"fmt"
	"runtime"
	"time"

	"github.com/fatih/color"
	sg "github.ibm.com/ckwaldon/swiftlygo/slo"
)

// speed is the time (in milliseconds) between console writes
const speed time.Duration = 200

// ANSI escape codes for printing to terminal
var ClearLine string = "\033[2K"
var upLine string = "\033[1A"

// color output wrappers
var Cyan (func(string, ...interface{}) string) = color.New(color.FgCyan, color.Bold).SprintfFunc()
var White (func(string, ...interface{}) string) = color.New(color.FgHiWhite, color.Bold).SprintfFunc()
var Green (func(string, ...interface{}) string) = color.New(color.FgGreen, color.Bold).SprintfFunc()
var Red (func(string, ...interface{}) string) = color.New(color.FgRed, color.Bold).SprintfFunc()

// ConsoleWriter asynchronously prints the current state to the console
type ConsoleWriter struct {
	quit         chan int
	currentStage chan string
	status       *sg.Status
	Write        func()
}

// NewConsoleWriter creates a new ConsoleWriter
func NewConsoleWriter() *ConsoleWriter {
	newWriter := &ConsoleWriter{
		quit:         make(chan int),
		currentStage: make(chan string),
		status:       nil,
	}

	// Disable color and escape sequences on unsupported systems
	if runtime.GOOS == "windows" {
		newWriter.Write = newWriter.writeWithoutANSI
		color.NoColor = true
		ClearLine = ""
		upLine = ""
	} else {
		newWriter.Write = newWriter.writeWithANSI
	}

	return newWriter
}

// Quit sends a kill signal to this ConsoleWriter
func (c *ConsoleWriter) Quit() {
	c.quit <- 0
}

// SetCurrentStage sets the current state
func (c *ConsoleWriter) SetCurrentStage(currentStage string) {
	c.currentStage <- currentStage
}

// SetStatus gives the writer the uploader's status, if available
func (c *ConsoleWriter) SetStatus(status *sg.Status) {
	c.status = status
}

// writeWithANSI prints output with ANSI support
func (c *ConsoleWriter) writeWithANSI() {
	loading := [6]string{" *    ", "  *   ", "   *  ", "    * ", "   *  ", "  *   "}
	count := 0
	first := true
	cur := ""

	writeHelper := func() {
		out := fmt.Sprintf("\r%s%s%s", ClearLine, loading[count], cur)

		if c.status != nil {
			out = getStats(c.status, out, first)
			first = false
		}

		fmt.Print(out)
		count = (count + 1) % len(loading)
	}

	for {
		select {
		case <-c.quit:
			if c.status != nil {
				fmt.Print("\r%s", upLine)
			}
			return
		case cur = <-c.currentStage:
			writeHelper()
		default:
			writeHelper()
		}
		time.Sleep(speed * time.Millisecond)
	}
}

// writeWithoutANSI prints output without ANSI support
func (c *ConsoleWriter) writeWithoutANSI() {
	for {
		select {
		case <-c.quit:
			return
		case stage := <-c.currentStage:
			fmt.Println(stage)
		}
	}
}

// getStats prints a progress bar for the SLO upload
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
		stats = "\r" + ClearLine + upLine + out + stats
	}

	return stats
}
