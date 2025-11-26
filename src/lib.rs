// lib.rs
use std::ffi::c_void;
use std::ptr;
use std::slice;
use libc::{malloc, free, size_t};
use std::os::raw::c_uchar;

use lz4_flex::{compress_prepend_size, decompress_size_prepended};

#[no_mangle]
pub extern "C" fn lz4_compress(
    input_ptr: *const c_uchar,
    input_len: size_t,
    out_len: *mut u64,
) -> *mut c_uchar {
    // Safety checks
    if input_ptr.is_null() || out_len.is_null() {
        return ptr::null_mut();
    }
    if input_len == 0 {
        unsafe { *out_len = 0u64; }
        return ptr::null_mut();
    }

    // Read input slice
    let in_slice = unsafe { slice::from_raw_parts(input_ptr as *const u8, input_len as usize) };

    // compress (with size prepended so decompressor knows original size)
    let compressed = match compress_prepend_size(in_slice) {
        v => v,
    };

    let clen = compressed.len();
    if clen == 0 {
        unsafe { *out_len = 0u64; }
        return ptr::null_mut();
    }

    // allocate C heap memory so caller can free with free()
    let buf = unsafe { malloc(clen as size_t) } as *mut c_uchar;
    if buf.is_null() {
        unsafe { *out_len = 0u64; }
        return ptr::null_mut();
    }

    // copy data
    unsafe {
        ptr::copy_nonoverlapping(compressed.as_ptr(), buf as *mut u8, clen);
        *out_len = clen as u64;
    }

    buf
}

#[no_mangle]
pub extern "C" fn lz4_decompress(
    input_ptr: *const c_uchar,
    input_len: size_t,
    out_len: *mut u64,
) -> *mut c_uchar {
    if input_ptr.is_null() || out_len.is_null() {
        return ptr::null_mut();
    }
    if input_len == 0 {
        unsafe { *out_len = 0u64; }
        return ptr::null_mut();
    }

    let in_slice = unsafe { slice::from_raw_parts(input_ptr as *const u8, input_len as usize) };

    // try to decompress; lz4_flex will read prepended size
    let decompressed = match decompress_size_prepended(in_slice) {
        Ok(v) => v,
        Err(_) => {
            unsafe { *out_len = 0u64; }
            return ptr::null_mut();
        }
    };

    let dlen = decompressed.len();
    if dlen == 0 {
        unsafe { *out_len = 0u64; }
        return ptr::null_mut();
    }

    let buf = unsafe { malloc(dlen as size_t) } as *mut c_uchar;
    if buf.is_null() {
        unsafe { *out_len = 0u64; }
        return ptr::null_mut();
    }

    unsafe {
        ptr::copy_nonoverlapping(decompressed.as_ptr(), buf as *mut u8, dlen);
        *out_len = dlen as u64;
    }

    buf
}

/// 释放由上面两个函数分配的内存（使用 libc::malloc 分配）
/// Go 侧也可以直接调用 C.free，但提供一个显式函数更方便跨语言调用。
#[no_mangle]
pub extern "C" fn lz4_free(ptr: *mut c_void) {
    if ptr.is_null() {
        return;
    }
    unsafe { free(ptr) }
}
