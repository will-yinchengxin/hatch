//go:build cgo

package hostmonitor

/*
#cgo LDFLAGS: -framework Foundation
#include <stdlib.h>
int hatchTrashStorageCache(const char *path, int fd, char **result);
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

const nativeStorageTrashAvailable = true

func trashStorageCache(path string, fd int) (string, error) {
	raw := C.CString(path)
	defer C.free(unsafe.Pointer(raw))
	var result *C.char
	code := C.hatchTrashStorageCache(raw, C.int(fd), &result)
	defer C.free(unsafe.Pointer(result))
	switch code {
	case 0:
		if result != nil {
			return C.GoString(result), nil
		}
		return "", errStorageOutcomeUncertain
	case 1:
		return "", errors.New("项目路径在移入废纸篓前发生变化")
	default:
		return "", fmt.Errorf("macOS 无法将项目移到废纸篓：%s", C.GoString(result))
	}
}
