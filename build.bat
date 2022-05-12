@echo off
go build -tags release -ldflags "-X 'main.buildDateTime=2022-05-12' -X 'main.gitCommitCode=' -s -w" -o OneDSS.exe