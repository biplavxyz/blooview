package main

import (
	"codeberg.org/tslocum/cview"
	"github.com/biplavxyz/blooview/ui"
	"github.com/gdamore/tcell/v2"
)

func main() {
	// Define the Application
	app := cview.NewApplication()
	defer app.HandlePanic()

	// Enable Mouse Usage
	app.EnableMouse(true)

	// Create Tabbed UI
	mainTabbedUI := ui.Tabbedpanel(app)

	// Input Capture info
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 113 { // 113 = q
			app.Stop()
		} else if event.Rune() == 49 { // 49 = 1
			mainTabbedUI.SetCurrentTab("filesystem")
		} else if event.Rune() == 50 { // 50 = 2
			mainTabbedUI.SetCurrentTab("network")
		}

		return event
	})

	// Set Panels as Root
	app.SetRoot(mainTabbedUI, true)

	// Run the app and handle panic
	if err := app.Run(); err != nil {
		panic(err)
	}
}
