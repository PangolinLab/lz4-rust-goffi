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
	"runtime"
	"unsafe"
)

func init() {
	var libPath string
	switch runtime.GOOS {
	case "windows":
		libPath = "bin/lz4.dll"
	case "darwin":
		libPath = "bin/liblz4.dylib"
	default:
		libPath = "bin/liblz4.so"
	}

	// 如果库不存在，则自动编译 Rust
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		// Rust 源码目录相对路径
		rustDir := "../" // 从 aes_256_gcm_siv_ffi 到 src 的上级目录
		cmd := exec.Command("cargo", "build", "--release", "--target-dir", "aes_256_gcm_siv_ffi/bin")
		cmd.Dir = rustDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
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
