package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

const MaxLines = 200

// event struct to hold pid info
type event struct {
	Pid      uint32
	Filename [512]byte
}

// ExecveMonitor struct
type ExecveMonitor struct {
	app  *cview.Application
	view *cview.TextView
	logs []string

	done    chan struct{}
	msgChan chan string
}

// NewExecveMonitor sets up view and everything
func NewExecveMonitor(app *cview.Application) *ExecveMonitor {
	m := &ExecveMonitor{
		app:     app,
		view:    cview.NewTextView(),
		done:    make(chan struct{}),
		msgChan: make(chan string, 1024),
	}

	m.view.SetDynamicColors(true)
	m.view.SetScrollable(true)
	m.view.SetBorder(true)
	m.view.SetWrap(true)
	m.view.SetTitle("[black:violet:blr] Execve Syscalls")
	m.view.SetChangedFunc(func() { app.Draw() })

	// Start main processing loop
	go m.runLoop()

	return m
}

func (m *ExecveMonitor) Stop() {
	close(m.done)
}

func (m *ExecveMonitor) View() *cview.TextView {
	return m.view
}

// append format messages properly and display maxLines  number of events
func (m *ExecveMonitor) append(msg string) {
	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("%s[-] %s", timestamp, msg)

	m.logs = append([]string{formatted}, m.logs...)
	if len(m.logs) > MaxLines {
		m.logs = m.logs[:MaxLines]
	}

	m.app.QueueUpdateDraw(func() {
		m.view.SetText(joinLines(m.logs))
	})
}

// joinLines avoids repeated strings.Join allocations
func joinLines(lines []string) string {
	var buf bytes.Buffer
	for i, l := range lines {
		buf.WriteString(l)
		if i < len(lines)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// runLoop receives BPF events and updates the view
func (m *ExecveMonitor) runLoop() {
	// Launch BPF goroutine
	go func() {
		err := runBPF(m.msgChan)
		if err != nil {
			m.msgChan <- fmt.Sprintf("[red]BPF error: %v[-]", err)
		}
	}()

	for {
		select {
		case <-m.done:
			// Stop requested
			return

		case msg, ok := <-m.msgChan:
			if !ok {
				return
			}
			m.append(msg)
		}
	}
}

// runBPF streams BPF events and sends formatted strings to msgChan
func runBPF(out chan<- string) error {
	// Allow locking memory
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("memlock: %w", err)
	}

	// Load BPF objects
	objs := struct {
		ProgramExecve *ebpf.Program `ebpf:"handle_execve"`
		Ringbuf       *ebpf.Map     `ebpf:"ringbuf"`
	}{}

	spec, err := ebpf.LoadCollectionSpec("bpf/execve.bpf.o")
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		return fmt.Errorf("load objs: %w", err)
	}
	defer objs.ProgramExecve.Close()
	defer objs.Ringbuf.Close()

	// Attach tracepoint
	tp, err := link.Tracepoint("syscalls", "sys_enter_execve", objs.ProgramExecve, nil)
	if err != nil {
		return fmt.Errorf("attach tracepoint: %w", err)
	}
	defer tp.Close()

	// Create ring buffer reader
	rd, err := ringbuf.NewReader(objs.Ringbuf)
	if err != nil {
		return fmt.Errorf("ringbuf reader: %w", err)
	}
	defer rd.Close()

	// Stream events
	for {
		record, err := rd.Read()
		if err != nil {
			if err == ringbuf.ErrClosed {
				return nil
			}
			out <- fmt.Sprintf("[red]ringbuf error: %v[-]", err)
			continue
		}

		var e event
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &e); err != nil {
			out <- fmt.Sprintf("[red]decode error: %v[-]", err)
			continue
		}

		// Convert to C string
		idx := bytes.IndexByte(e.Filename[:], 0)
		filename := string(e.Filename[:idx])

		out <- fmt.Sprintf("[green]PID %d[-]  [cyan]%s[-]", e.Pid, filename)
	}
}
