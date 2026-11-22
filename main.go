package main

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

//go:embed index.html
var indexHTML []byte

const appID = "REST-TESTER-v1"
const preferredPort = 18765

var storeMu sync.Mutex

type sendReq struct {
	Method     string      `json:"method"`
	URL        string      `json:"url"`
	Headers    [][2]string `json:"headers"`
	Body       string      `json:"body"`
	Insecure   bool        `json:"insecure"`
	Timeout    int         `json:"timeout"` // seconds
	KeepCookie bool        `json:"keepCookie"`
}

// 실제로 전송되는 요청(쿠키·자동 헤더 포함)을 원문 그대로 기록하는 전송 계층
type dumpTransport struct {
	base http.RoundTripper
	mu   sync.Mutex
	logs []string
}

func (d *dumpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if b, err := httputil.DumpRequestOut(r, true); err == nil {
		d.mu.Lock()
		d.logs = append(d.logs, string(b))
		d.mu.Unlock()
	}
	return d.base.RoundTrip(r)
}

func (d *dumpTransport) text() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.logs) == 1 {
		return d.logs[0]
	}
	var sb strings.Builder
	for i, l := range d.logs {
		fmt.Fprintf(&sb, "########## 요청 %d/%d %s\n", i+1, len(d.logs), map[bool]string{true: "(리다이렉트로 다시 보냄)", false: ""}[i > 0])
		sb.WriteString(l)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// 프로그램 실행 중 유지되는 쿠키 저장소 (로그인 세션 유지용)
var jarMu sync.Mutex
var jar, _ = cookiejar.New(nil)

type sendResp struct {
	OK         bool        `json:"ok"`
	Error      string      `json:"error,omitempty"`
	Status     int         `json:"status"`
	StatusText string      `json:"statusText"`
	Headers    [][2]string `json:"headers"`
	Body       string      `json:"body"`
	Binary     bool        `json:"binary"`
	BodyB64    string      `json:"bodyB64,omitempty"`
	TimeMs     int64       `json:"timeMs"`
	RawRequest string      `json:"rawRequest,omitempty"` // 실제로 서버에 보낸 요청 원문
	SentAt     int64       `json:"sentAt,omitempty"`     // 요청 발송 시각 (Unix ms)
	RecvAt     int64       `json:"recvAt,omitempty"`     // 응답 수신 완료 시각 (Unix ms)
	Size       int         `json:"size"`
}

func dataFile() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		// macOS 앱 번들(.app) 안에 저장하면 앱을 교체할 때 데이터가 사라지므로 사용자 폴더 사용
		if strings.Contains(dir, ".app/Contents/MacOS") {
			goto userDir
		}
		p := filepath.Join(dir, "RestTester_data.json")
		// 쓰기 가능 여부 확인
		f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0644)
		if err == nil {
			f.Close()
			return p
		}
	}
userDir:
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "RestTester")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "RestTester_data.json")
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	var req sendReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, sendResp{Error: "요청 해석 실패: " + err.Error()})
		return
	}
	u := strings.TrimSpace(req.URL)
	if !strings.HasPrefix(strings.ToLower(u), "http://") && !strings.HasPrefix(strings.ToLower(u), "https://") {
		u = "http://" + u
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}
	var body io.Reader
	if req.Body != "" && req.Method != "GET" && req.Method != "HEAD" {
		body = strings.NewReader(req.Body)
	}
	hr, err := http.NewRequest(req.Method, u, body)
	if err != nil {
		writeJSON(w, sendResp{Error: "주소가 올바르지 않습니다: " + err.Error()})
		return
	}
	for _, h := range req.Headers {
		k := strings.TrimSpace(h[0])
		if k == "" {
			continue
		}
		if strings.EqualFold(k, "Host") {
			hr.Host = h[1]
			continue
		}
		hr.Header.Add(k, h[1])
	}
	if hr.Header.Get("User-Agent") == "" {
		hr.Header.Set("User-Agent", "RestTester/1.0")
	}
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: req.Insecure},
		DialContext:     (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}
	dt := &dumpTransport{base: tr}
	client := &http.Client{Transport: dt, Timeout: time.Duration(req.Timeout) * time.Second}
	if req.KeepCookie {
		jarMu.Lock()
		client.Jar = jar
		jarMu.Unlock()
	}
	start := time.Now()
	resp, err := client.Do(hr)
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "certificate"):
			msg = "SSL 인증서 오류입니다. 사내 서버라면 '인증서 오류 무시'를 켜고 다시 시도하세요.\n\n" + msg
		case strings.Contains(msg, "refused"):
			msg = "서버가 연결을 거부했습니다. 주소/포트가 맞는지, 서버가 켜져 있는지 확인하세요.\n\n" + msg
		case strings.Contains(msg, "Timeout") || strings.Contains(msg, "timeout"):
			msg = "응답 시간이 초과되었습니다. 네트워크나 서버 상태를 확인하세요.\n\n" + msg
		case strings.Contains(msg, "no such host"):
			msg = "서버 이름을 찾을 수 없습니다. 주소 철자를 확인하세요.\n\n" + msg
		}
		end := time.Now()
		writeJSON(w, sendResp{Error: msg, TimeMs: end.Sub(start).Milliseconds(), SentAt: start.UnixMilli(), RecvAt: end.UnixMilli(), RawRequest: dt.text()})
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 50<<20))
	end := time.Now()
	out := sendResp{
		OK:         true,
		Status:     resp.StatusCode,
		StatusText: http.StatusText(resp.StatusCode),
		TimeMs:     end.Sub(start).Milliseconds(),
		SentAt:     start.UnixMilli(),
		RecvAt:     end.UnixMilli(),
		RawRequest: dt.text(),
		Size:       len(data),
	}
	for k, vs := range resp.Header {
		for _, v := range vs {
			out.Headers = append(out.Headers, [2]string{k, v})
		}
	}
	if utf8.Valid(data) {
		out.Body = string(data)
	} else {
		out.Binary = true
		if len(data) <= 10<<20 {
			out.BodyB64 = base64.StdEncoding.EncodeToString(data)
		}
	}
	writeJSON(w, out)
}

