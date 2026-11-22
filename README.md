# rest_tester
REST API TESTER , 윈도우용 , MAC용 , Linux용 

# RestTester 소스 매뉴얼

Nov 22, 2026 · @윤성훈

## 1. 개요

RestTester는 Go 실행 파일 하나에 웹 화면(HTML)을 넣어 둔 REST API 테스트 도구다. 실행하면 내 PC의 `127.0.0.1:18765`에 작은 HTTP 서버를 띄우고, 브라우저로 화면을 연다. 화면에서 보내기를 누르면 브라우저가 직접 대상 서버를 부르지 않고, 이 로컬 서버가 대신 요청을 보낸 뒤 결과를 화면에 돌려준다.

이렇게 만든 이유는 세 가지다.

- 브라우저의 CORS 제한 없이 사내 서버(Spring Boot, Genian NAC 등)를 호출할 수 있다.
- 설치나 런타임(Python, Java, Node)이 필요 없다. 실행 파일 하나만 있으면 된다.
- 화면은 HTML 한 장이라 수정이 쉽고, Windows·macOS·Linux에서 같은 화면을 쓴다.

외부에서 이 서버로 접속할 수 없다. `127.0.0.1`에만 열리기 때문에 같은 PC의 브라우저만 화면을 쓸 수 있다.

&#91;embedded content: RestTester 구조 · 화면, 로컬 서버, 대상 서버\]

응답은 같은 길을 거꾸로 돌아온다. 화면은 대상 서버와 직접 통신하지 않고, 파일도 직접 읽거나 쓰지 않는다.

## 2. 소스 파일 구성

프로그램 로직은 `main.go`(서버)와 `index.html`(화면) 두 파일에 거의 다 들어 있다. 나머지는 플랫폼별 처리와 빌드·배포용 파일이다.

| 파일 | 역할 | 비고 |
| --- | --- | --- |
| `main.go` | 로컬 HTTP 서버, API 엔드포인트, 요청 대리 전송, 저장 파일 읽기/쓰기 | 모든 플랫폼 공통 |
| `index.html` | 화면 전체(HTML + CSS + JavaScript 한 파일) | `go:embed`로 실행 파일에 들어감 |
| `tray_windows.go` | Windows 알림 영역(트레이) 아이콘, 메시지 창 | Windows 빌드에만 포함 (`//go:build windows`) |
| `tray_other.go` | macOS·Linux용 대체 구현(콘솔 안내, Ctrl+C 종료) | `//go:build !windows` |
| `icon.bin` | 트레이 아이콘 이미지(32×32 BMP 데이터) | `tray_windows.go`가 embed |
| `icon.icns` | macOS 앱 아이콘 | `.app` 번들에 복사됨 |
| `build.sh` | 3개 플랫폼 배포본을 한 번에 만드는 스크립트 | 결과는 `dist/` |
| `tools/Info.plist` | macOS 앱 번들 설정(이름, 버전, Dock 아이콘 숨김) | `@VER@`을 빌드 때 치환 |
| `tools/lipo.py` | Intel·Apple Silicon 실행 파일을 유니버설 파일 하나로 합침 | macOS `lipo` 명령 대체 |
| `readme/README_*.txt` | 플랫폼별 사용자 안내문 | Windows용은 빌드 때 CP949로 변환 |
| `RestTester_Genian_NAC.json` | 화면의 \[불러오기\]용 예제 요청 모음 | 배포본에 함께 들어감 |
| `go.mod` | Go 모듈 선언 | 외부 라이브러리 없음(표준 라이브러리만 사용) |

## 3. 빌드 방법

Go 1.24 이상과 bash, zip, python3, iconv가 있는 PC(Linux 또는 macOS)에서 `./build.sh` 한 줄이면 세 플랫폼 배포본이 `dist/`에 만들어진다. 외부 Go 라이브러리를 쓰지 않아서 인터넷 연결 없이도 빌드된다. Windows에서는 WSL이나 Git Bash를 쓰면 된다.

```
cd RestTester_source
./build.sh              # 버전 지정: VER=1.1 ./build.sh
```

