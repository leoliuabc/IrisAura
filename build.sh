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