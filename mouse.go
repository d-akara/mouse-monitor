package main

import (
	"fmt"
	"time"

	"github.com/dakaraphi/mouse-monitor/winapi"
)

type MouseData struct {
	xPos           int32
	yPos           int32
	msgCount       int32
	msgHz          float64
	msgHzLastTime  time.Time
	msgHzLastCount int32
}

var displaySignal chan bool
var mouseTracking MouseData

func handleMouseInput(mouseInput winapi.RAWMOUSE) {
	mouseTracking.xPos += mouseInput.LastX
	mouseTracking.yPos += mouseInput.LastY
	mouseTracking.msgCount++

	if mouseInput.ButtonData == 1 {
		displaySignal <- true
		mouseTracking = MouseData{}
	} else if mouseTracking.msgCount > mouseTracking.msgHzLastCount+100 {
		currentTime := time.Now()

		currentHz := float64(1) / (currentTime.Sub(mouseTracking.msgHzLastTime).Seconds() / 100)
		if currentHz > mouseTracking.msgHz {
			mouseTracking.msgHz = currentHz
		}
		mouseTracking.msgHzLastTime = currentTime
		mouseTracking.msgHzLastCount = mouseTracking.msgCount
	}
}

func displayLoop() {
	for {
		select {
		case <-time.After(time.Millisecond * 50):
			fmt.Printf("current position: x %v  y %v   max update Hz %1.f    \r", mouseTracking.xPos, mouseTracking.yPos, mouseTracking.msgHz)
		case <-displaySignal:
			fmt.Printf("position delta: x %v  y %v   max update Hz %1.f    \r\n", mouseTracking.xPos, mouseTracking.yPos, mouseTracking.msgHz)
		}
	}
}

func main() {
	displaySignal = make(chan bool)
	mouseTracking.msgHzLastTime = time.Now()
	go winapi.StartWindowsMessageLoop(winapi.MakeMouseRawInputReceiver(handleMouseInput))
	displayLoop()

	return
}