| 산출물 | 대상 | 내용 |
| --- | --- | --- |
| `RestTester_Windows.zip` | Windows 10/11 x64 | `RestTester.exe`(콘솔 창 없는 GUI 실행 파일), 예제 JSON, README |
| `RestTester_macOS.zip` | macOS 11 이상, Intel + Apple Silicon | `RestTester.app`(유니버설), 예제 JSON, README |
| `RestTester_Linux.tar.gz` | Linux x64 / ARM64 | `RestTester-linux-x64`, `RestTester-linux-arm64`, 예제 JSON, README |

한 플랫폼만 빠르게 만들 때는 Go 명령만 쓴다. Go는 어느 OS에서나 다른 OS용 파일을 만들 수 있다(크로스 컴파일).

```
# Windows (콘솔 창 숨김: -H=windowsgui)
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -H=windowsgui" -o RestTester.exe .

# 개발 중 확인용 (콘솔 창 보임, 로그 확인 가능)
GOOS=windows GOARCH=amd64 go build -o RestTester_debug.exe .

# macOS Apple Silicon / Linux
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o RestTester-mac .
GOOS=linux  GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o RestTester-linux-x64 .
```

- `-s -w`는 디버그 정보를 빼서 파일 크기를 줄인다. `-trimpath`는 빌드한 PC의 폴더 경로가 실행 파일에 남지 않게 한다.
- `index.html`을 고친 뒤에도 반드시 다시 빌드해야 반영된다. 화면 파일이 실행 파일 안에 들어가 있기 때문이다.
- macOS Apple Silicon용 파일은 Go 링커가 임시 서명(ad-hoc)을 자동으로 넣는다. 서명이 없으면 Apple Silicon에서 실행되지 않는다.

## 4. 백엔드 (main.go)

`main.go`는 화면 파일을 내보내고, 화면이 부르는 API 7개를 처리한다. 외부 라이브러리 없이 Go 표준 라이브러리(`net/http`, `net/http/cookiejar`, `net/http/httputil`)만 쓴다.

### 4.1 시작 순서 (`main()`)

1. `127.0.0.1:18765`에 이미 RestTester가 떠 있는지 `/api/ping`으로 확인한다. 응답이 `REST-TESTER-v1`이면 브라우저만 열고 종료한다(중복 실행 방지).
2. 포트 18765를 열고, 다른 프로그램이 쓰고 있으면 빈 포트를 아무거나 잡는다. 둘 다 실패하면 오류 메시지 창을 띄운다.
3. 라우트를 등록하고 HTTP 서버를 별도 고루틴으로 띄운다.
4. `runTray()`를 호출한다. Windows는 트레이 아이콘을 띄우고, Mac·Linux는 콘솔에 안내를 출력한 뒤 종료 신호를 기다린다. 준비가 끝나면 기본 브라우저로 화면을 연다.

### 4.2 API 엔드포인트

| 경로 | 방식 | 하는 일 |
| --- | --- | --- |
| `/` | GET | 화면(`index.html`) 반환. 캐시 안 함(`no-store`) |
| `/api/ping` | GET | `REST-TESTER-v1` 문자열 반환. 중복 실행 확인용 |
| `/api/info` | GET | OS, CPU 종류, 저장 파일 경로 반환(사용법 창에 표시) |
| `/api/send` | POST | 대상 서버로 요청을 대신 보내고 결과를 JSON으로 반환 (4.3) |
| `/api/store` | GET / POST | 저장 파일 전체 읽기 / 통째로 저장 (4.4) |
| `/api/cookies` | GET / DELETE | `?url=` 대상에 보관 중인 쿠키 이름 목록 / 쿠키 전부 초기화 |
| `/api/quit` | POST | 응답을 보낸 뒤 0.3초 후 트레이 아이콘을 지우고 프로그램 종료 |

### 4.3 요청 대리 전송 (`handleSend`)

화면은 보낼 내용을 이미 완성된 형태(변수 치환, 인증 헤더, 본문 인코딩까지 끝난 상태)로 서버에 넘긴다. 서버는 그대로 전송만 한다.

