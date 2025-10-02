# Memory Leak Analysis Report

**Date**: 2025-10-02
**Project**: MyReviser (Go + Rust FFI)
**Analysis Focus**: Memory management across FFI boundary

## Executive Summary

✅ **No critical memory leaks detected**
⚠️ **1 potential issue found** (hotkey callback CString lifetime)
✅ **Proper cleanup implemented** throughout the codebase

---

## 1. Rust FFI Layer Analysis

### 1.1 Resource Allocation & Cleanup

#### ✅ Clipboard Manager (`ffi_clipboard.rs`)
**Allocation**: Line 13
```rust
Box::into_raw(Box::new(clipboard)) as ClipboardHandle
```

**Cleanup**: Line 164
```rust
let _ = Box::from_raw(handle as *mut ClipboardManager);
```

**Status**: ✅ **SAFE** - Proper Box allocation and deallocation

---

#### ✅ Key Simulator (`ffi_simulator.rs`)
**Allocation**: Line 13
```rust
Box::into_raw(Box::new(simulator)) as SimulatorHandle
```

**Cleanup**: Line 85
```rust
let _ = Box::from_raw(handle as *mut KeySimulator);
```

**Status**: ✅ **SAFE** - Proper Box allocation and deallocation

---

#### ✅ Hotkey Manager (`ffi_hotkey.rs`)
**Allocation**: Line 257-258
```rust
let manager = Box::new(SimpleHotkeyManager::new());
Box::into_raw(manager) as HotkeyManagerHandle
```

**Cleanup**: Line 354
```rust
let _ = Box::from_raw(handle as *mut SimpleHotkeyManager);
```

**Status**: ✅ **SAFE** - Proper Box allocation and deallocation

---

### 1.2 String Memory Management

#### ✅ C String Returns (`ffi_clipboard.rs` Line 45)
**Allocation**:
```rust
string_to_c_str(text) // Returns CString::into_raw()
```

**Cleanup** (Go side):
```go
defer C.myreviser_free_string(cStr)  // ffi_clipboard.go:53
```

**Rust Free Function** (`ffi/mod.rs`):
```rust
pub unsafe extern "C" fn myreviser_free_string(s: *mut c_char) {
    if !s.is_null() {
        let _ = CString::from_raw(s);
    }
}
```

**Status**: ✅ **SAFE** - Strings are properly freed on Go side

---

#### ⚠️ Callback CString Lifetime Issue (`ffi_hotkey.rs` Line 121-123)

**Current Code**:
```rust
let action_cstr = std::ffi::CString::new(binding.action.clone()).unwrap();
(binding.callback)(action_cstr.as_ptr());
// ⚠️ action_cstr dropped here - pointer may be invalid in callback
```

**Issue**: The `CString` is created and immediately dropped after passing the pointer to the callback. If the Go side doesn't immediately copy the string, it could access freed memory.

**Current Go Side** (`ffi_hotkeys.go` Line 220):
```go
actionStr := C.GoString(action)  // ✅ Immediately copies
```

**Status**: ⚠️ **CURRENTLY SAFE** but fragile. The Go side immediately copies the string with `C.GoString()`, which prevents the issue. However, this is a **dangerous pattern** that could break if the Go code changes.

**Recommendation**: See Section 5.1

---

### 1.3 Tokio Runtime Creation

#### ✅ Temporary Runtimes (`ffi_clipboard.rs`)

Multiple functions create temporary Tokio runtimes:
- Line 33-42: `myreviser_clipboard_get_text`
- Line 80-89: `myreviser_clipboard_set_text`
- Line 110-119: `myreviser_clipboard_save`
- Line 140-149: `myreviser_clipboard_restore`

**Pattern**:
```rust
let rt = tokio::runtime::Builder::new_current_thread()
    .enable_all()
    .build()?;
match rt.block_on(clipboard.get_text()) { ... }
// Runtime dropped here
```

**Status**: ✅ **SAFE** - Runtimes are properly dropped when they go out of scope. This is inefficient but not a memory leak.

**Note**: Creating a runtime per call has overhead but doesn't leak memory.

---

## 2. Go CGO Layer Analysis

### 2.1 C String Management

#### ✅ Input Strings (Go → Rust)
**Pattern** (`ffi_clipboard.go` Line 64-65):
```go
cText := C.CString(text)
defer C.free(unsafe.Pointer(cText))
```

**Status**: ✅ **SAFE** - All `C.CString()` calls have `defer C.free()`

**Verified Locations**:
- `ffi_clipboard.go:64-65` (SetText)
- `ffi_hotkeys.go:99-102` (RegisterHotkey - both binding and action)

---

#### ✅ Output Strings (Rust → Go)
**Pattern** (`ffi_clipboard.go` Line 49-53):
```go
cStr := C.myreviser_clipboard_get_text(c.handle)
if cStr == nil {
    return "", fmt.Errorf(...)
}
defer C.myreviser_free_string(cStr)
return C.GoString(cStr), nil
```

