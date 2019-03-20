package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/dakaraphi/mouse-monitor/winapi"

	"github.com/lxn/win"
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
var messages chan string
var mouseTracking MouseData

func handleMouseInput(mouseInput winapi.RAWMOUSE) {
	mouseTracking.xPos += mouseInput.LastX
	mouseTracking.yPos += mouseInput.LastY
	mouseTracking.msgCount++

	if mouseInput.ButtonData == 1 {
		fmt.Printf("logged current position: x %v  y %v \r\n", mouseTracking.xPos, mouseTracking.yPos)
		mouseTracking.xPos = 0
		mouseTracking.yPos = 0
		mouseTracking.msgCount = 0
		mouseTracking.msgHzLastCount = 0
		mouseTracking.msgHz = 0
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

func windowsMessageReceiver(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_CREATE:
		fmt.Println("Registering raw input mouse")
		devices := winapi.GetRawInputDeviceMouseDefinition(hWnd)
		len := uint32(len(devices))
		size := uint32(unsafe.Sizeof(devices[0]))
		if !win.RegisterRawInputDevices(&devices[0], len, size) {
			panic("Unable to register devices")
		}

	case win.WM_INPUT:
		// reference - https://docs.microsoft.com/en-us/windows/desktop/DxTechArts/taking-advantage-of-high-dpi-mouse-movement
		var raw winapi.RAWINPUT
		cbSize := uint32(unsafe.Sizeof(raw))
		win.GetRawInputData((win.HRAWINPUT)(unsafe.Pointer(uintptr(lParam))), win.RID_INPUT, unsafe.Pointer(&raw), &cbSize, uint32(unsafe.Sizeof(winapi.RAWINPUTHEADER{})))
		handleMouseInput(raw.Mouse)

	case win.WM_DESTROY:
		fmt.Println("destroying window")
		win.PostQuitMessage(0)
	default:
		return win.DefWindowProc(hWnd, msg, wParam, lParam)
	}

	return 0
}

func display() {
	for {
		<-time.After(time.Millisecond * 50)
		fmt.Printf("mouse moved: x %v  y %v   max update Hz %1.f    \r", mouseTracking.xPos, mouseTracking.yPos, mouseTracking.msgHz)
	}
}

func main() {
	displaySignal = make(chan bool)
	mouseTracking.msgHzLastTime = time.Now()
	go winapi.StartWindowsMessageLoop(windowsMessageReceiver)
	display()

	return
}