```
화면 → /api/send   {"method":"PUT","url":"https://...","headers":[["이름","값"]],
                     "body":"password=...&isEncrytion=true","insecure":false,
                     "timeout":30,"keepCookie":true}

/api/send → 화면   {"ok":true,"status":200,"statusText":"OK","headers":[[...]],
                     "body":"...","binary":false,"bodyB64":"","size":1234,"timeMs":84,
                     "sentAt":1795...,"recvAt":1795...,"rawRequest":"PUT /mc2/... HTTP/1.1..."}
```

- 주소에 `http://`나 `https://`가 없으면 `http://`를 붙인다. GET·HEAD는 본문을 보내지 않는다.
- `Host` 헤더는 헤더 목록이 아니라 `req.Host`로 넣는다(Go 규칙). `User-Agent`가 없으면 `RestTester/1.0`을 붙인다.
- 연결 대기 10초, 전체 대기는 화면의 설정값(기본 30초). OS의 프록시 환경변수(`HTTPS_PROXY` 등)를 따른다.
- 인증서 오류 무시(`insecure`)를 켜면 TLS 검증을 끄고 보낸다.
- 응답 본문은 최대 50MB까지 읽는다. UTF-8 텍스트가 아니면 `binary:true`로 표시하고, 10MB 이하면 Base64(`bodyB64`)로 넘겨 화면에서 이미지 표시나 파일 저장이 되게 한다.
- 연결 실패 메시지(인증서, 연결 거부, 시간 초과, 서버 이름 없음)는 한글 안내문을 앞에 붙여 `error`로 돌려준다.
- `sentAt`은 전송 직전, `recvAt`은 응답 본문을 다 읽은 시각이다(Unix 밀리초). `timeMs`는 둘의 차이다.

**보낸 요청 원문 (`dumpTransport`)**: `http.Client`의 전송 계층을 감싼 래퍼다. 실제로 나가는 요청을 `httputil.DumpRequestOut`으로 기록해서, 쿠키나 `Content-Length`처럼 Go가 자동으로 붙인 헤더까지 남는다. 리다이렉트로 여러 번 보내면 모두 이어 붙인다.

**세션(쿠키) 유지**: 프로그램 전체에서 쿠키 저장소(`cookiejar`) 하나를 공유한다. `keepCookie:true`인 요청만 이 저장소를 쓴다. 저장소는 메모리에만 있어서 프로그램을 종료하면 사라진다.

### 4.4 저장 파일 (`handleStore`, `dataFile`)

화면의 상태(저장된 요청, 최근 기록, 변수 등)는 JSON 파일 하나 `RestTester_data.json`에 통째로 저장한다. 서버는 내용을 해석하지 않고, 올바른 JSON인지만 확인해서 저장한다. 임시 파일에 먼저 쓴 뒤 이름을 바꿔서, 저장 도중에 꺼져도 파일이 깨지지 않는다.

| 조건 | 저장 위치 |
| --- | --- |
| 기본 | 실행 파일과 같은 폴더 |
| 실행 파일 폴더에 쓰기 권한 없음 | OS 사용자 설정 폴더의 `RestTester/` (Windows `%APPDATA%`, Linux `~/.config`) |
| macOS `.app` 안에서 실행 | `~/Library/Application Support/RestTester/` (앱을 교체해도 데이터 유지) |

## 5. 플랫폼별 동작 (트레이)

두 파일은 같은 함수 5개(`runTray`, `trayBalloon`, `trayRemove`, `showMessage`, `showError`)를 각자 구현한다. Go의 빌드 태그가 OS에 맞는 파일 하나만 골라서 컴파일하므로, `main.go`는 OS를 신경 쓰지 않고 이 함수들을 부른다.

### 5.1 Windows (`tray_windows.go`)

외부 라이브러리 없이 `syscall`로 Win32 API를 직접 호출한다. 배포용 exe는 `-H=windowsgui`로 빌드해서 콘솔 창이 없고, 트레이 아이콘이 유일한 표시다.