func handleStore(w http.ResponseWriter, r *http.Request) {
	storeMu.Lock()
	defer storeMu.Unlock()
	p := dataFile()
	if r.Method == http.MethodPost {
		data, err := io.ReadAll(r.Body)
		if err != nil || !json.Valid(data) {
			http.Error(w, "bad data", 400)
			return
		}
		var buf bytes.Buffer
		json.Indent(&buf, data, "", "  ")
		tmp := p + ".tmp"
		if err := os.WriteFile(tmp, buf.Bytes(), 0644); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		os.Rename(tmp, p)
		writeJSON(w, map[string]any{"ok": true, "path": p})
		return
	}
	data, err := os.ReadFile(p)
	if err != nil || len(bytes.TrimSpace(data)) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func openBrowser(u string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	case "darwin":
		cmd = exec.Command("open", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	cmd.Start()
}

func alreadyRunning(port int) bool {
	c := http.Client{Timeout: 700 * time.Millisecond}
	resp, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/api/ping", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return strings.Contains(string(b), appID)
}

var serverAddr string

func quitApp() {
	trayRemove()
	os.Exit(0)
}

func main() {
	if alreadyRunning(preferredPort) {
		// 이미 실행 중이면 화면만 다시 열고 끝냄
		openBrowser(fmt.Sprintf("http://127.0.0.1:%d/", preferredPort))
		return
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", preferredPort))
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			showError("REST API 테스트 도구", "프로그램을 시작하지 못했습니다.\n\n"+err.Error())
			return
		}
	}
	port := ln.Addr().(*net.TCPAddr).Port
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(indexHTML)
	})
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(appID)) })
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"os": runtime.GOOS, "arch": runtime.GOARCH, "dataFile": dataFile()})
	})
	mux.HandleFunc("/api/send", handleSend)
	mux.HandleFunc("/api/store", handleStore)
	mux.HandleFunc("/api/quit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		writeJSON(w, map[string]any{"ok": true})
		go func() { time.Sleep(300 * time.Millisecond); quitApp() }()
	})
	mux.HandleFunc("/api/cookies", func(w http.ResponseWriter, r *http.Request) {
		jarMu.Lock()
		defer jarMu.Unlock()
		if r.Method == http.MethodDelete {
			jar, _ = cookiejar.New(nil)
			writeJSON(w, map[string]any{"ok": true})
			return
		}
		u, err := url.Parse(r.URL.Query().Get("url"))
		if err != nil || u.Host == "" {
			writeJSON(w, []string{})
			return
		}
		var names []string
		for _, c := range jar.Cookies(u) {
			names = append(names, c.Name)
		}
		writeJSON(w, names)
	})

	addr := fmt.Sprintf("http://127.0.0.1:%d/", port)
	serverAddr = addr
	go http.Serve(ln, mux)

	// 콘솔 창 없이 알림 영역(트레이) 아이콘으로 동작
	runTray("REST API 테스트 도구 (클릭: 화면 열기)",
		func() { openBrowser(addr) },
		quitApp,
		func() {
			trayBalloon("REST API 테스트 도구 실행 중", "화면 아래 알림 영역 아이콘을 클릭하면 화면이 열립니다.\n끝낼 때는 아이콘 오른쪽 클릭 → 종료")
			go func() { time.Sleep(300 * time.Millisecond); openBrowser(addr) }()
		})
}
