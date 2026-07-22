package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
)

func main() {

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("[-] Connection failed: %v", err)
	}
	defer X.Conn().Close()

	keybind.Initialize(X)
	root := X.RootWin()

	//Setting up signal handler before grabbing the keyboard
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[!] Cntl + C detected! Releasing keyboard grab...")
		xproto.UngrabKeyboard(X.Conn(), xproto.TimeCurrentTime)
		X.Conn().Sync()
		X.Conn().Close()
		fmt.Println("[+] Released! Exiting cleanly.")
		os.Exit(0)
	}()

	fmt.Println("[+] Requesting global keyboard grab via XGrabKeyboard...")

	reply, err := xproto.GrabKeyboard(
		X.Conn(),
		false,
		root,
		xproto.TimeCurrentTime,
		xproto.GrabModeAsync,
		xproto.GrabModeAsync,
	).Reply()

	if err != nil {
		log.Fatalf("[-] XGrabKeyboard RPC failed: %v", err)
	}

	if reply.Status != xproto.GrabStatusSuccess {
		log.Fatalf("[-] Failed to acquire keyboard grab. Status code: %d (Another app likely holds a grab)", reply.Status)
	}

	fmt.Println("[+] Keyboard grabbed globally! Type anywhere - all keys belong to us.")
	// fmt.Println("[+] Press Cntl + C in this terminal to release the grab.")

	// // Release on cntl + x
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	<-c
	// 	fmt.Println("\n[+] Releasing keyboard grab and exiting...")
	// 	xproto.UngrabKeyboard(X.Conn(), xproto.TimeCurrentTime)
	// 	X.Conn().Close()
	// 	os.Exit(0)
	// }()

	//Attach event handler for keypress events on root window
	ctrlPressed := false
	xevent.KeyPressFun(
		func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
			keyStr := keybind.LookupString(X, e.State, e.Detail)
			if keyStr == "" {
				keyStr = fmt.Sprintf("KeyCode(%d)", e.Detail)
			}
			fmt.Printf("[GRABBED KEY] Code: %-3d | Char: %s\n", e.Detail, keyStr)

			if e.Detail == 37 || e.Detail == 105 {
				ctrlPressed = true
			}

			isEsc := e.Detail == 9
			isQuitKey := e.Detail == 24 // 'q'
			isCtrlC := ctrlPressed && (e.Detail == 54)
			if isEsc || isQuitKey || isCtrlC {
				fmt.Println("\n[!] Exit trigger detected! Releasing keyboard grab...")

				// Explicitly ungrab keyboard
				xproto.UngrabKeyboard(X.Conn(), xproto.TimeCurrentTime)
				X.Conn().Sync()
				X.Conn().Close()

				fmt.Println("[+] Grab released cleanly. Exiting.")
				xevent.Quit(X)
			}
		},
	).Connect(X, root)

	xevent.Main(X)
}