1. 화면에 보이지 않는 창을 하나 만든다(`RegisterClassExW`, `CreateWindowExW`). 트레이 이벤트를 받기 위한 창이다.
2. `icon.bin`으로 아이콘을 만들고(`CreateIconFromResourceEx`, 실패하면 Windows 기본 아이콘), `Shell_NotifyIconW`로 트레이에 등록한다.
3. 메시지 루프(`GetMessageW`)를 돈다. 이 스레드는 `runtime.LockOSThread()`로 고정한다. Win32 창은 만든 스레드에서만 메시지를 받기 때문이다.
4. 아이콘 왼쪽 클릭은 화면 열기, 오른쪽 클릭은 팝업 메뉴(화면 열기 / 종료)다.
5. 탐색기가 재시작되면 `TaskbarCreated` 메시지를 받아 아이콘을 다시 등록한다.

Win32 구조체(`notifyIconData`, `wndClassEx`, `msg`)는 64비트 Windows의 메모리 배치와 크기가 같아야 한다(각각 976, 80, 48 바이트로 확인함). 필드 순서나 타입을 바꾸면 아이콘이 안 뜨거나 프로그램이 죽을 수 있다.

### 5.2 macOS · Linux (`tray_other.go`)

트레이 아이콘은 없다. 콘솔에 접속 주소와 저장 위치를 출력하고, Ctrl+C나 종료 신호(`SIGTERM`)를 받으면 끝난다. 화면의 \[종료\] 버튼은 `/api/quit`을 불러 어느 OS에서나 똑같이 동작한다.

macOS `.app`은 `Info.plist`의 `LSUIElement`를 켜서 Dock 아이콘 없이 백그라운드로 실행된다. Go 프로그램은 macOS 이벤트 루프를 돌지 않아서, Dock에 아이콘을 띄우면 응답 없음으로 보이기 때문이다.

브라우저 열기(`openBrowser`)는 OS별 명령을 쓴다: Windows `rundll32 url.dll,FileProtocolHandler`, macOS `open`, Linux `xdg-open`.

## 6. 프론트엔드 (index.html)

화면은 프레임워크 없는 순수 JavaScript 한 파일이다. 모든 상태는 전역 객체 `store` 하나에 담고, 바뀔 때마다 `persist()`가 0.3초 묶어서 `/api/store`로 통째로 저장한다. 브라우저 저장소(localStorage)는 쓰지 않는다. 포트가 바뀌면 브라우저는 다른 사이트로 보고 데이터를 따로 두기 때문이다.

### 6.1 저장 데이터 구조 (`RestTester_data.json`)

```
{
  "saved":   [ 요청, ... ],          // 저장된 요청 (왼쪽 목록)
  "history": [ 요청+결과, ... ],     // 최근 기록, 최신순 최대 100건
  "envs":    [ {"name":"Genian NAC", "vars":[["baseUrl","https://...",true], ...]} ],
  "envIdx":  1,                          // 상단에서 고른 환경 번호
  "tvHidden": { "<열 목록 서명>": ["숨긴 열", ...] }  // 표 보기 열 선택
}

요청 = { id, name, folder, method, url,
         params:[[이름,값,사용여부]], headers:[[...]],
         bodyType:"none|json|form|text", body, form:[[...]],
         auth:{type:"none|bearer|basic|apikey", token, user, pass, keyName, keyVal, keyIn:"query|header"},
         insecure, timeout, keepCookie }

기록 = 요청 + { at, status, htmlRes, sentUrl, res:{/api/send 응답 전체} }
```

- 키/값 목록은 모두 `[이름, 값, 켜짐여부]` 배열이다. 체크를 끈 항목은 저장은 되지만 전송하지 않는다.
- 최근 기록 중 마지막 10건은 응답 본문을 1MB까지, 그 이전 기록은 2만 자까지만 남긴다.
- 불러오기는 같은 폴더·같은 이름이면 덮어쓴다. 시작할 때 내용이 완전히 같은 저장 요청은 자동으로 하나만 남긴다.

### 6.2 요청이 만들어지는 순서

