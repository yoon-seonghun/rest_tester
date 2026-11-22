#!/bin/bash
# REST API 테스트 도구 — 멀티 플랫폼 빌드 (Windows / macOS / Linux)
set -e
cd "$(dirname "$0")"
VER=${VER:-1.0}
OUT=dist; rm -rf "$OUT"; mkdir -p "$OUT/tmp"
LD="-s -w"

# ---------- Windows (x64) : 콘솔 창 없는 GUI 실행 파일 + 트레이 ----------
W="$OUT/tmp/RestTester"; mkdir -p "$W"
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$LD -H=windowsgui" -o "$W/RestTester.exe" .
cp RestTester_Genian_NAC.json "$W/"
iconv -f utf-8 -t cp949 readme/README_windows.txt > "$W/README.txt"
(cd "$OUT/tmp" && zip -qr ../RestTester_Windows.zip RestTester); rm -rf "$W"

# ---------- macOS : Intel + Apple Silicon 유니버설 .app ----------
GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$LD" -o "$OUT/tmp/mac_amd64" .
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$LD" -o "$OUT/tmp/mac_arm64" .
M="$OUT/tmp/RestTester"; APP="$M/RestTester.app/Contents"; mkdir -p "$APP/MacOS" "$APP/Resources"
python3 tools/lipo.py "$APP/MacOS/RestTester" "$OUT/tmp/mac_amd64" "$OUT/tmp/mac_arm64"
chmod +x "$APP/MacOS/RestTester"
cp icon.icns "$APP/Resources/RestTester.icns"
sed "s/@VER@/$VER/g" tools/Info.plist > "$APP/Info.plist"
cp RestTester_Genian_NAC.json "$M/"; cp readme/README_mac.txt "$M/README.txt"
(cd "$OUT/tmp" && zip -qry ../RestTester_macOS.zip RestTester); rm -rf "$M" "$OUT"/tmp/mac_*

# ---------- Linux : x64 / ARM64 ----------
L="$OUT/tmp/RestTester"; mkdir -p "$L"
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$LD" -o "$L/RestTester-linux-x64" .
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$LD" -o "$L/RestTester-linux-arm64" .
chmod +x "$L"/RestTester-linux-*
cp RestTester_Genian_NAC.json "$L/"; cp readme/README_linux.txt "$L/README.txt"
(cd "$OUT/tmp" && tar czf ../RestTester_Linux.tar.gz RestTester); rm -rf "$OUT/tmp"

ls -la "$OUT"
