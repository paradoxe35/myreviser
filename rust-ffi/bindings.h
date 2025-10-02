#ifndef MYREVISER_FFI_H
#define MYREVISER_FFI_H

#pragma once

#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

typedef void *ClipboardHandle;

/**
 * Opaque pointer types for safe cross-FFI boundary object passing
 */
typedef void *HotkeyManagerHandle;

/**
 * Hotkey callback function type
 * The callback receives the action string that was registered
 */
typedef void (*HotkeyCallback)(const char*);

typedef void *SimulatorHandle;

/**
 * Get the last error message
 * Returns: C string (must be freed with myreviser_free_string) or NULL if no error
 */
const char *myreviser_get_last_error(void);

/**
 * Free a string allocated by Rust
 * This must be called for all strings returned by Rust functions
 */
void myreviser_free_string(char *s);

/**
 * Create a new clipboard manager
 * Returns: Opaque handle to clipboard manager or NULL on failure
 */
ClipboardHandle myreviser_clipboard_new(void);

/**
 * Get text from clipboard
 * Returns: C string (must be freed with myreviser_free_string) or NULL on failure
 */
char *myreviser_clipboard_get_text(ClipboardHandle handle);

/**
 * Set text to clipboard
 * Returns: 0 on success, error code on failure
 */
int myreviser_clipboard_set_text(ClipboardHandle handle, const char *text);

/**
 * Save current clipboard content
 * Returns: 0 on success, error code on failure
 */
int myreviser_clipboard_save(ClipboardHandle handle);

/**
 * Restore previously saved clipboard content
 * Returns: 0 on success, error code on failure
 */
int myreviser_clipboard_restore(ClipboardHandle handle);

/**
 * Free clipboard manager resources
 */
void myreviser_clipboard_free(ClipboardHandle handle);

/**
 * Create a new hotkey manager
 * Returns: Opaque handle to hotkey manager or NULL on failure
 */
HotkeyManagerHandle myreviser_hotkey_manager_new(void);

/**
 * Register a hotkey with callback
 * binding: Hotkey string like "ctrl+alt+space"
 * action: Action identifier (passed to callback)
 * callback: Function to call when hotkey is pressed
 * Returns: 0 on success, error code on failure
 */
int myreviser_hotkey_register(HotkeyManagerHandle handle,
                              const char *binding,
                              const char *action,
                              HotkeyCallback callback);

/**
 * Start listening for hotkeys
 * Returns: 0 on success, error code on failure
 */
int myreviser_hotkey_start(HotkeyManagerHandle handle);

/**
 * Stop listening for hotkeys
 * Returns: 0 on success, error code on failure
 */
int myreviser_hotkey_stop(HotkeyManagerHandle handle);

/**
 * Free hotkey manager resources
 */
void myreviser_hotkey_manager_free(HotkeyManagerHandle handle);

/**
 * Create a new key simulator
 * Returns: Opaque handle to key simulator or NULL on failure
 */
SimulatorHandle myreviser_simulator_new(void);

/**
 * Simulate select all (Ctrl+A or Cmd+A on macOS)
 * Returns: 0 on success, error code on failure
 */
int myreviser_simulate_select_all(SimulatorHandle handle);

/**
 * Simulate copy (Ctrl+C or Cmd+C on macOS)
 * Returns: 0 on success, error code on failure
 */
int myreviser_simulate_copy(SimulatorHandle handle);

/**
 * Simulate paste (Ctrl+V or Cmd+V on macOS)
 * Returns: 0 on success, error code on failure
 */
int myreviser_simulate_paste(SimulatorHandle handle);

/**
 * Free key simulator resources
 */
void myreviser_simulator_free(SimulatorHandle handle);

#endif  /* MYREVISER_FFI_H */
