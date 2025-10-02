use std::os::raw::c_int;

use super::ffi_types::*;
use crate::core::KeySimulator;

/// Create a new key simulator
/// Returns: Opaque handle to key simulator or NULL on failure
#[no_mangle]
pub unsafe extern "C" fn myreviser_simulator_new() -> SimulatorHandle {
    init_logging();

    match KeySimulator::new() {
        Ok(simulator) => Box::into_raw(Box::new(simulator)) as SimulatorHandle,
        Err(e) => {
            set_last_error(format!("Failed to create key simulator: {:#}", e));
            std::ptr::null_mut()
        }
    }
}

/// Simulate select all (Ctrl+A or Cmd+A on macOS)
/// Returns: 0 on success, error code on failure
#[no_mangle]
pub unsafe extern "C" fn myreviser_simulate_select_all(handle: SimulatorHandle) -> c_int {
    if handle.is_null() {
        set_last_error("Null simulator handle provided".to_string());
        return FFIErrorCode::NullPointer as c_int;
    }

    let simulator = &mut *(handle as *mut KeySimulator);

    match simulator.select_all() {
        Ok(_) => FFIErrorCode::Success as c_int,
        Err(e) => {
            set_last_error(format!("Select all simulation failed: {:#}", e));
            FFIErrorCode::OperationFailed as c_int
        }
    }
}

/// Simulate copy (Ctrl+C or Cmd+C on macOS)
/// Returns: 0 on success, error code on failure
#[no_mangle]
pub unsafe extern "C" fn myreviser_simulate_copy(handle: SimulatorHandle) -> c_int {
    if handle.is_null() {
        set_last_error("Null simulator handle provided".to_string());
        return FFIErrorCode::NullPointer as c_int;
    }

    let simulator = &mut *(handle as *mut KeySimulator);

    match simulator.copy() {
        Ok(_) => FFIErrorCode::Success as c_int,
        Err(e) => {
            set_last_error(format!("Copy simulation failed: {:#}", e));
            FFIErrorCode::OperationFailed as c_int
        }
    }
}

/// Simulate paste (Ctrl+V or Cmd+V on macOS)
/// Returns: 0 on success, error code on failure
#[no_mangle]
pub unsafe extern "C" fn myreviser_simulate_paste(handle: SimulatorHandle) -> c_int {
    if handle.is_null() {
        set_last_error("Null simulator handle provided".to_string());
        return FFIErrorCode::NullPointer as c_int;
    }

    let simulator = &mut *(handle as *mut KeySimulator);

    match simulator.paste() {
        Ok(_) => FFIErrorCode::Success as c_int,
        Err(e) => {
            set_last_error(format!("Paste simulation failed: {:#}", e));
            FFIErrorCode::OperationFailed as c_int
        }
    }
}

/// Free key simulator resources
#[no_mangle]
pub unsafe extern "C" fn myreviser_simulator_free(handle: SimulatorHandle) {
    if !handle.is_null() {
        let _ = Box::from_raw(handle as *mut KeySimulator);
    }
}
