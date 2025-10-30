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

extern void *CGEventCreateKeyboardEvent(void *source, uint16_t virtual_key, bool key_down);

extern void CGEventSetFlags(void *event, uint64_t flags);

extern void CGEventPost(uint32_t tap, void *event);

extern void CFRelease(void *cf);

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

ClipboardHandle myreviser_clipboard_new(void);

char *myreviser_clipboard_get_text(ClipboardHandle handle);

int myreviser_clipboard_set_text(ClipboardHandle handle, const char *text);

int myreviser_clipboard_save(ClipboardHandle handle);

int myreviser_clipboard_restore(ClipboardHandle handle);

void myreviser_clipboard_free(ClipboardHandle handle);

HotkeyManagerHandle myreviser_hotkey_manager_new(void);

int myreviser_hotkey_clear(HotkeyManagerHandle handle);

int myreviser_hotkey_register(HotkeyManagerHandle handle,
                              const char *binding,
                              const char *action,
                              HotkeyCallback callback);

int myreviser_hotkey_start(HotkeyManagerHandle handle);

int myreviser_hotkey_stop(HotkeyManagerHandle handle);

void myreviser_hotkey_manager_free(HotkeyManagerHandle handle);

SimulatorHandle myreviser_simulator_new(void);

int myreviser_simulate_select_all(SimulatorHandle handle);

int myreviser_simulate_copy(SimulatorHandle handle);

int myreviser_simulate_paste(SimulatorHandle handle);

void myreviser_simulator_free(SimulatorHandle handle);

#endif  /* MYREVISER_FFI_H */
