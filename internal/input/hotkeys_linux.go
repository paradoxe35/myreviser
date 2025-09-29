// +build linux

package input

import "golang.design/x/hotkey"

// parseModifier parses a modifier string on Linux
func parseModifier(mod string) (hotkey.Modifier, bool) {
	switch mod {
	case "ctrl", "control":
		return hotkey.ModCtrl, true
	case "alt":
		return hotkey.Mod1, true // Alt on X11
	case "shift":
		return hotkey.ModShift, true
	case "super", "win", "windows", "cmd", "command":
		return hotkey.Mod4, true // Super/Windows key on X11
	default:
		return 0, false
	}
}