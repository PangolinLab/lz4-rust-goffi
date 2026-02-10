//go:build cgo
// +build cgo

package lz4ffi

/*
	#cgo CFLAGS: -I${SRCDIR}/include
	#cgo LDFLAGS: -lkernel32 -lntdll -luserenv -lws2_32 -ldbghelp -L${SRCDIR}/bin -llz4
	#include <stdlib.h>
	#include <lz4_interface.h>
*/
import "C"
import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"unsafe"
)

func init() {
	// 动态库最终路径
	var libFile string
	switch runtime.GOOS {
	case "windows":
		libFile = "bin/lz4.dll"
	case "darwin":
		libFile = "bin/lilz4.dylib"
	default:
		libFile = "bin/liblz4.so"
	}

	// 如果库不存在，则编译 Rust 并复制到 bin/
	if _, err := os.Stat(libFile); os.IsNotExist(err) {
		// Rust 源码目录（Cargo.toml 所在目录）
		rustDir := "../" // 根据你的目录结构调整
		buildCmd := exec.Command("cargo", "build", "--release")
		buildCmd.Dir = rustDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			panic("Failed to build Rust library: " + err.Error())
		}

		// 源文件路径（默认 target/release/）
		var srcLib string
		switch runtime.GOOS {
		case "windows":
			srcLib = filepath.Join(rustDir, "target", "release", "lz4.dll")
		case "darwin":
			srcLib = filepath.Join(rustDir, "target", "release", "liblz4.dylib")
		default:
			srcLib = filepath.Join(rustDir, "target", "release", "liblz4.so")
		}

		// 确保 bin 目录存在
		_ = os.MkdirAll("bin", 0755)

		// 复制库到 bin/
		input, err := os.ReadFile(srcLib)
		if err != nil {
			panic("Failed to read Rust library: " + err.Error())
		}
		if err := os.WriteFile(libFile, input, 0644); err != nil {
			panic("Failed to write library to bin/: " + err.Error())
		}
	}
}

// Compress 压缩数据，返回压缩后的字节切片（调用者负责复制/使用后释放）
func Compress(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, errors.New("lz4 compress: empty input")
	}

	// 调用 Rust 函数
	var outLen C.uint64_t
	ptr := C.lz4_compress((*C.uchar)(unsafe.Pointer(&src[0])), C.size_t(len(src)), &outLen)
	if ptr == nil || outLen == 0 {
		return nil, errors.New("lz4 compress failed")
	}
	// 把 C 内存拷贝到 Go 堆上，然后释放 C 内存
	out := C.GoBytes(unsafe.Pointer(ptr), C.int(outLen))
	C.lz4_free(unsafe.Pointer(ptr))
	return out, nil
}

// Decompress 解压缩
func Decompress(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, errors.New("lz4 decompress: empty input")
	}
	var outLen C.uint64_t
	ptr := C.lz4_decompress((*C.uchar)(unsafe.Pointer(&src[0])), C.size_t(len(src)), &outLen)
	if ptr == nil || outLen == 0 {
		return nil, errors.New("lz4 decompress failed")
	}
	out := C.GoBytes(unsafe.Pointer(ptr), C.int(outLen))
	C.lz4_free(unsafe.Pointer(ptr))
	return out, nil
}
