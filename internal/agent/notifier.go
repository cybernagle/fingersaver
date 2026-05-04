package agent

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// HookListener is the transport layer for receiving agent notifications.
// Implementations include UnixSocketListener (local hooks) and can be
// extended with HTTPListener (remote/phone) later.
type HookListener interface {
	Start(ctx context.Context, handler func(session, status string)) error
	Stop()
}

// hookPayload is the JSON message sent to the socket.
type hookPayload struct {
	Session string `json:"session"`
	Status  string `json:"status"`
}

// AgentNotifier tracks stop notifications from coding agents.
type AgentNotifier struct {
	mu       sync.Mutex
	waiters  map[string][]chan struct{} // session → wait channels
	recent   map[string]string          // session → status (arrived before WaitCh)
	sockPath string
	listener net.Listener
	done     chan struct{}
}

func NewAgentNotifier() *AgentNotifier {
	home, _ := os.UserHomeDir()
	return &AgentNotifier{
		waiters:  make(map[string][]chan struct{}),
		recent:   make(map[string]string),
		sockPath: filepath.Join(home, ".fingersaver", "hooks.sock"),
		done:     make(chan struct{}),
	}
}

// WaitCh returns a channel that closes when the session's agent stops.
// If a notification was already received (Notify called before WaitCh),
// returns an already-closed channel.
func (n *AgentNotifier) WaitCh(session string) <-chan struct{} {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Check recent notifications first.
	if _, ok := n.recent[session]; ok {
		delete(n.recent, session)
		ch := make(chan struct{})
		close(ch)
		return ch
	}

	ch := make(chan struct{})
	n.waiters[session] = append(n.waiters[session], ch)
	return ch
}

// Notify marks a session as stopped and wakes all waiters.
func (n *AgentNotifier) Notify(session, status string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	channels := n.waiters[session]
	for _, ch := range channels {
		close(ch)
	}
	delete(n.waiters, session)

	// Store in recent in case WaitCh is called later.
	if len(channels) == 0 {
		n.recent[session] = status
	}
}

// Clear resets the notification state for a session.
func (n *AgentNotifier) Clear(session string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.waiters, session)
	delete(n.recent, session)
}

// Start begins listening on the Unix socket for hook notifications.
func (n *AgentNotifier) Start(ctx context.Context) error {
	os.Remove(n.sockPath)

	if err := os.MkdirAll(filepath.Dir(n.sockPath), 0o755); err != nil {
		return err
	}

	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "unix", n.sockPath)
	if err != nil {
		return err
	}
	n.listener = ln

	go n.acceptLoop(ctx)
	return nil
}

// Stop closes the listener and cleans up the socket file.
func (n *AgentNotifier) Stop() {
	close(n.done)
	if n.listener != nil {
		n.listener.Close()
	}
	os.Remove(n.sockPath)
}

func (n *AgentNotifier) acceptLoop(ctx context.Context) {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			select {
			case <-n.done:
				return
			case <-ctx.Done():
				return
			default:
				log.Printf("[notifier] accept error: %v", err)
				continue
			}
		}
		go n.handleConn(conn)
	}
}

func (n *AgentNotifier) handleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 4096)
	nr, err := conn.Read(buf)
	if err != nil {
		return
	}

	var msg hookPayload
	if err := json.Unmarshal(buf[:nr], &msg); err != nil {
		return
	}
	if msg.Session == "" {
		return
	}

	n.Notify(msg.Session, msg.Status)
	conn.Write([]byte("ok\n"))
}

// SockPath returns the Unix socket path, used by the notify subcommand.
func (n *AgentNotifier) SockPath() string {
	return n.sockPath
}