1. `collect()`가 화면 입력값을 요청 객체로 모은다. 반대로 `apply()`는 요청 객체를 화면에 채운다.
2. `build()`가 실제 전송 형태로 바꾼다. `sub()`로 `{{변수}}`를 치환하고, 본문 형식에 맞는 `Content-Type`을 붙이고(직접 넣은 값이 우선), 인증값을 헤더나 주소에 넣는다.
3. `send()`가 보내기 전에 검사한다. 남은 변수, 깨진 JSON, 주소의 `//`·역슬래시·한글·공백, 주소에 들어간 비밀번호를 확인한 뒤 `/api/send`를 호출한다.
4. 결과를 `showRes()`가 세 곳에 나눠 그린다. `showResRaw()`는 원문과 상태 줄, `renderTable()`은 표 보기, `showReq()`는 보낸 요청이다.

### 6.3 코드 구역

스크립트는 `/* ---------- 구역명 ---------- */` 주석으로 나뉘어 있다. 편집기에서 구역명으로 검색하면 바로 찾을 수 있다.

| 구역 | 주요 함수 | 내용 |
| --- | --- | --- |
| 저장소 | `persist`, `load` | 서버에 저장·읽기, 중복 요청 정리 |
| 키/값 테이블 | `kvTable`, `kvGet`, `kvActive` | 파라미터·헤더·폼·변수 입력 표 (마지막 줄에 입력하면 새 줄 자동 추가) |
| 파라미터 ↔ URL 동기화 | `urlToParams`, `paramsToUrl` | 주소의 `?…` 부분과 파라미터 탭을 서로 맞춤 |
| 요청 상태 수집/복원 | `collect`, `apply` | 화면 ↔ 요청 객체 |
| 변수 치환 | `vars`, `sub`, `build` | `{{변수}}` 치환, 전송용 요청 조립 |
| 보내기 | `send` | 검사, 전송, 기록 저장 |
| 표 보기 | `renderTable`, `drawTable`, `colPicker`, `csvExport` | JSON → 표, 정렬·검색·날짜 변환·CSV |
| (응답 표시) | `showRes`, `showResRaw`, `showReq`, `fmtTs`, `hlJSON` | 원문, 상태 줄, 보낸 요청, JSON 색 구분 |
| 사이드바 | `side` | 저장된 요청 / 최근 기록 목록 |
| 저장 모달 | `openSave`, `doSave` | 이름·폴더를 정해 요청 저장 |
| 환경/변수 | `envSel`, `envTable` | 환경 선택, 변수 관리 창 |
| 내보내기/불러오기 | (이벤트) | 저장된 요청·환경을 JSON 파일로 주고받음 |
| 기타 이벤트 | `cookieInfo` 등 | 버튼, 단축키(Ctrl/⌘ + Enter·S), 종료 버튼 |

### 6.4 표 보기 동작

- 응답 JSON을 5단계 깊이까지 훑어 객체 목록(배열)을 찾고, 가장 긴 목록을 처음에 보여준다(`findArrays`). 목록이 여러 개면 ‘데이터’ 선택 상자가 나온다.
- 목록이 아니면 ‘항목 / 값’ 두 칸 표로 보여준다. 안쪽에 묶음 데이터가 있는 셀은 보라색으로 표시되고, 누르면 그 안으로 들어간다(`tv.path`).
- 날짜 변환: 2000\~2100년 범위의 13자리 정수(밀리초)는 날짜로 바꾼다. 10자리(초)는 열 이름이 `time`, `date`, `created`, `updated`, `last`, `at` 등으로 끝날 때만 바꾼다(`cellText`).
- 화면에는 1,000행까지 그린다. CSV 저장은 걸러낸·정렬된 전체 행을 UTF-8 BOM을 붙여 내보낸다. BOM이 있어야 엑셀에서 한글이 깨지지 않는다.

### 6.5 원문 검색과 크게 보기

