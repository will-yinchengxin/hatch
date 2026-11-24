//go:build !cgo

package hostmonitor

import "errors"

const nativeStorageTrashAvailable = false

func trashStorageCache(string, int) (string, error) {
	return "", errors.New("当前构建缺少 macOS 废纸篓支持")
}
