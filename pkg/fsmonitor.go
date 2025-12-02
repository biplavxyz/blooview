// Package pkg
package pkg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
)

const (
	MaxLogs  = 4000 // maximum lines to keep
	TrimLogs = 1000 // number of lines to remove when exceeding MaxLogs
)

// FsMonitor struct to hold view state
type FsMonitor struct {
	app     *cview.Application
	view    *cview.TextView
	watcher *fsnotify.Watcher
	logs    []string
	done    chan struct{}
}

// NewFsMonitor is the main view for Filesystem changes
func NewFsMonitor(app *cview.Application) *FsMonitor {
	m := &FsMonitor{
		app:  app,
		view: cview.NewTextView(),
		done: make(chan struct{}),
	}

	m.view.SetBorder(true)
	m.view.SetTitle("[black:violet:blr] Filesystem Events ")
	m.view.SetDynamicColors(true)
	m.view.SetScrollable(true)
	m.view.SetScrollBarColor(tcell.ColorOrange)
	m.view.SetWrap(true)
	m.view.SetChangedFunc(func() {
		app.Draw()
	})

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		m.append("ERROR: " + err.Error())
		return m
	}
	m.watcher = watcher

	paths := ReadToml(GetConfigPath())
	for _, p := range paths {
		if err := m.watcher.Add(p); err != nil {
			m.append("Failed to watch: " + p + " → " + err.Error())
		}
	}

	go m.runLoop()

	return m
}

// appending messages
func (m *FsMonitor) append(msg string) {
	m.logs = append(m.logs, msg)

	// If the log exceeds the maximum, trim the oldest TrimLogs lines
	if len(m.logs) > MaxLogs {
		m.logs = m.logs[TrimLogs:] // keep the most recent logs
	}

	m.app.QueueUpdateDraw(func() {
		m.view.SetText(strings.Join(m.logs, "\n"))
		m.view.ScrollToEnd()
	})
}

// Runs Loop for updating events constantly
func (m *FsMonitor) runLoop() {
	for {
		select {
		case <-m.done:
			return

		case evt := <-m.watcher.Events:
			m.append(m.coloredEvent(evt))

		case err := <-m.watcher.Errors:
			m.append("[red]ERROR: " + err.Error())
		}
	}
}

// Stop can be called if needed
func (m *FsMonitor) Stop() {
	close(m.done)
	if m.watcher != nil {
		m.watcher.Close()
	}
}

// View is called from panels
func (m *FsMonitor) View() *cview.TextView {
	return m.view
}

// Beautified events with color highlighting
func (m *FsMonitor) coloredEvent(evt fsnotify.Event) string {
	// Get Current Time
	curTime := time.Now().Format("2006-01-02 15:04:05")

	// Get file's owner and group
	uid, gid := GetUID(evt.Name)

	// Determine event type and color
	var eventType string
	switch {
	case evt.Op&fsnotify.Create == fsnotify.Create:
		eventType = "[green:black:ru]CREATE  [white]"
	case evt.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = "[red:black:ru]REMOVE  [white]"
	case evt.Op&fsnotify.Write == fsnotify.Write:
		eventType = "[yellow:black:ru]WRITE   [white]"
	case evt.Op&fsnotify.Rename == fsnotify.Rename:
		eventType = "[yellow:black:ru]RENAME  [white]"
	case evt.Op&fsnotify.Chmod == fsnotify.Chmod:
		eventType = "[yellow:black:ru]CHMOD   [white]"
	default:
		eventType = "[white:black:ru]UNKNOWN [white]"
	}

	return fmt.Sprintf("%s %s %s %s %s", curTime, eventType, uid, gid, evt.Name)
}

// GetUID returns UID for a certain event
func GetUID(path string) (string, string) {
	info, err := os.Stat(path)
	if err != nil {
		return "N/A     ", "N/A     "
	}

	stat := info.Sys().(*syscall.Stat_t)

	uidStr := strconv.FormatUint(uint64(stat.Uid), 10)
	gidStr := strconv.FormatUint(uint64(stat.Gid), 10)

	uidPadded := fmt.Sprintf("%-8s", uidStr)
	gidPadded := fmt.Sprintf("%-8s", gidStr)

	return uidPadded, gidPadded
}