- **검색** (`/* ---------- 원문 / 보낸 요청 검색 ---------- */`): `addSearch(판ID)`가 원문·보낸 요청 탭 위에 검색창을 붙인다. `findIn()`은 `pre.out` 안의 글자를 이어 붙여 대소문자 구분 없이 찾고, 찾은 위치를 브라우저의 CSS Custom Highlight API(`CSS.highlights`)로 칠한다. JSON 색 구분 때문에 글자가 여러 태그로 나뉘어 있어도 찾는다. 최대 5,000건. 응답이 새로 오면 같은 검색어로 다시 찾는다.
- **크게 보기** (`/* ---------- 결과 영역 크게 보기 ---------- */`): `popRes(true)`는 결과 카드(`.card.res`) 요소를 그대로 전체 화면 레이어(`#resPop`)로 옮기고, 닫으면 원래 자리 표식(주석 노드 `res-home`) 뒤로 되돌린다. 복사본을 만들지 않고 요소를 옮기므로 탭 상태, 검색, 표 정렬이 그대로 유지된다.
- 단축키: Esc는 열린 창을 안쪽부터 닫는다(변수·저장 창 → 열 선택 → 크게 보기). 검색어가 있는 검색창에서는 검색어부터 지운다. Ctrl/⌘+F는 크게 보기 중이거나 결과 영역에 커서가 있을 때만 검색창으로 가고, 그 외에는 브라우저 기본 찾기가 동작한다.

### 6.6 사이드바 선택 상태

마우스를 올린 항목(회색, `:hover`)과 선택된 항목(파란 배경 + 왼쪽 파란 줄, `.sel`)은 별개다. 선택은 변수 두 개로 관리한다. `currentId`는 저장된 요청, `histSel`은 최근 기록이며, 둘 중 하나만 값을 갖는다.

| 상황 | 저장된 요청 선택 | 최근 기록 선택 |
| --- | --- | --- |
| 항목 클릭 | 선택(반대쪽 해제) | 선택(반대쪽 해제) |
| 내용 수정 | 유지 + `● 수정됨` 표시 | 해제 (새 내용) |
| 보내기 | 유지 | 해제 (새 기록이 맨 위에 추가됨) |
| 저장 | 저장한 항목 선택, `수정됨` 사라짐 | 해제 |
| 새 요청 | 해제 | 해제 |

내용 수정 여부는 `onReqEdit()`가 요청 입력 영역의 입력 이벤트를 받아 0.15초 뒤 `sameReq()`로 비교한다. 비교 대상은 방식, 주소, 헤더, 본문, 인증, 설정이며 `normReq()`가 빈 행과 기본값 차이를 없앤 뒤 비교한다.

## 7. 자주 하는 수정

어느 수정이든 고친 뒤에는 다시 빌드해야 반영된다. 화면만 고칠 때도 마찬가지다.

| 하고 싶은 것 | 고칠 곳 |
| --- | --- |
| 포트 변경 | `main.go`의 `preferredPort = 18765`. 바꾸면 브라우저는 다른 사이트로 보지만, 저장 파일은 서버 쪽에 있어서 그대로 유지된다. |
| 구버전과 동시 실행 막기 | `appID`(현재 `REST-TESTER-v1`)가 같아야 중복 실행으로 인식한다. 바꾸면 구버전이 포트를 잡고 있을 때 새 버전은 다른 포트로 뜨니 주의한다. |
| 기본 대기 시간 | 서버 기본값은 `handleSend`의 `req.Timeout = 30`, 화면 기본값은 `index.html`의 `optTimeout` 입력칸 |
| 응답 크기 제한 | `handleSend`의 `50<<20`(50MB)과 바이너리 표시 제한 `10<<20`(10MB) |
| 기록 개수·보관 크기 | `send()`의 `slice(0,100)`(기록 수)과 `cap=i<10?1048576:20000`(본문 보관 크기) |
| 보내기 전 검사 추가 | `send()` 안의 `warn.push(…)` 목록에 조건을 한 줄 추가 |
| 인증 방식 추가 | 화면: `authType` 선택 상자와 입력칸, `showAuth()` / 전송: `build()`의 `a.type` 분기 / 저장: `collect()`·`apply()`의 `auth` |
| 본문 형식 추가 (예: multipart 파일 업로드) | 화면: `name="bt"` 라디오와 `showBody()` / 전송: `build()`의 `bodyType` 분기. 파일은 텍스트 `body`로 못 보내므로 `sendReq`에 필드를 추가하고 서버에서 `mime/multipart`로 조립해야 한다. |
| 서버 API 추가 | `main()`의 `mux.HandleFunc(...)` 목록에 등록하고, 화면에서는 `fetch("/api/...")`로 호출 |
| 표 보기 날짜 판별 규칙 | `cellText()`의 범위 값과 열 이름 정규식 |
| 화면 색·글꼴 | `index.html` 상단 `<style>`의 `:root` 변수(`--accent`, `--mono` 등) |
| 사용법 창 문구 | `index.html`의 `helpModal` 영역 |
| 트레이 메뉴·풍선 문구 | 메뉴: `tray_windows.go`의 `showMenu()` / 풍선·툴팁: `main.go` 끝부분 `runTray(...)` 호출 인자 |
| 아이콘 교체 | 트레이: `icon.bin`(32×32, 32비트 BMP 헤더 + 픽셀 + 마스크) / macOS: `icon.icns`. 두 파일은 같은 그림을 Python(Pillow)으로 그려 만든 것이다. exe 파일 자체의 아이콘은 현재 없다(리소스 편집 도구 필요). |
| 버전 표기 | `VER=1.1 ./build.sh` (macOS `Info.plist`에 들어감) |
| 예제 요청 수정 | `RestTester_Genian_NAC.json`. 화면에서 요청을 만든 뒤 \[내보내기\]한 파일을 그대로 써도 된다. API 키는 비워 두고 배포한다. |

