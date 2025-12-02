// Package ui
package ui

import (
	"codeberg.org/tslocum/cview"
	"github.com/biplavxyz/blooview/pkg"
	"github.com/gdamore/tcell/v2"
)

func Tabbedpanel(app *cview.Application) *cview.TabbedPanels {
	// Create Panel
	panel := cview.NewTabbedPanels()

	// Redraw on some change
	panel.SetChangedFunc(func() { app.Draw() })

	// Properties of Panel
	panel.SetTabBackgroundColor(tcell.ColorBlueViolet)
	panel.SetTabTextColor(tcell.ColorWhite)
	panel.SetTabBackgroundColorFocused(tcell.ColorOrange)

	// Set Panels Here
	fsPanel := pkg.NewFsMonitor(app)
	networkMonitor := pkg.NewNetworkMonitor(app)

	// Tabs for the Panel
	panel.AddTab("blooview", "BlooView - Filesystem and Network Monitoring Tool", BoxViewHome("BlooView Stat"))
	panel.AddTab("filesystem", "[1]-Files", fsPanel.View())
	panel.AddTab("network", "[2]-Network Processes", networkMonitor.View())

	// Update every 1 second
	networkMonitor.StartRefresh(1000)

	// panel.SetCurrentTab("filesystem")

	// Return Panel
	return panel
}
