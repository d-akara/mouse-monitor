package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dakaraphi/mouse-monitor/winapi"
	"github.com/fatih/color"
)

const messagesPerFrequencyCalculation = 100

type MouseData struct {
	xPos           int32
	yPos           int32
	msgCount       int64
	msgHz          float64
	msgHzLastTime  time.Time
	msgHzLastCount int64
}

var displaySignal chan bool
var mouseTracking MouseData

func handleMouseInput(mouseInput winapi.RAWMOUSE) {
	mouseTracking.xPos += mouseInput.LastX
	mouseTracking.yPos += mouseInput.LastY
	mouseTracking.msgCount++

	if mouseInput.ButtonData == 1 {
		displaySignal <- true // send display notification
		<-displaySignal       // wait for display complete
		mouseTracking = MouseData{}
	} else if mouseTracking.msgCount > mouseTracking.msgHzLastCount+messagesPerFrequencyCalculation {
		currentTime := time.Now()

		currentHz := float64(1) / (currentTime.Sub(mouseTracking.msgHzLastTime).Seconds() / messagesPerFrequencyCalculation)
		if currentHz > mouseTracking.msgHz {
			mouseTracking.msgHz = currentHz
		}
		mouseTracking.msgHzLastTime = currentTime
		mouseTracking.msgHzLastCount = mouseTracking.msgCount
	}
}

func consoleDisplayLoop() {
	valuesToColorStrings := func() (string, string, string) {
		x := color.GreenString(strconv.FormatInt(int64(mouseTracking.xPos), 10))
		y := color.GreenString(strconv.FormatInt(int64(mouseTracking.yPos), 10))
		hz := color.CyanString(strconv.FormatFloat(mouseTracking.msgHz, 'f', 2, 64))
		return x, y, hz
	}
	for {
		select {
		case <-time.After(time.Millisecond * 50):
			x, y, hz := valuesToColorStrings()
			fmt.Fprintf(color.Output, "current position: x %v  y %v   max update Hz %v    \r", x, y, hz)
		case <-displaySignal:
			x, y, hz := valuesToColorStrings()
			fmt.Fprintf(color.Output, "position delta: x %v  y %v   max update Hz %v          \r\n", x, y, hz)
			displaySignal <- true
		}
	}
}

func main() {
	displaySignal = make(chan bool)
	mouseTracking.msgHzLastTime = time.Now()
	go winapi.StartWindowsMessageLoop(winapi.MakeMouseRawInputReceiver(handleMouseInput))
	consoleDisplayLoop()

	return
}
