package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xfixes"
	"github.com/BurntSushi/xgb/xproto"
)

// Helper to intern an X11 Atom string identifier
func getAtom(X *xgb.Conn, name string) xproto.Atom {
	reply, err := xproto.InternAtom(X, false, uint16(len(name)), name).Reply()
	if err != nil {
		log.Fatalf("[-] Failed to intern atom %s: %v", name, err)
	}
	return reply.Atom
}

// Extract converted string from selection owner
func fetchClipboardText(X *xgb.Conn, root xproto.Window, clipboardAtom, utf8Atom xproto.Atom) (string, error) {
	propertyAtom := getAtom(X, "XSEL_DATA")

	winID, err := xproto.NewWindowId(X)
	if err != nil {
		return "", err
	}

	screen := xproto.Setup(X).DefaultScreen(X)
	xproto.CreateWindow(
		X,
		screen.RootDepth,
		winID,
		root,
		0, 0, 1, 1, 0,
		xproto.WindowClassInputOutput,
		screen.RootVisual,
		xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange},
	)
	defer xproto.DestroyWindow(X, winID)

	// Initiate transfer protocol
	xproto.ConvertSelection(
		X,
		winID,
		clipboardAtom,
		utf8Atom,
		propertyAtom,
		xproto.TimeCurrentTime,
	)

	// Wait for SelectionNotify
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for SelectionNotify")
		default:
			ev, err := X.PollForEvent()
			if err != nil {
				return "", err
			}
			if ev == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// Check for SelectionNotify (Opcode 31)
			if ev.Bytes()[0]&0x7f == xproto.SelectionNotify {
				notify := xproto.SelectionNotifyEvent(ev.(xproto.SelectionNotifyEvent))
				if notify.Property == xproto.AtomNone {
					return "", fmt.Errorf("owner declined UTF8_STRING conversion")
				}

				// FIXED: xproto uses GetProperty (not GetWindowProperty)
				propReply, err := xproto.GetProperty(
					X,
					true, // Delete property after reading
					winID,
					propertyAtom,
					xproto.GetPropertyTypeAny,
					0,
					1024*1024,
				).Reply()

				if err != nil {
					return "", err
				}

				return string(propReply.Value), nil
			}
		}
	}
}

func main() {
	// 1. Connect to X Server
	X, err := xgb.NewConn()
	if err != nil {
		log.Fatalf("[-] Connection failed: %v", err)
	}
	defer X.Close()

	// 2. Initialize XFixes extension
	err = xfixes.Init(X)
	if err != nil {
		log.Fatalf("[-] Failed to initialize XFixes extension: %v", err)
	}

	// Query XFixes version
	_, err = xfixes.QueryVersion(X, 1, 0).Reply()
	if err != nil {
		log.Fatalf("[-] XFixes QueryVersion failed: %v", err)
	}

	setup := xproto.Setup(X)
	root := setup.DefaultScreen(X).Root

	clipboardAtom := getAtom(X, "CLIPBOARD")
	utf8Atom := getAtom(X, "UTF8_STRING")

	fmt.Println("[+] Registering XFixes selection notify mask on root window...")

	// FIXED: Mask bit for SelectionOwnerNotify is xfixes.SelectionEventMaskSetSelectionOwnerNotify
	err = xfixes.SelectSelectionInputChecked(
		X,
		root,
		clipboardAtom,
		1,
	).Check()

	if err != nil {
		log.Fatalf("[-] Failed to set SelectionInput mask: %v", err)
	}

	fmt.Println("[+] Monitoring CLIPBOARD changes in real time. Copy text anywhere to test...")

	var lastText string

	// 3. Asynchronous Event Loop
	for {
		ev, err := X.WaitForEvent()
		if err != nil {
			log.Printf("[-] Event error: %v", err)
			continue
		}

		if ev != nil {
			// FIXED: Type-assert directly to xfixes.SelectionNotifyEvent
			if _, ok := ev.(xfixes.SelectionNotifyEvent); ok {
				text, err := fetchClipboardText(X, root, clipboardAtom, utf8Atom)
				if err != nil {
					continue
				}

				if text != "" && text != lastText {
					lastText = text
					fmt.Printf("\n[+] [CLIPBOARD UPDATED] (%s)\n", time.Now().Format("15:04:05"))
					fmt.Printf("%s\n", text)
					fmt.Println("--------------------------------------------------")
				}
			}
		}
	}
}
