//go:build linux

package main

import "fyne.io/fyne/v2"

func configureWindow(w fyne.Window) {
	// On Linux, Fyne works with both X11 and Wayland.
	// Always-on-top behavior is managed by the compositor on Wayland.
	// On X11, Fyne sets the appropriate window manager hints.
}
