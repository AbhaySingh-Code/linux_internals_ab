package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
)

func main() {
	// 1. Connect to the X server ($DISPLAY)
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("[-] Failed to connect to X server: %v", err)
	}
	defer X.Conn().Close()

	// Initialize keycode <-> keysym mapping tables (needed by keybind)
	keybind.Initialize(X)

	// 2. Define the hotkey you want to listen for.
	// Modifier string uses xgbutil's format: "Mod1" = Alt, "Control", "Shift", "Mod4" = Super/Win
	// Change this to whatever combo you want, e.g. "Shift-F5", "Mod4-space", etc.
	hotkey := "Control-Mod1-k" // Ctrl+Alt+K

	root := X.RootWin()

	// 3. Grab ONLY this specific key combination on the root window.
	// This does NOT capture other keystrokes - the X server only sends
	// an event when this exact key+modifier combo is pressed.
	err = keybind.KeyPressFun(
		func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
			fmt.Println("[+] Hotkey triggered: Ctrl+Alt+K pressed")
			// Put whatever action you want triggered here.
		}).Connect(X, root, hotkey, true)
	if err != nil {
		log.Fatalf("[-] Failed to grab hotkey %q: %v", hotkey, err)
	}

	fmt.Printf("[+] Listening for hotkey: %s (Ctrl+C to exit)\n", hotkey)

	// 4. Run the event loop. Only delivers events for grabbed keys/windows.
	xevent.Main(X)
}
