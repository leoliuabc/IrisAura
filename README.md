# IrisAura - 安装与构建指南

> 一个基于 Wails v2 的跨平台图片压缩桌面应用

## 📋 目录

- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [构建应用](#构建应用)
- [Windows 专用构建指南](#windows-专用构建指南)
- [打包发布](#打包发布)
- [自定义配置](#自定义配置)
- [优化构建](#优化构建)
- [常见问题](#常见问题)
- [使用说明](#使用说明)

## 🔧 环境要求

### 开发环境
- **Go**: 1.21+
- **Wails**: v2
- **Node.js**: 18+
- **操作系统**: 
  - macOS 10.15+ (Catalina)
  - Windows 10/11
  - Linux Ubuntu 18.04+ 或其他现代发行版

### 运行环境
- **macOS**: 10.13+ (High Sierra)
- **Windows**: Windows 7+
- **Linux**: 支持 GTK3 的发行版

## 🚀 快速开始

### 1. 安装 Wails

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 添加到 PATH（如果需要）
export PATH=$PATH:$(go env GOPATH)/bin
```

### 2. 项目初始化

```bash
# 创建项目
wails init -n IrisAura -t vanilla
cd IrisAura

# 初始化 Go 模块
go mod init IrisAura

# 安装前端依赖
cd frontend
npm install
npm run build
cd ..

# 安装 Go 依赖
go mod tidy
go mod download
```

### 3. 开发模式运行

```bash
# 启动开发服务器
wails dev
```

## 🔨 构建应用

### 单平台构建（推荐）

```bash
# 构建当前平台版本
wails build
```

### 跨平台构建

#### Windows 64位
```bash
GOOS=windows GOARCH=amd64 wails build \
  -platform windows/amd64 \
  -o build/IrisAura-windows-amd64.exe
```

#### macOS Intel
```bash
GOOS=darwin GOARCH=amd64 wails build \
  -platform darwin/amd64 \
  -o build/IrisAura-macos-intel.app
```

#### macOS Apple Silicon
```bash
GOOS=darwin GOARCH=arm64 wails build \
  -platform darwin/arm64 \
  -o build/IrisAura-macos-arm64.app
```

#### Linux 64位
```bash
GOOS=linux GOARCH=amd64 wails build \
  -platform linux/amd64 \
  -o build/IrisAura-linux-amd64
```

### 批量构建脚本

创建 `build.sh` 文件：

```bash
#!/bin/bash
set -e

# ==========================
# 配置
# ==========================
VERSION="1.0.0"
APP_NAME="IrisAura"
BUILD_DIR="build"
FRONTEND_DIR="frontend"

echo "🚀 开始构建 ${APP_NAME} v$VERSION"

# ==========================
# 安装 Wails CLI
# ==========================
echo "📦 安装 Wails CLI..."
go install github.com/wailsapp/wails/v2/cmd/wails@latest
export PATH=$PATH:$(go env GOPATH)/bin

# ==========================
# 前端构建
# ==========================
echo "🖼️ 构建前端资源..."
cd $FRONTEND_DIR
npm install
npm run build
npm audit fix --force
cd ..

# ==========================
# Go 依赖
# ==========================
go mod tidy
go mod download

# ==========================
# 创建构建目录
# ==========================
mkdir -p $BUILD_DIR/bin

# ==========================
# 平台构建函数
# ==========================
build_platform() {
    local os=$1
    local arch=$2
    echo "🔨 构建 $os/$arch ..."
    GOOS=$os GOARCH=$arch wails build -platform $os/$arch
}

# ==========================
# macOS 构建（Intel + ARM -> Universal）
# ==========================
if [[ "$(uname -s)" == "Darwin" ]]; then
    echo "🍎 构建 macOS..."

    # 构建 Intel
    build_platform darwin amd64
    mv build/bin/${APP_NAME}.app build/bin/${APP_NAME}-intel.app

    # 构建 ARM64
    build_platform darwin arm64
    mv build/bin/${APP_NAME}.app build/bin/${APP_NAME}-arm64.app

    # 合并成 Universal
    echo "🔗 合并为 Universal 二进制..."
    cp -R build/bin/${APP_NAME}-intel.app build/bin/${APP_NAME}.app
    lipo -create \
        build/bin/${APP_NAME}-intel.app/Contents/MacOS/${APP_NAME} \
        build/bin/${APP_NAME}-arm64.app/Contents/MacOS/${APP_NAME} \
        -output build/bin/${APP_NAME}.app/Contents/MacOS/${APP_NAME}

    # 自签名（ad-hoc）
    echo "🔐 自签名..."
    codesign --deep --force --sign - build/bin/${APP_NAME}.app

    echo "✅ macOS 构建完成: build/bin/${APP_NAME}.app"
fi

# ==========================
# Linux 构建
# ==========================
if [[ "$(uname -s)" == "Linux" ]]; then
    echo "🐧 构建 Linux..."
    build_platform linux amd64
    echo "✅ Linux 构建完成: build/bin/${APP_NAME}"
fi

echo "🎉 全平台构建完成！输出文件在 $BUILD_DIR/bin 目录中"
```

使用方法：
```bash
chmod +x build.sh
./build.sh
```

## 🪟 Windows 专用构建指南

由于 Windows 下需要 CGO 编译 libwebp，需要特殊配置。

### 1. 安装 C 编译环境

#### 安装 MSYS2

1. 下载并安装 [MSYS2](https://www.msys2.org/)

2. 打开 **MSYS2 MinGW 64-bit** 终端，安装编译器：
```bash
pacman -Syu
pacman -S mingw-w64-x86_64-gcc
```

3. 将 GCC 添加到系统 PATH：
```powershell
# PowerShell 中执行
$env:PATH = "C:\msys64\mingw64\bin;$env:PATH"
```

4. 验证安装：
```powershell
gcc --version
```

### 2. 配置 libwebp

#### 下载预编译库

1. 从 [Google WebP 官网](https://developers.google.com/speed/webp/download) 下载 Windows 预编译版本

2. 解压后设置环境变量：
```powershell
$env:CGO_CFLAGS = '-IC:\libwebp-1.6.0-windows-x64\include'
$env:CGO_LDFLAGS = '-LC:\libwebp-1.6.0-windows-x64\lib -lwebp'
```

### 3. Windows 构建步骤

```powershell
# 1. 构建前端
cd frontend
npm install --no-fund --no-audit
npm run build --if-present
cd ..

# 2. 构建应用
wails build -platform windows/amd64 -o build\IrisAura-windows-amd64.exe
```

## 📦 打包发布

### 创建发布包

```bash
# 进入构建目录
cd build

# Windows
zip -r IrisAura-v1.0.0-windows-amd64.zip IrisAura-windows-amd64.exe

# macOS
tar -czf IrisAura-v1.0.0-macos-intel.tar.gz IrisAura-macos-intel.app
tar -czf IrisAura-v1.0.0-macos-arm64.tar.gz IrisAura-macos-arm64.app

# Linux
tar -czf IrisAura-v1.0.0-linux-amd64.tar.gz IrisAura-linux-amd64
```

### 安装包制作建议

- **Windows**: 使用 NSIS 或 Inno Setup 制作安装程序
- **macOS**: 创建 DMG 安装包，建议进行代码签名
- **Linux**: 制作 AppImage、Snap 或 Flatpak 包

## 🎨 自定义配置

### 应用图标

在 `build/` 目录放置：
- `appicon.png` - 应用图标（512x512 PNG）

### 应用信息

编辑 `wails.json`：

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "IrisAura",
  "outputfilename": "IrisAura",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Leo Liu",
    "email": "leoliuabc@gmail.com"
  }
}
```

## ⚡ 优化构建

### 减小文件体积

```bash
# 去除调试信息
wails build -tags release -ldflags "-s -w"

# 使用 UPX 压缩（可选）
upx --best build/IrisAura-*.exe
```

### 添加版本信息

```bash
# 添加构建信息
wails build -ldflags "-X main.version=1.0.0 -X main.buildTime=$(date -u +%Y%m%d%H%M%S)"
```

## 🚦 常见问题

### CGO 编译错误

**问题**: `gcc: command not found` 或类似编译错误

**解决方案**:
- **Windows**: 安装 TDM-GCC 或 MinGW-w64
- **macOS**: `xcode-select --install`
- **Linux**: `sudo apt install build-essential`

### WebP 支持问题

**问题**: libwebp 相关编译错误

**解决方案**:
- **Ubuntu/Debian**: `sudo apt install libwebp-dev`
- **macOS**: `brew install webp`
- **Windows**: 按照上述 Windows 构建指南配置

### 交叉编译问题

**问题**: 交叉编译失败

**解决方案**:
```bash
# 启用 CGO 交叉编译
export CGO_ENABLED=1

# 设置交叉编译工具链（示例：Windows）
export CC=x86_64-w64-mingw32-gcc
```

### macOS 安全问题

**问题**: "IrisAura.app 已损坏，无法打开"

**解决方案**:
```bash
# M1/M2 用户需要移除扩展属性
xattr -cr ~/Downloads/IrisAura.app

# 或者在系统偏好设置中允许运行
```

## 📖 使用说明

### 基本功能

1. **选择输入文件夹**: 包含需要压缩的图片文件
2. **选择输出文件夹**: 压缩后图片的保存位置
3. **设置压缩参数**:
   - 输出格式：WebP/JPEG/PNG
   - 压缩质量：1-100
   - 最大尺寸限制
4. **开始压缩**: 支持批量处理，显示实时进度

### 支持的图片格式

**输入格式**: JPEG, PNG, WebP, BMP, TIFF, GIF  
**输出格式**: WebP, JPEG, PNG

### 特性

- ✅ 批量图片处理
- ✅ 实时压缩进度显示
- ✅ 多种输出格式支持
- ✅ 质量可调节
- ✅ 跨平台支持

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 联系方式

- 作者：Leo Liu
- 邮箱：leoliuabc@gmail.com

---

> 💡 **提示**: 首次构建可能需要较长时间下载依赖，请耐心等待。