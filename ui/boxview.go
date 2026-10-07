package ui

import (
	"codeberg.org/tslocum/cview"
	"github.com/biplavxyz/blooview/pkg"
	"github.com/gdamore/tcell/v2"
)

func BoxViewNetwork() *cview.TextView {
	// Box Text View Primitive
	myView := cview.NewTextView()

	// Set the properties for the view
	myView.SetBorder(true)
	myView.SetBorderColor(tcell.ColorMaroon)

	myView.SetTitle("[black:violet:blr]Events Related To Network")
	myView.SetTitleAlign(cview.AlignCenter)
	myView.SetTitleColor(tcell.ColorGreen)
	myView.SetDynamicColors(true)
	myView.SetBorderAttributes(tcell.AttrBold)

	// return the view
	return myView
}

func BoxViewHome(title string) *cview.TextView {
	// Box Text View Primitive
	homeView := cview.NewTextView()

	// Set the properties for the view
	homeView.SetBorder(true)
	homeView.SetBorderColor(tcell.ColorMaroon)
	homeView.SetTitle(title)
	homeView.SetTitleAlign(cview.AlignCenter)
	homeView.SetTitleColor(tcell.ColorGreen)
	homeView.SetBorderAttributes(tcell.AttrBold)
	homeView.SetTextColor(tcell.ColorGreenYellow)

	fPaths, err := pkg.GetFilepaths(pkg.GetConfigPath())
	if err != nil {
		fPaths = []string{"Error decoding filepaths from TOML file"}
	}
	files := pkg.GetFilePathsSingle(fPaths)

	// Get Config File Info
	systemInfo := "----------------------------------\nWatching provided paths:\n----------------------------------\n" + files

	homeView.SetText(systemInfo)

	// return the view
	return homeView
}
