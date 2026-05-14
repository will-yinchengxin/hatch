## ─────────────────────────────────────────────────────────────────────────────
##  Hatch — build & release commands
##  Usage: make <target>
## ─────────────────────────────────────────────────────────────────────────────

VERSION   ?= 0.1.0
PLATFORM  := darwin/arm64
WAILS     := $(shell go env GOPATH)/bin/wails
APP       := build/bin/Hatch.app
DMG       := build/bin/Hatch-$(VERSION)-arm64.dmg
CREATE_DMG := /opt/homebrew/bin/create-dmg

.PHONY: help dev build test release sign dmg install clean

## help — 显示所有可用命令（默认目标）
help:
	@awk 'BEGIN{FS="[ \t]+—[ \t]+"} /^## /{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST) \
	  | sed 's/^  ## //'

## ─── 开发 ────────────────────────────────────────────────────────────────────

## dev      — 启动热更新开发模式（自动打开窗口）
dev:
	$(WAILS) dev

## ─── 测试 ────────────────────────────────────────────────────────────────────

## test     — 运行所有单元测试
test:
	go test ./internal/... -v

## test-short — 只显示 PASS/FAIL，不输出详情
test-short:
	go test ./internal/...

## ─── 构建 ────────────────────────────────────────────────────────────────────

## build    — 编译 darwin/arm64（增量，不清理旧产物）
build:
	$(WAILS) build -platform $(PLATFORM)

## build-clean — 完整重新编译（清理旧产物后再构建）
build-clean:
	$(WAILS) build -platform $(PLATFORM) -clean

## ─── 签名 & 打包 ─────────────────────────────────────────────────────────────

## sign     — 对 Hatch.app 做 ad-hoc 自签名（不需要 Apple 开发者账号）
sign: $(APP)
	codesign --force --deep --sign - $(APP)
	codesign --verify --deep --strict $(APP)
	@echo "✓ codesign OK"

## dmg      — 打 DMG 安装包（依赖 sign）
dmg: sign
	@[ -f $(DMG) ] && rm $(DMG) || true
	$(CREATE_DMG) \
	  --volname "Hatch" \
	  --volicon "build/appicon.png" \
	  --window-pos 200 120 \
	  --window-size 600 400 \
	  --icon-size 100 \
	  --icon "Hatch.app" 160 185 \
	  --hide-extension "Hatch.app" \
	  --app-drop-link 440 185 \
	  $(DMG) \
	  $(APP)
	@echo "✓ $(shell ls -lh $(DMG) | awk '{print $$5, $$NF}')"

## release  — 完整发布：build-clean → sign → dmg
release: build-clean sign dmg

## ─── 安装 & 运行 ─────────────────────────────────────────────────────────────

## install  — 将 Hatch.app 复制到 /Applications（需要管理员权限）
install: $(APP)
	cp -r $(APP) /Applications/Hatch.app
	xattr -cr /Applications/Hatch.app
	@echo "✓ 已安装到 /Applications/Hatch.app"

## open     — 直接运行当前构建的 Hatch.app（绕过 Gatekeeper）
open: $(APP)
	xattr -cr $(APP)
	open $(APP)

## ─── 清理 ────────────────────────────────────────────────────────────────────

## clean    — 删除所有构建产物
clean:
	rm -rf build/bin

## ─────────────────────────────────────────────────────────────────────────────

$(APP): ;
	@echo "Hatch.app not found. Run 'make build' first." && exit 1
