package pkg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/shirou/gopsutil/v4/process"
)

type ConnInfo struct {
	PID         int
	User        string
	UID         int
	Exe         string
	Cwd         string
	LocalAddr   string
	RemoteAddr  string
	IsChildProc bool
}

func (c *ConnInfo) String() string {
	label := "PARENT"
	if c.IsChildProc {
		label = "CHILD "
	}
	return fmt.Sprintf("[neonpink:darkpurple:bl] %-8s %-8d %-8s %-8d %-55s %-40s %-s",
		label, c.PID, c.User, c.UID, c.RemoteAddr, c.Exe, c.Cwd,
	)
}

// NetworkMonitor holds the table and refresh loop
type NetworkMonitor struct {
	app      *cview.Application
	view     *cview.TextView
	done     chan struct{}
	prevCwds map[int]string // track PID → Cwd for highlighting
}

func NewNetworkMonitor(app *cview.Application) *NetworkMonitor {
	m := &NetworkMonitor{
		app:      app,
		view:     cview.NewTextView(),
		done:     make(chan struct{}),
		prevCwds: make(map[int]string),
	}

	m.view.SetDynamicColors(true)
	m.view.SetScrollable(true)
	m.view.SetBorder(true)
	m.view.SetWrap(true)
	m.view.SetTitle("[black:violet:blr] Established Network Connections")
	m.view.SetChangedFunc(func() {
		app.Draw()
	})

	return m
}

func (m *NetworkMonitor) Stop() {
	close(m.done)
}

func (m *NetworkMonitor) View() *cview.TextView {
	return m.view
}

// StartRefresh starts the background loop that updates the TextView every intervalMS milliseconds
func (m *NetworkMonitor) StartRefresh(intervalMS int) {
	go func() {
		ticker := time.NewTicker(time.Millisecond * time.Duration(intervalMS))
		defer ticker.Stop()

		for {
			select {
			case <-m.done:
				return
			case <-ticker.C:
				m.refresh()
			}
		}
	}()
}

// refresh retrieves the current network connections and updates the TextView
func (m *NetworkMonitor) refresh() {
	conns, err := GetEstablishedConnectionTree()
	if err != nil {
		conns = []ConnInfo{}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[neonpink:darkpurple:blr] %-8s %-8s %-8s %-8s %-55s %-40s %-s\n",
		"Type", "PID", "User", "UID", "RemoteAddr", "Exe", "Cwd"))

	for _, c := range conns {
		cwdChanged := false
		if prev, ok := m.prevCwds[c.PID]; ok && prev != c.Cwd {
			cwdChanged = true
		}
		m.prevCwds[c.PID] = c.Cwd

		if cwdChanged {
			// Highlight changed CWD
			sb.WriteString(fmt.Sprintf("[black:blue:bl] %-8s %-8d %-8s %-8d %-55s %-40s %-s\n",
				func() string {
					if c.IsChildProc {
						return "CHILD "
					} else {
						return "PARENT"
					}
				}(),
				c.PID, c.User, c.UID, c.RemoteAddr, c.Exe, c.Cwd,
			))
		} else {
			sb.WriteString(c.String())
			sb.WriteRune('\n')
		}
	}

	output := sb.String()
	m.app.QueueUpdateDraw(func() {
		m.view.SetText(output)
	})
}

// --- Network connection helpers ---

func GetEstablishedConnectionTree() ([]ConnInfo, error) {
	socks, err := getAllEstablishedSocks()
	if err != nil {
		return nil, err
	}

	var results []ConnInfo

	for _, sock := range socks {
		// Skip if PID is missing
		if sock.Process == nil {
			continue
		}

		pid := sock.Process.Pid

		// Descendant PIDs
		children, _ := getAllChildPIDs(pid)

		// Parent
		if info, err := buildConnInfo(sock, pid, false); err == nil {
			results = append(results, info)
		}

		// Children
		for _, cp := range children {
			if info, err := buildConnInfo(sock, cp, true); err == nil {
				results = append(results, info)
			}
		}
	}

	return results, nil
}

func getAllEstablishedSocks() ([]netstat.SockTabEntry, error) {
	filter := func(s *netstat.SockTabEntry) bool {
		return s.State == netstat.Established
	}

	v4, err := netstat.TCPSocks(filter)
	if err != nil {
		return nil, err
	}

	v6, err := netstat.TCP6Socks(filter)
	if err != nil {
		return nil, err
	}

	return append(v4, v6...), nil
}

func buildConnInfo(sock netstat.SockTabEntry, pid int, isChild bool) (ConnInfo, error) {
	ps, err := process.NewProcess(int32(pid))
	if err != nil {
		return ConnInfo{}, err
	}

	user, _ := ps.Username()
	exe, _ := ps.Exe()
	cwd, _ := ps.Cwd() // get current working directory

	return ConnInfo{
		PID:         pid,
		User:        user,
		UID:         int(sock.UID),
		Exe:         exe,
		Cwd:         cwd,
		LocalAddr:   sock.LocalAddr.String(),
		RemoteAddr:  sock.RemoteAddr.String(),
		IsChildProc: isChild,
	}, nil
}

func getAllChildPIDs(pid int) ([]int, error) {
	path := fmt.Sprintf("/proc/%d/task/%d/children", pid, pid)

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, err
	}

	content := strings.TrimSpace(string(raw))
	if content == "" {
		return []int{}, nil
	}

	fields := strings.Fields(content)
	var results []int

	for _, f := range fields {
		childPID, err := strconv.Atoi(f)
		if err != nil {
			continue
		}

		results = append(results, childPID)

		grand, err := getAllChildPIDs(childPID)
		if err == nil {
			results = append(results, grand...)
		}
	}

	return results, nil
}
