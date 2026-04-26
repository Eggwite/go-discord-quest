//go:build windows

package main

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	CW_USEDEFAULT       = 0x80000000
	WM_DESTROY          = 0x0002
	WM_CLOSE            = 0x0010
	WM_QUIT             = 0x0012
	WM_PAINT            = 0x000F
	SRCCOPY             = 0x00CC0020
)

var (
	user32                     = windows.NewLazySystemDLL("user32.dll")
	kernel32                   = windows.NewLazySystemDLL("kernel32.dll")
	gdi32                      = windows.NewLazySystemDLL("gdi32.dll")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procGetModuleHandleW       = kernel32.NewProc("GetModuleHandleW")
	procBeginPaint             = user32.NewProc("BeginPaint")
	procEndPaint               = user32.NewProc("EndPaint")
	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procFillRect               = user32.NewProc("FillRect")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procDrawTextW              = user32.NewProc("DrawTextW")
	procCreateFontW            = gdi32.NewProc("CreateFontW")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procGetClientRect          = user32.NewProc("GetClientRect")
	procGetMessageW            = user32.NewProc("GetMessageW")
)

type msg struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type rect struct{ Left, Top, Right, Bottom int32 }

type paintStruct struct {
	Hdc         windows.Handle
	FErase      int32
	RcPaint     rect
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

func main() {
	title := "System Stub"
	if len(os.Args) > 2 && os.Args[1] == "--title" {
		title = os.Args[2]
	}
	runWindow(title)
}

func runWindow(title string) {
	instance, _, _ := procGetModuleHandleW.Call(0)
	className := windows.StringToUTF16Ptr("StaticAestheticClass")

	wc := wndClassEx{
		Size:       uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:    windows.NewCallback(wndProc),
		Instance:   windows.Handle(instance),
		Background: windows.Handle(5),
		ClassName:  className,
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(title))),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		CW_USEDEFAULT, CW_USEDEFAULT, 420, 180, // Slightly shorter window
		0, 0, instance, 0,
	)

	var message msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) <= 0 {
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func wndProc(h windows.Handle, msgID uint32, wParam, lParam uintptr) uintptr {
	switch msgID {
	case 0x0014: // WM_ERASEBKGND
		return 1
	case WM_PAINT:
		drawStaticBufferedContent(h)
		return 0
	case WM_CLOSE, WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(uintptr(h), uintptr(msgID), wParam, lParam)
		return ret
	}
}

func drawStaticBufferedContent(hwnd windows.Handle) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))

	var r rect
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	width, height := r.Right-r.Left, r.Bottom-r.Top

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	memBitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	oldBitmap, _, _ := procSelectObject.Call(memDC, memBitmap)

	defer func() {
		procSelectObject.Call(memDC, oldBitmap)
		procDeleteObject.Call(memBitmap)
		procDeleteDC.Call(memDC)
	}()

	hMemDC := windows.Handle(memDC)

	// Background: Deep Midnight
	bgBrush, _, _ := procCreateSolidBrush.Call(0x000F0A08)
	procFillRect.Call(uintptr(hMemDC), uintptr(unsafe.Pointer(&r)), bgBrush)
	procDeleteObject.Call(bgBrush)

	// UI Text
	procSetBkMode.Call(uintptr(hMemDC), 1)

	// Main Title
	font, _, _ := procCreateFontW.Call(22, 0, 0, 0, 800, 0, 0, 0, 0, 0, 0, 5, 0, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Segoe UI"))))
	oldFont, _, _ := procSelectObject.Call(uintptr(hMemDC), font)

	procSetTextColor.Call(uintptr(hMemDC), 0x00FFFF00) // Cyber Cyan
	titleRect := rect{25, 30, r.Right, r.Bottom}
	str1 := windows.StringToUTF16Ptr("STUB_MODULE :: CLOAK_ENGAGED")
	procDrawTextW.Call(uintptr(hMemDC), uintptr(unsafe.Pointer(str1)), ^uintptr(0), uintptr(unsafe.Pointer(&titleRect)), 0)
	procDeleteObject.Call(font)

	// Status Line
	fontSmall, _, _ := procCreateFontW.Call(14, 0, 0, 0, 400, 0, 0, 0, 0, 0, 0, 5, 0, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Consolas"))))
	procSelectObject.Call(uintptr(hMemDC), fontSmall)
	procSetTextColor.Call(uintptr(hMemDC), 0x0057F287) // Success Green

	subRect := rect{25, 65, r.Right, r.Bottom}
	str2 := windows.StringToUTF16Ptr("> INITIALISING_OPTIMISED_STANDBY... [OK]\n> ENCRYPTING_PROCESS_ID... [OK]")
	procDrawTextW.Call(uintptr(hMemDC), uintptr(unsafe.Pointer(str2)), ^uintptr(0), uintptr(unsafe.Pointer(&subRect)), 0)

	procSelectObject.Call(uintptr(hMemDC), oldFont)
	procDeleteObject.Call(fontSmall)

	// Copy to screen
	procBitBlt.Call(hdc, 0, 0, uintptr(width), uintptr(height), memDC, 0, 0, SRCCOPY)
}
