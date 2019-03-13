package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"./winapi"

	"github.com/lxn/win"
)

type MouseData struct {
	xPos         int32
	yPos         int32
	msgCount     int32
	previousTime time.Time
}

var mouseTracking MouseData

func WndProc(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_CREATE:
		fmt.Println("Registering raw input mouse")
		devices := getRawInputDevices(hWnd)
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
		mouseTracking.xPos += raw.Mouse.LastX
		mouseTracking.yPos += raw.Mouse.LastY
		if raw.Mouse.ButtonData == 1 {
			fmt.Printf("logged current position: x %v  y %v \r\n", mouseTracking.xPos, mouseTracking.yPos)
			mouseTracking.xPos = 0
			mouseTracking.yPos = 0
			mouseTracking.msgCount = 0
		} else {
			mouseTracking.msgCount++
			currentTime := time.Now()
			diffTime := float64(1) / currentTime.Sub(mouseTracking.previousTime).Seconds()
			fmt.Printf("mouse moved: x %v  y %v   msg %1.f    \r", mouseTracking.xPos, mouseTracking.yPos, diffTime)
			mouseTracking.previousTime = currentTime
		}

	case win.WM_DESTROY:
		fmt.Println("destroying window")
		win.PostQuitMessage(0)
	default:
		return win.DefWindowProc(hWnd, msg, wParam, lParam)
	}

	return 0
}

func getRawInputDevices(hWnd win.HWND) []win.RAWINPUTDEVICE {
	devices := make([]win.RAWINPUTDEVICE, 1)
	devices[0].UsUsagePage = 0x01
	devices[0].UsUsage = 0x02
	devices[0].DwFlags = win.RIDEV_INPUTSINK
	devices[0].HwndTarget = hWnd
	return devices
}

/* create window so we can receive messages from message loop */
func createMessageLoop() int {
	hInstance := win.GetModuleHandle(syscall.StringToUTF16Ptr(""))
	lpszClassName := syscall.StringToUTF16Ptr("WNDclass")
	var wcex win.WNDCLASSEX
	wcex.HInstance = hInstance
	wcex.LpszClassName = lpszClassName
	wcex.LpfnWndProc = syscall.NewCallback(WndProc)
	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	win.RegisterClassEx(&wcex)
	win.CreateWindowEx(
		0, lpszClassName, syscall.StringToUTF16Ptr("Message Receiver Window"),
		win.WS_OVERLAPPEDWINDOW,
		win.CW_USEDEFAULT, win.CW_USEDEFAULT, 400, 400, 0, 0, hInstance, nil)
	var msg win.MSG
	for {
		if win.GetMessage(&msg, 0, 0, 0) == 0 {
			break
		}
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
	return int(msg.WParam)
}

func main() {
	mouseTracking.previousTime = time.Now()
	createMessageLoop()
	return
}