수정한 뒤에는 콘솔 창이 보이는 디버그용 exe(3장의 `go build -o RestTester_debug.exe .`)로 먼저 실행해 보면 오류를 바로 확인할 수 있다. 화면 쪽 오류는 브라우저의 개발자 도구(F12) 콘솔에 나온다.

## 8. 주의사항과 알려진 제약

**보안**

- 저장 파일은 암호화하지 않는다. API 키, 비밀번호, 보낸 요청 원문, 응답 본문(개인정보 포함 가능)이 그대로 들어 있다. 파일을 다른 사람에게 넘기기 전에는 최근 기록을 지우고, 공유는 \[내보내기\] 파일로 한다.
- &#91;내보내기\] 파일에는 응답은 없지만 환경 변수(API 키 포함)는 들어간다.
- 로컬 서버 API에는 인증이 없다. `127.0.0.1`에만 열려 있어 다른 PC에서는 접속할 수 없지만, 같은 PC의 다른 프로그램은 호출할 수 있다. 여러 사람이 함께 쓰는 서버(원격 데스크톱 등)에서는 쓰지 않는 것이 안전하다.
- 서버를 `0.0.0.0`으로 열어 외부에서 접속하게 바꾸면 안 된다. 누구나 이 PC를 거쳐 사내망으로 요청을 보낼 수 있게 된다.

**배포**

- 실행 파일에 코드 서명이 없다. Windows는 SmartScreen 경고가, macOS는 ‘열 수 없음’ 경고가 처음 실행 때 뜬다. 넓게 배포하려면 Windows는 코드 서명 인증서, macOS는 Apple 개발자 계정으로 서명과 공증이 필요하다.
- 창 없이 실행되는 서명 없는 exe는 백신이 의심할 수 있다. 막히면 사내 백신에 예외 등록이 필요하다.
- 브라우저나 메신저는 exe 다운로드를 막는 경우가 많아서 zip으로 배포한다.

**확인되지 않은 부분 (2026-11-22 기준)**

- Windows 트레이 아이콘의 표시·클릭·메뉴: 빌드와 구조체 크기만 확인했고, 실제 Windows에서 눌러 보진 않았다.
- macOS 실행: 유니버설 파일 구조와 Apple Silicon 서명 포함 여부만 확인했다.
- Linux x64는 실제로 실행해 화면, 저장, 종료까지 확인했다.

**기능 제약**

- multipart/form-data(파일 업로드) 본문은 지원하지 않는다.
- 세션 쿠키는 프로그램을 종료하면 사라진다.
- 화면은 동시에 여러 탭으로 열면 마지막에 저장한 탭의 내용이 남는다. 탭 간 동기화는 없다.
- WebSocket, gRPC처럼 HTTP 요청·응답이 아닌 통신은 테스트할 수 없다.
