//go:build darwin

package main

import "fyne.io/fyne/v2"

func configureWindow(w fyne.Window) {
	// On macOS, Fyne handles standard window management.
	// For true overlay behavior (always-on-top, transparency),
	// CGo with NSWindow APIs could be added as an enhancement.
}
