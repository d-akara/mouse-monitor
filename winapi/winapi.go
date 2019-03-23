package winapi

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

type WndProc func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr

// reference - https://docs.microsoft.com/en-us/windows/desktop/api/winuser/ns-winuser-tagrawmouse
type RAWINPUTHEADER struct {
	Type   uint32
	Size   uint32
	Device uintptr
	Param  uintptr
}

// reference - https://docs.microsoft.com/en-us/windows/desktop/api/winuser/ns-winuser-tagrawinput
type RAWINPUT struct {
	Header RAWINPUTHEADER
	Mouse  RAWMOUSE
}

// reference - https://docs.microsoft.com/en-us/windows/desktop/api/winuser/ns-winuser-tagrawmouse
type RAWMOUSE struct {
	Flags            uint16
	ButtonFlags      uint16
	ButtonData       uint16
	RawButtons       uint32
	LastX            int32
	LastY            int32
	ExtraInformation uint32
}

/* create window so we can receive messages from message loop */
func StartWindowsMessageLoop(windowsMessageReceiver WndProc) int {
	hInstance := win.GetModuleHandle(syscall.StringToUTF16Ptr(""))
	lpszClassName := syscall.StringToUTF16Ptr("WNDclass")
	var wcex win.WNDCLASSEX
	wcex.HInstance = hInstance
	wcex.LpszClassName = lpszClassName
	wcex.LpfnWndProc = syscall.NewCallback(windowsMessageReceiver)
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

func GetRawInputDeviceMouseDefinition(hWnd win.HWND) []win.RAWINPUTDEVICE {
	devices := make([]win.RAWINPUTDEVICE, 1)
	devices[0].UsUsagePage = 0x01
	devices[0].UsUsage = 0x02
	devices[0].DwFlags = win.RIDEV_INPUTSINK
	devices[0].HwndTarget = hWnd
	return devices
}

func MakeMouseRawInputReceiver(mouseInputHandler func(RAWMOUSE)) WndProc {
	return func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case win.WM_CREATE:
			fmt.Println("Registering raw input mouse")
			devices := GetRawInputDeviceMouseDefinition(hWnd)
			len := uint32(len(devices))
			size := uint32(unsafe.Sizeof(devices[0]))
			if !win.RegisterRawInputDevices(&devices[0], len, size) {
				panic("Unable to register devices")
			}

		case win.WM_INPUT:
			// reference - https://docs.microsoft.com/en-us/windows/desktop/DxTechArts/taking-advantage-of-high-dpi-mouse-movement
			var raw RAWINPUT
			cbSize := uint32(unsafe.Sizeof(raw))
			win.GetRawInputData((win.HRAWINPUT)(unsafe.Pointer(uintptr(lParam))), win.RID_INPUT, unsafe.Pointer(&raw), &cbSize, uint32(unsafe.Sizeof(RAWINPUTHEADER{})))
			mouseInputHandler(raw.Mouse)

		case win.WM_DESTROY:
			fmt.Println("destroying window")
			win.PostQuitMessage(0)
		default:
			return win.DefWindowProc(hWnd, msg, wParam, lParam)
		}

		return 0
	}
}
