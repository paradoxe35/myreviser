//go:build windows
// +build windows

package input

import "golang.design/x/hotkey"

// parseModifier parses a modifier string on Windows
func parseModifier(mod string) (hotkey.Modifier, bool) {
	switch mod {
	case "ctrl", "control":
		return hotkey.ModCtrl, true
	case "alt":
		return hotkey.ModAlt, true
	case "shift":
		return hotkey.ModShift, true
	case "win", "windows", "super", "cmd", "command":
		return hotkey.ModWin, true
	default:
		return 0, false
	}
}
