package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"
	"unsafe"

	"github.com/gonutz/ide/w32"
)

func main() {
	defer handlePanics()

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

func handlePanics() {
	// After a panic the user/developer is shown the stack trace. To be sure
	// that the message is seen, it is not only printed to stdout but also saved
	// to disk and a message box pops up.
	if err := recover(); err != nil {
		message := fmt.Sprintf("panic: %v\nstack:\n\n%s\n", err, debug.Stack())

		// print to standard output
		fmt.Println(message)

		// write as a log file to disk
		logFile := filepath.Join(
			os.Getenv("APPDATA"),
			"ide_log_"+time.Now().Format("2006_01_02__15_04_05")+".txt",
		)
		ioutil.WriteFile(logFile, []byte(message), 0666)

		// open crash log file with the default text viewer
		exec.Command("cmd", "/C", logFile).Start()

		// pop up a message box
		w32.MessageBox(
			0,
			message,
			"The program crashed",
			w32.MB_OK|w32.MB_ICONERROR|w32.MB_TOPMOST,
		)
	}
}