**Status**: ✅ **SAFE** - All Rust-allocated strings are freed with `myreviser_free_string`

**Verified Locations**:
- `ffi_clipboard.go:49-55` (GetText)
- `ffi_clipboard.go:184-190` (getLastError)

---

### 2.2 Handle Lifecycle

#### ✅ FFIClipboardManager
**Creation**: `ffi_clipboard.go:34-40`
```go
handle := C.myreviser_clipboard_new()
return &FFIClipboardManager{handle: handle}, nil
```

**Cleanup**: `ffi_clipboard.go:104-109`
```go
func (c *FFIClipboardManager) Close() {
    if c.handle != nil {
        C.myreviser_clipboard_free(c.handle)
        c.handle = nil  // Prevents double-free
    }
}
```

**Status**: ✅ **SAFE** - Proper nil check and handle clearing

---

#### ✅ FFIKeySimulator
**Creation**: `ffi_simulator.go:35-41`
**Cleanup**: `ffi_simulator.go:90-95`

**Status**: ✅ **SAFE** - Same pattern as clipboard manager

---

#### ✅ FFIHotkeyManager
**Creation**: `ffi_hotkeys.go:48-65`
**Cleanup**: `ffi_hotkeys.go:180-194`

**Special Note**: Also clears global reference
```go
globalFFIMu.Lock()
if globalFFIHotkeyManager == h {
    globalFFIHotkeyManager = nil
}
globalFFIMu.Unlock()
```

**Status**: ✅ **SAFE** - Proper cleanup including global reference

---

## 3. Application Lifecycle Analysis

### 3.1 Cleanup Chain

#### ✅ Application Shutdown (`app.go:130-147`)
```go
func (a *Application) Stop() {
    // 1. Stop hotkey manager
    if a.hotkeyManager != nil {
        a.hotkeyManager.Stop()    // Stops listener thread
        a.hotkeyManager.Close()   // Frees FFI resources
    }

    // 2. Close processor (which closes clipboard manager)
    if a.processor != nil {
        a.processor.Close()
    }

    // 3. Quit app
    a.app.Quit()
}
```

**Status**: ✅ **SAFE** - Proper cleanup order

---

#### ✅ Processor Cleanup (`revision/processor.go:274-279`)
```go
func (p *Processor) Close() {
    if p.clipboardManager != nil {
        p.clipboardManager.Close()
    }
}
```

**Status**: ✅ **SAFE** - Closes underlying clipboard manager

---

### 3.2 Temporary Resources

#### ✅ Simulator in CaptureSelectedText (`ffi_clipboard.go:119-123`)
```go
sim, err := NewFFIKeySimulator()
if err != nil {
    return "", fmt.Errorf(...)
}
defer sim.Close()  // ✅ Always cleaned up
```

**Status**: ✅ **SAFE** - Uses defer for guaranteed cleanup

---

#### ✅ Simulator in ReplaceSelectedText (`ffi_clipboard.go:160-164`)
Same pattern as CaptureSelectedText

**Status**: ✅ **SAFE**

---

## 4. Thread Safety Analysis

### 4.1 Hotkey Manager Thread

#### ⚠️ Listener Thread Cleanup (`ffi_hotkey.rs:152-160`)
```rust
pub fn stop(&mut self) -> Result<(), String> {
    let mut active = self.active.lock();
    *active = false;
    drop(active);

    // Note: rdev doesn't provide a way to cleanly stop listening
    // The thread will exit when the program exits
    Ok(())
}
```

**Issue**: The listener thread (line 86-146) runs in an infinite loop with `rdev::listen()`. Setting `active = false` will prevent hotkey execution, but the thread continues running until process exit.

**Impact**:
- Thread continues consuming resources until program termination
- Not a memory leak per se, but a resource leak
- The thread **will** be cleaned up when the process exits

**Status**: ⚠️ **ACCEPTABLE** but not ideal. This is a limitation of the `rdev` library which doesn't provide a stop mechanism.

**Recommendation**: Document this behavior

---

### 4.2 Global Hotkey Manager Reference

#### ✅ Global State (`ffi_hotkeys.go:44-45`)
```go
var globalFFIHotkeyManager *FFIHotkeyManager
var globalFFIMu sync.Mutex
```

**Cleanup** (`ffi_hotkeys.go:188-193`):
```go
globalFFIMu.Lock()
if globalFFIHotkeyManager == h {
    globalFFIHotkeyManager = nil
}
globalFFIMu.Unlock()
```

**Status**: ✅ **SAFE** - Properly cleared on Close()

---

## 5. Recommendations

### 5.1 PRIORITY: Fix Callback CString Lifetime

**Current Issue** (`ffi_hotkey.rs:121-123`):
```rust
let action_cstr = std::ffi::CString::new(binding.action.clone()).unwrap();
(binding.callback)(action_cstr.as_ptr());
// ⚠️ action_cstr dropped here
```

