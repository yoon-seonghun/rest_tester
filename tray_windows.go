//go:build windows

package main

// Windows 알림 영역(트레이) 아이콘 — 외부 라이브러리 없이 Win32 API 직접 호출

import (
	_ "embed"
	"runtime"
	"syscall"
	"unsafe"
)

//go:embed icon.bin
var iconBits []byte

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pRegisterClassExW         = user32.NewProc("RegisterClassExW")
	pCreateWindowExW          = user32.NewProc("CreateWindowExW")
	pDefWindowProcW           = user32.NewProc("DefWindowProcW")
	pGetMessageW              = user32.NewProc("GetMessageW")
	pTranslateMessage         = user32.NewProc("TranslateMessage")
	pDispatchMessageW         = user32.NewProc("DispatchMessageW")
	pCreatePopupMenu          = user32.NewProc("CreatePopupMenu")
	pAppendMenuW              = user32.NewProc("AppendMenuW")
	pTrackPopupMenu           = user32.NewProc("TrackPopupMenu")
	pDestroyMenu              = user32.NewProc("DestroyMenu")
	pSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	pPostMessageW             = user32.NewProc("PostMessageW")
	pGetCursorPos             = user32.NewProc("GetCursorPos")
	pLoadIconW                = user32.NewProc("LoadIconW")
	pCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	pMessageBoxW              = user32.NewProc("MessageBoxW")
	pRegisterWindowMessageW   = user32.NewProc("RegisterWindowMessageW")
	pShellNotifyIconW         = shell32.NewProc("Shell_NotifyIconW")
	pGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmCommand   = 0x0111
	wmNull      = 0x0000
	wmApp       = 0x8000
	wmTray      = wmApp + 1
	wmLButtonUp = 0x0202
	wmRButtonUp = 0x0205
	wmLDblClick = 0x0203

	nimAdd     = 0
	nimModify  = 1
	nimDelete  = 2
	nifMessage = 0x1
	nifIcon    = 0x2
	nifTip     = 0x4
	nifInfo    = 0x10
	niifInfo   = 0x1

	mfString       = 0x0
	mfSeparator    = 0x800
	mfDefault      = 0x1000
	tpmReturnCmd   = 0x100
	tpmNoNotify    = 0x80
	tpmRightButton = 0x2

	menuOpen = 1
	menuQuit = 2
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type point struct{ x, y int32 }

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
	private uint32
}

var (
	trayNID    notifyIconData
	trayHwnd   uintptr
	trayOnOpen func()
	trayOnQuit func()
	taskbarMsg uintptr
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func copyU16(dst []uint16, s string) {
	src, _ := syscall.UTF16FromString(s)
	if len(src) > len(dst) {
		src = src[:len(dst)]
		src[len(src)-1] = 0
	}
	copy(dst, src)
}

func showMessage(title, text string) {
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), 0x40) // MB_ICONINFORMATION
}

func showError(title, text string) {
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), 0x10) // MB_ICONERROR
}

func trayBalloon(title, text string) {
	if trayHwnd == 0 {
		return
	}
	n := trayNID
	n.uFlags = nifInfo
	copyU16(n.szInfoTitle[:], title)
	copyU16(n.szInfo[:], text)
	n.dwInfoFlags = niifInfo
	pShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&n)))
}

func trayRemove() {
	if trayHwnd != 0 {
		pShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&trayNID)))
	}
}

func showMenu(hwnd uintptr) {
	menu, _, _ := pCreatePopupMenu.Call()
	pAppendMenuW.Call(menu, mfString|mfDefault, menuOpen, uintptr(unsafe.Pointer(u16("화면 열기"))))
	pAppendMenuW.Call(menu, mfSeparator, 0, 0)
	pAppendMenuW.Call(menu, mfString, menuQuit, uintptr(unsafe.Pointer(u16("종료"))))
	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(hwnd) // 메뉴 밖을 누르면 닫히도록
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmReturnCmd|tpmNoNotify|tpmRightButton, uintptr(pt.x), uintptr(pt.y), 0, hwnd, 0)
	pPostMessageW.Call(hwnd, wmNull, 0, 0)
	pDestroyMenu.Call(menu)
	switch cmd {
	case menuOpen:
		trayOnOpen()
	case menuQuit:
		trayOnQuit()
	}
}

func wndProc(hwnd, umsg, wparam, lparam uintptr) uintptr {
	switch {
	case umsg == wmTray:
		switch lparam & 0xFFFF {
		case wmLButtonUp, wmLDblClick:
			go trayOnOpen()
		case wmRButtonUp:
			showMenu(hwnd)
		}
		return 0
	case taskbarMsg != 0 && umsg == taskbarMsg:
		// 탐색기가 재시작되면 아이콘을 다시 등록
		pShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&trayNID)))
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, umsg, wparam, lparam)
	return r
}

// runTray: 트레이 아이콘을 띄우고 메시지 루프를 돈다 (반환하지 않음)
func runTray(tip string, onOpen, onQuit func(), onReady func()) {
	runtime.LockOSThread()
	trayOnOpen, trayOnQuit = onOpen, onQuit

	hInst, _, _ := pGetModuleHandleW.Call(0)
	cls := u16("RestTesterTrayWnd")
	wc := wndClassEx{lpfnWndProc: syscall.NewCallback(wndProc), hInstance: hInst, lpszClassName: cls}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	// HWND_MESSAGE(-3): 화면에 보이지 않는 메시지 전용 창 — 단, 탐색기 재시작 알림을 받으려면 일반 숨김 창 사용
	trayHwnd, _, _ = pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(u16("RestTester"))),
		0, 0, 0, 0, 0, 0, 0, hInst, 0)
	taskbarMsg, _, _ = pRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(u16("TaskbarCreated"))))

	icon, _, _ := pCreateIconFromResourceEx.Call(uintptr(unsafe.Pointer(&iconBits[0])), uintptr(len(iconBits)), 1, 0x00030000, 32, 32, 0)
	if icon == 0 {
		icon, _, _ = pLoadIconW.Call(0, 32512) // IDI_APPLICATION
	}
	trayNID.cbSize = uint32(unsafe.Sizeof(trayNID))
	trayNID.hWnd = trayHwnd
	trayNID.uID = 1
	trayNID.uFlags = nifMessage | nifIcon | nifTip
	trayNID.uCallbackMessage = wmTray
	trayNID.hIcon = icon
	copyU16(trayNID.szTip[:], tip)
	pShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&trayNID)))
	if onReady != nil {
		onReady()
	}

	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	trayRemove()
	onQuit()
}
