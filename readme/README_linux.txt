REST API 테스트 도구 - Linux

[파일]
- RestTester-linux-x64   : 일반 PC/서버 (Intel/AMD 64비트)
- RestTester-linux-arm64 : ARM 64비트 (라즈베리파이 4/5, ARM 서버 등)
  내 PC가 어떤 것인지 모르면 터미널에서  uname -m  (x86_64 -> x64, aarch64 -> arm64)

[실행]
  tar xzf RestTester_Linux.tar.gz
  cd RestTester
  ./RestTester-linux-x64
- 실행 권한이 없다고 나오면:  chmod +x RestTester-linux-*
- 브라우저가 자동으로 열립니다. 안 열리면 터미널에 표시된 주소(http://127.0.0.1:18765/)를 직접 여세요.
- 종료: 화면 상단 [종료] 버튼 또는 터미널에서 Ctrl+C
- 저장 위치: 실행 파일 옆 RestTester_data.json
  (실행 파일 폴더에 쓰기 권한이 없으면 ~/.config/RestTester/)
- 파일 관리자에서 더블클릭으로 실행해도 됩니다. (이때는 화면의 [종료] 버튼으로 끝내세요)
- 화면이 없는 서버에서는 같은 PC의 브라우저로만 접속할 수 있습니다. (127.0.0.1 전용, 외부 접속 차단)
