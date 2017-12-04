if ($Env:APPVEYOR_REPO_TAG -eq $true) {
  curl -T build/mbt_windows_x86.zip -u "buddyspike:$($Env:BINTRAY_APIKEY)" "https://api.bintray.com/content/buddyspike/bin/mbt_windows_x86/$($Env:APPVEYOR_REPO_TAG_NAME)/mbt_windows_x86.zip"
} else {
  Write-Host "Publishing skipped - not a tagged build"
}