**Recommended Fix**:
```rust
// Option 1: Leak the string and let Go free it (safest)
let action_cstr = std::ffi::CString::new(binding.action.clone()).unwrap();
let action_ptr = action_cstr.into_raw();
(binding.callback)(action_ptr);
// Go side must free with myreviser_free_string

// Option 2: Document that Go MUST immediately copy (current approach)
// Add documentation and safety comment
```

**Update Go Side** (`ffi_hotkeys.go:210`):
```go
//export hotkeyCallbackGateway
func hotkeyCallbackGateway(action *C.char) {
    // IMPORTANT: Immediately copy the string as it may be freed after this function
    actionStr := C.GoString(action)  // ✅ Already doing this

    // If we change to Option 1, add:
    // defer C.myreviser_free_string(action)

    // ... rest of function
}
```

---

### 5.2 Consider Reusable Tokio Runtime

**Current**: Creating runtime per clipboard operation
**Impact**: Performance overhead, not a memory leak
**Recommendation**:
```rust
// Store runtime in ClipboardManager
pub struct ClipboardManager {
    clipboard: Arc<Mutex<Clipboard>>,
    original_content: Arc<Mutex<Option<String>>>,
    runtime: Runtime,  // Add this
}
```

**Benefit**: Better performance, same memory safety

---

### 5.3 Document Thread Behavior

**Add to documentation**:
```markdown
## Known Limitations

### Hotkey Listener Thread
The hotkey listener thread (using `rdev::listen`) cannot be cleanly stopped
due to library limitations. The thread will continue running (in a paused
state) until program termination. This is expected behavior and does not
cause memory leaks, though it does keep the thread alive.

To minimize resource usage, call `Stop()` before `Close()` to prevent
hotkey execution, even though the listener thread remains active.
```

---

### 5.4 Add Memory Leak Tests

**Create test file**: `rust-ffi/tests/memory_leak_test.rs`
```rust
#[test]
fn test_no_memory_leak_clipboard() {
    unsafe {
        for _ in 0..1000 {
            let handle = myreviser_clipboard_new();
            assert!(!handle.is_null());
            myreviser_clipboard_free(handle);
        }
    }
    // If this doesn't OOM, we're good
}

#[test]
fn test_no_memory_leak_simulator() {
    // Similar test for simulator
}

#[test]
fn test_no_memory_leak_hotkeys() {
    // Similar test for hotkeys
}
```

---

## 6. Summary

### ✅ What's Working Well

1. **FFI Handle Management**: All Rust resources properly allocated/freed via Box
2. **String Management**: C strings properly freed on both sides
3. **Application Lifecycle**: Proper cleanup chain from app → processor → managers
4. **Temporary Resources**: defer used correctly for cleanup
5. **Double-Free Prevention**: Handles set to nil after freeing
6. **Thread Safety**: Proper mutex usage for shared state

---

### ⚠️ Areas for Improvement

1. **Callback CString Lifetime** (Priority: HIGH)
   - Currently works but fragile
   - Should be fixed proactively

2. **Hotkey Thread Cleanup** (Priority: LOW)
   - Thread stays alive until process exit
   - Not a leak, but not optimal
   - Limitation of `rdev` library

3. **Tokio Runtime Creation** (Priority: LOW - Performance)
   - Creates runtime per operation
   - Consider using shared runtime

---

### 🎯 Action Items

1. **CRITICAL**: Fix callback CString lifetime in `ffi_hotkey.rs:121-123`
2. **RECOMMENDED**: Add memory leak tests
3. **NICE TO HAVE**: Optimize Tokio runtime usage
4. **DOCUMENTATION**: Document thread behavior limitation

---

## 7. Memory Safety Checklist

| Category | Item | Status |
|----------|------|--------|
| **Rust FFI** | Box allocation/deallocation | ✅ SAFE |
| **Rust FFI** | String memory management | ⚠️ FIX NEEDED |
| **Rust FFI** | Tokio runtime cleanup | ✅ SAFE |
| **Go CGO** | C.CString + defer C.free | ✅ SAFE |
| **Go CGO** | Rust string + myreviser_free_string | ✅ SAFE |
| **Go CGO** | Handle lifecycle | ✅ SAFE |
| **Application** | Cleanup chain | ✅ SAFE |
| **Application** | Temporary resources | ✅ SAFE |
| **Threading** | Listener thread | ⚠️ DOCUMENTED |
| **Threading** | Global state cleanup | ✅ SAFE |

---

## 8. Conclusion

The codebase demonstrates **excellent memory management practices** overall. There is **one critical issue** that should be fixed (callback CString lifetime), though it currently works due to immediate string copying on the Go side.

**Risk Assessment**:
- **Current Risk**: LOW (code works correctly as-is)
- **Future Risk**: MEDIUM (fragile pattern that could break with changes)
- **Recommended Action**: Implement fix in Section 5.1 to eliminate future risk

**Overall Grade**: A- (would be A+ with callback fix)
