package main

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/ide/w32"
)

func main() {
	const windowClassName = "GoIDEWindowClass"
	class := w32.WNDCLASSEX{
		WndProc:   syscall.NewCallback(windowMessageHandler),
		Cursor:    w32.LoadCursor(0, w32.IDC_ARROW),
		ClassName: syscall.StringToUTF16Ptr(windowClassName),
	}
	class.Size = uint32(unsafe.Sizeof(class))

	atom := w32.RegisterClassEx(&class)
	if atom == 0 {
		panic("RegisterClassEx failed")
	}

	window := w32.CreateWindowEx(
		0,
		windowClassName,
		"Go IDE",
		w32.WS_OVERLAPPEDWINDOW|w32.WS_VISIBLE,
		10, 10, 850, 800,
		0, 0, 0, 0,
	)
	if window == 0 {
		panic("CreateWindowEx failed")
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) > 0 {
		w32.TranslateMessage(&msg)
		w32.DispatchMessage(&msg)
	}
}

func windowMessageHandler(window, message, w, l uintptr) uintptr {
	switch message {
	case w32.WM_DESTROY:
		w32.PostQuitMessage(0)
		return 1
	default:
		return w32.DefWindowProc(window, message, w, l)
	}
}
