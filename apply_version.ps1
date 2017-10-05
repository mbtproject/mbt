$v = $Env:APPVEYOR_REPO_COMMIT.Substring(0,4)
(Get-Content .\cmd\version.go).replace('#development#', $v) | Set-Content .\cmd\version.go