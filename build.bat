@echo off
SETLOCAL ENABLEDELAYEDEXPANSION

:: ==========================
:: 配置
:: ==========================
SET VERSION=1.0.0
SET APP_NAME=IrisAura
SET BUILD_DIR=build
SET FRONTEND_DIR=frontend

echo Starting build %APP_NAME% v%VERSION%

:: ==========================
:: 安装 Wails CLI（如果还没装的话）
:: ==========================
echo Installing Wails CLI...
go install github.com/wailsapp/wails/v2/cmd/wails@latest

:: 确保 Go bin 和 Wails CLI 在 PATH 中
SET PATH=%PATH%;%USERPROFILE%\go\bin

:: ==========================
:: 前端构建
:: ==========================
echo Building frontend...
cd %FRONTEND_DIR%
call npm install --no-fund --no-audit
call npm run build --if-present
cd ..

:: ==========================
:: Go 依赖
:: ==========================
go mod tidy
go mod download

:: ==========================
:: 创建构建目录
:: ==========================
if not exist %BUILD_DIR% (
    mkdir %BUILD_DIR%
)

:: ==========================
:: 构建 Windows 64位
:: ==========================
echo Building Windows 64-bit...
wails build -platform windows/amd64 -o %BUILD_DIR%\%APP_NAME%-windows-amd64.exe

echo Windows build finished!
pause