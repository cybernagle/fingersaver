package agent

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifierNotifyWakesWaiter(t *testing.T) {
	n := NewAgentNotifier()
	ch := n.WaitCh("auth")

	go func() {
		time.Sleep(50 * time.Millisecond)
		n.Notify("auth", "done")
	}()

	select {
	case <-ch:
		// Expected: channel closed.
	case <-time.After(2 * time.Second):
		t.Fatal("WaitCh was not closed after Notify")
	}
}

func TestNotifierMultipleWaiters(t *testing.T) {
	n := NewAgentNotifier()

	var wg sync.WaitGroup
	woken := make([]bool, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		ch := n.WaitCh("worker")
		idx := i
		go func() {
			defer wg.Done()
			<-ch
			woken[idx] = true
		}()
	}

	time.Sleep(50 * time.Millisecond)
	n.Notify("worker", "done")
	wg.Wait()

	for i, w := range woken {
		assert.True(t, w, "waiter %d should be woken", i)
	}
}

func TestNotifierNotifyBeforeWait(t *testing.T) {
	n := NewAgentNotifier()
	n.Notify("auth", "done")

	ch := n.WaitCh("auth")
	select {
	case <-ch:
		// Expected: already closed.
	default:
		t.Fatal("WaitCh should be closed when Notify was called first")
	}
}

func TestNotifierClearResets(t *testing.T) {
	n := NewAgentNotifier()
	n.Notify("auth", "done")
	n.Clear("auth")

	ch := n.WaitCh("auth")
	select {
	case <-ch:
		t.Fatal("WaitCh should not be closed after Clear")
	case <-time.After(100 * time.Millisecond):
		// Expected: still open.
	}
}

func TestNotifierNoNotifyNoWake(t *testing.T) {
	n := NewAgentNotifier()
	ch := n.WaitCh("auth")

	select {
	case <-ch:
		t.Fatal("WaitCh should not close without Notify")
	case <-time.After(200 * time.Millisecond):
		// Expected: still open.
	}
}

func newNotifierWithShortPath(t *testing.T) *AgentNotifier {
	t.Helper()
	// macOS Unix socket path limit is ~104 chars. Use /tmp to stay short.
	sockDir, err := os.MkdirTemp("", "fs-ht")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(sockDir) })
	n := NewAgentNotifier()
	n.sockPath = filepath.Join(sockDir, "h.sock")
	return n
}

func TestNotifierUnixSocket(t *testing.T) {
	n := newNotifierWithShortPath(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, n.Start(ctx))
	defer n.Stop()

	// Connect and send notification.
	conn, err := net.Dial("unix", n.sockPath)
	require.NoError(t, err)
	defer conn.Close()

	msg, _ := json.Marshal(hookPayload{Session: "auth", Status: "done"})
	conn.Write(msg)

	buf := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	nr, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "ok\n", string(buf[:nr]))

	// Verify waiter is woken.
	ch := n.WaitCh("auth")
	select {
	case <-ch:
		// Expected.
	case <-time.After(2 * time.Second):
		t.Fatal("WaitCh not closed after socket notification")
	}
}

func TestNotifierUnixSocketIgnoresEmptySession(t *testing.T) {
	n := newNotifierWithShortPath(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, n.Start(ctx))
	defer n.Stop()

	conn, err := net.Dial("unix", n.sockPath)
	require.NoError(t, err)
	msg, _ := json.Marshal(hookPayload{Session: "", Status: "done"})
	conn.Write(msg)
	conn.Close()

	time.Sleep(100 * time.Millisecond)

	// No notification should have been recorded.
	ch := n.WaitCh("auth")
	select {
	case <-ch:
		t.Fatal("empty session should not trigger notification")
	case <-time.After(100 * time.Millisecond):
		// Expected.
	}
}
