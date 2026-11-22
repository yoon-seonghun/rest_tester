//go:build !windows

package main

// macOS / Linux: 트레이 아이콘 없이 동작.
// 터미널에서 실행했다면 안내 문구가 보이고, 종료는 화면 상단 [종료] 버튼 또는 Ctrl+C.

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func showMessage(title, text string) { fmt.Println("[" + title + "] " + text) }
func showError(title, text string)   { fmt.Fprintln(os.Stderr, "["+title+"] "+text) }
func trayBalloon(title, text string) {}
func trayRemove()                    {}

func runTray(tip string, onOpen, onQuit func(), onReady func()) {
	fmt.Println("==============================================")
	fmt.Println("  REST API 테스트 도구 실행 중")
	fmt.Println("  화면 주소:", serverAddr)
	fmt.Println("  저장 파일:", dataFile())
	fmt.Println()
	fmt.Println("  브라우저가 자동으로 열리지 않으면 위 주소를 직접 여세요.")
	fmt.Println("  종료: 화면 상단 [종료] 버튼 또는 이 창에서 Ctrl+C")
	fmt.Println("==============================================")
	if onReady != nil {
		onReady()
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	onQuit()
}
