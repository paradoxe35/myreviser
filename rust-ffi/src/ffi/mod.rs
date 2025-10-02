pub mod ffi_types;
pub mod ffi_clipboard;
pub mod ffi_simulator;
pub mod ffi_hotkey;

use std::os::raw::c_char;
pub use ffi_types::*;

/// Get the last error message
/// Returns: C string (must be freed with myreviser_free_string) or NULL if no error
#[no_mangle]
pub unsafe extern "C" fn myreviser_get_last_error() -> *const c_char {
    match take_last_error() {
        Some(err) => string_to_c_str(err),
        None => std::ptr::null(),
    }
}

/// Free a string allocated by Rust
/// This must be called for all strings returned by Rust functions
#[no_mangle]
pub unsafe extern "C" fn myreviser_free_string(s: *mut c_char) {
    if !s.is_null() {
        let _ = std::ffi::CString::from_raw(s);
    }
}
