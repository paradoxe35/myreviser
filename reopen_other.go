//go:build !darwin

package main

// Only macOS reopens a running app instead of starting another process.
func installReopenHandler(show func()) {}
