//! MyReviser FFI Library
//!
//! This library provides C-compatible FFI bindings for system input functionality
//! (hotkeys, clipboard, keyboard simulation) to be used from Go via CGO.
//!
//! # Safety
//!
//! All public FFI functions are marked as `unsafe` because they cross the FFI boundary.
//! Callers must ensure:
//! - Pointers are valid and non-null (unless explicitly allowed)
//! - Strings are valid UTF-8 null-terminated C strings
//! - Handles are not used after being freed
//! - Callbacks remain valid for the lifetime of the hotkey manager

pub mod core;
pub mod ffi;

// Re-export FFI functions for C bindings generation
pub use ffi::*;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_library_smoke() {
        // Basic smoke test to ensure library compiles
        assert!(true);
    }
}
