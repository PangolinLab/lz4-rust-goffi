// lz4_interface.h
#ifndef LZ4_INTERFACE_H
#define LZ4_INTERFACE_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// 压缩
// input_ptr: 指向输入数据
// input_len: 输入字节长度
// out_len: 输出长度（由函数填充）
// 返回: 指向 malloc 分配的 buffer，失败返回 NULL。调用方必须 free() 或 lz4_free().
unsigned char* lz4_compress(const unsigned char* input_ptr, size_t input_len, uint64_t* out_len);

// 解压
unsigned char* lz4_decompress(const unsigned char* input_ptr, size_t input_len, uint64_t* out_len);

// 释放由上面函数分配的内存
void lz4_free(void* ptr);

#ifdef __cplusplus
}
#endif

#endif // LZ4_INTERFACE_H
