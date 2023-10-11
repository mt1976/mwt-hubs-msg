#Before we can use the script, we have to make it executable with the chmod command:
#chmod +x ./go-executable-build.sh
#then we can use it  ./go-executable-build.sh yourpackage
#!/usr/bin/env bash

NOW=$(date +"%y%m")
figlet -f standard "Building PROTEUSHUB" 
echo Building Windows x86_64 
go-winres simply --icon app.png
env GOOS=windows GOARCH=amd64 go build  -o "./proteushub.exe" github.com/mt1976/proteushub
echo Building Crossplatform 
echo Building Windows x86_64 
go-winres simply --icon app.png
env GOOS=windows GOARCH=amd64 go build  -o "./exec/windows/proteushub_"$NOW".exe" github.com/mt1976/proteushub
echo Building MacOs x86_64 
go-winres simply --icon app.png
env GOOS=darwin GOARCH=amd64 go build -o "./exec/mac/intel/proteushub_"$NOW github.com/mt1976/proteushub 
echo Building MacOs arm64 
go-winres simply --icon app.png
env GOOS=darwin GOARCH=arm64 go build -o "./exec/mac/apple/proteushub_"$NOW github.com/mt1976/proteushub 
echo Done 