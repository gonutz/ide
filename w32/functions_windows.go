package w32

import (
	"syscall"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
)

var (
	defWindowProc    = user32.NewProc("DefWindowProcW")
	postQuitMessage  = user32.NewProc("PostQuitMessage")
	loadCursor       = user32.NewProc("LoadCursorW")
	registerClassEx  = user32.NewProc("RegisterClassExW")
	createWindowEx   = user32.NewProc("CreateWindowExW")
	getMessage       = user32.NewProc("GetMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	dispatchMessage  = user32.NewProc("DispatchMessageW")
)

func DefWindowProc(window, msg, wParam, lParam uintptr) uintptr {
	ret, _, _ := defWindowProc.Call(window, msg, wParam, lParam)
	return ret
}

func PostQuitMessage(exitCode int) {
	postQuitMessage.Call(uintptr(exitCode))
}

func LoadCursor(instance uintptr, cursor int) uintptr {
	ret, _, _ := loadCursor.Call(instance, uintptr(cursor))
	return ret
}

func RegisterClassEx(wndClassEx *WNDCLASSEX) uintptr {
	ret, _, _ := registerClassEx.Call(uintptr(unsafe.Pointer(wndClassEx)))
	return ret
}

func CreateWindowEx(
	exStyle uintptr,
	className string,
	windowName string,
	style uintptr,
	x, y, width, height int,
	parent uintptr,
	menu uintptr,
	instance uintptr,
	param uintptr,
) uintptr {
	var classNamePtr *uint16
	if className != "" {
		classNamePtr = syscall.StringToUTF16Ptr(className)
	}

	var windowNamePtr *uint16
	if windowName != "" {
		windowNamePtr = syscall.StringToUTF16Ptr(windowName)
	}

	ret, _, _ := createWindowEx.Call(
		exStyle,
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(windowNamePtr)),
		style,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		parent,
		menu,
		instance,
		param,
	)
	return ret
}

func GetMessage(message *MSG, window, msgFilterMin, msgFilterMax uintptr) int {
	ret, _, _ := getMessage.Call(
		uintptr(unsafe.Pointer(message)),
		window,
		msgFilterMin,
		msgFilterMax,
	)
	return int(ret)
}

func TranslateMessage(message *MSG) bool {
	ret, _, _ := translateMessage.Call(uintptr(unsafe.Pointer(message)))
	return ret != 0
}

func DispatchMessage(message *MSG) uintptr {
	ret, _, _ := dispatchMessage.Call(uintptr(unsafe.Pointer(message)))
	return ret
}
