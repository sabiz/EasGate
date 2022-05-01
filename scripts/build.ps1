$DIST_LIST='windows/arm64', 'linux/arm64'

$ownPath = split-Path -Parent (split-Path -Parent $MyInvocation.MyCommand.Path)

cd $ownPath

foreach ($d in $DIST_LIST) {
  $sp = $d.Split("/")
  $Env:GOOS=$sp[0]
  $Env:GOARCH=$sp[1]
  $ext = ""
  if ($sp[0] -eq "windows") {
    $ext = ".exe"
  }
  echo "Build $d"
  go build -o "build/EasyGate-${Env:GOOS}-${Env:GOARCH}${ext}" main.go
}

echo "All done."
