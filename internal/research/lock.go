package research

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// lockHandle wraps an flock'd lock file.
type lockHandle struct {
	f  *os.File
	ex bool
}

// acquireLock takes a shared or exclusive flock on data/.research-rebuild.lock
// with a bounded retry; unavailable after timeout is exit 4 (§1).
func acquireLock(root string, exclusive bool, timeout time.Duration) (*lockHandle, error) {
	p := filepath.Join(root, filepath.FromSlash(LockPath))
	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, iof("open lock %s: %v", LockPath, err)
	}
	how := syscall.LOCK_SH
	if exclusive {
		how = syscall.LOCK_EX
	}
	deadline := time.Now().Add(timeout)
	for {
		err := syscall.Flock(int(f.Fd()), how|syscall.LOCK_NB)
		if err == nil {
			return &lockHandle{f: f, ex: exclusive}, nil
		}
		if time.Now().After(deadline) {
			f.Close()
			return nil, &Error{Code: ExitIO, Msg: "lock " + LockPath + " unavailable after timeout"}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (l *lockHandle) Release() {
	if l == nil || l.f == nil {
		return
	}
	syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	l.f.Close()
	l.f = nil
}
