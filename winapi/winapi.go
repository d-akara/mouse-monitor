package winapi

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
