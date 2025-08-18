@echo off
REM 编译
go build -o kiro2cc.exe main.go
IF %ERRORLEVEL% NEQ 0 (
    echo 编译失败!
    pause
    exit /b 1
)
REM 压缩
upx --best --lzma kiro2cc.exe
 