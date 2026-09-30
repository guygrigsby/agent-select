package routing

import "sync/atomic"

// Engine pins one immutable snapshot generation per call and swaps
// generations atomically. A nil-constructed engine is unavailable, not an
// empty usable snapshot.
type Engine struct {
	cur atomic.Pointer[Snapshot]
}

// NewEngine creates an engine over the given snapshot. A nil snapshot (or
// one that is not a constructed value) yields an unavailable engine whose
// ActiveHash is empty.
func NewEngine(s *Snapshot) *Engine {
	e := &Engine{}
	if s != nil && s.hash != "" {
		e.cur.Store(s)
	}
	return e
}

// Adopt validates replacement bytes and swaps generations atomically. Any
// failure retains the prior immutable generation.
func (e *Engine) Adopt(data []byte) error {
	s, err := DecodeSnapshot(data)
	if err != nil {
		return err
	}
	e.cur.Store(s)
	return nil
}

// ActiveHash returns the current snapshot hash, empty when unavailable.
func (e *Engine) ActiveHash() string {
	if s := e.cur.Load(); s != nil {
		return s.Hash()
	}
	return ""
}

// pinned returns the snapshot generation pinned for one call.
func (e *Engine) pinned() *Snapshot { return e.cur.Load() }
