//go:build darwin
// +build darwin

package input

import "golang.design/x/hotkey"

// parseModifier parses a modifier string on macOS
func parseModifier(mod string) (hotkey.Modifier, bool) {
	switch mod {
	case "ctrl", "control":
		return hotkey.ModCtrl, true
	case "alt", "option":
		return hotkey.ModOption, true
	case "shift":
		return hotkey.ModShift, true
	case "cmd", "command", "super":
		return hotkey.ModCmd, true
	default:
		return 0, false
	}
}
