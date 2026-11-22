REST API 테스트 도구 - macOS (Intel / Apple Silicon 공용)

[설치]
1. RestTester.app 을 "응용 프로그램(Applications)" 폴더로 옮깁니다. (다른 폴더도 무방)

[처음 실행 - 보안 경고 해제]
Apple 개발자 서명이 없는 앱이라 처음에는 "열 수 없음" 경고가 뜹니다. 아래 중 하나로 한 번만 허용하면 됩니다.

 방법 A) 시스템 설정 이용
   1) RestTester.app 더블클릭 -> 경고 창에서 [완료]
   2) 시스템 설정 > 개인정보 보호 및 보안 > 아래쪽 "RestTester... 차단됨" 옆 [그래도 열기]
   3) 다시 더블클릭 -> [열기]

 방법 B) 터미널 명령 한 줄 (앱 위치에 맞게 경로 수정)
   xattr -dr com.apple.quarantine /Applications/RestTester.app

[사용]
- 더블클릭하면 브라우저에 화면이 열립니다. Dock 아이콘 없이 백그라운드로 동작합니다.
- 브라우저 탭을 닫았으면 RestTester.app 을 다시 실행하면 화면만 다시 열립니다.
- 종료: 화면 상단 [종료] 버튼
- 저장 위치: ~/Library/Application Support/RestTester/RestTester_data.json
  (앱을 새 버전으로 바꿔도 저장된 요청은 유지됩니다)
- RestTester_Genian_NAC.json : 화면의 [불러오기]로 Genian NAC 예제 요청을 넣을 수 있습니다.
