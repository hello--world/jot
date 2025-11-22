@echo off
go build -o jot.exe .
if %errorlevel% == 0 (
    echo Build successful!
) else (
    echo Build failed!
    exit /b %errorlevel%
)

