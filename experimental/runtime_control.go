package experimental

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/rendis/statepro/v3/instrumentation"
)

// contextMutex preserves synchronous locking for context-free legacy methods,
// while allowing execution callers to cancel their wait without spawning goroutines.
type contextMutex struct {
	once  sync.Once
	token chan struct{}
}

func (m *contextMutex) init() {
	m.once.Do(func() { m.token = make(chan struct{}, 1); m.token <- struct{}{} })
}
func (m *contextMutex) Lock()   { m.init(); <-m.token }
func (m *contextMutex) Unlock() {
	select {
	case m.token <- struct{}{}:
	default:
		// Match sync.Mutex: unlocking an unlocked mutex is a bug, not a deadlock.
		panic("statepro: unlock of unlocked contextMutex")
	}
}
func (m *contextMutex) LockContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	m.init()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.token:
		if err := ctx.Err(); err != nil {
			m.Unlock()
			return err
		}
		return nil
	}
}

type callbackKey struct{}
type callbackScope struct {
	owner  *ExQuantumMachine
	active atomic.Bool
	parent *callbackScope
}

func (qm *ExQuantumMachine) callbackContext(ctx context.Context) (context.Context, func()) {
	parent, _ := ctx.Value(callbackKey{}).(*callbackScope)
	scope := &callbackScope{owner: qm, parent: parent}
	scope.active.Store(true)
	return context.WithValue(ctx, callbackKey{}, scope), func() { scope.active.Store(false) }
}

func (qm *ExQuantumMachine) lockContext(ctx context.Context) error {
	if err := qm.checkCallbackContext(ctx); err != nil {
		return err
	}
	if err := qm.quantumMachineMtx.LockContext(ctx); err != nil {
		return err
	}
	if qm.invokes != nil && qm.invokes.isClosed() {
		qm.quantumMachineMtx.Unlock()
		return instrumentation.ErrMachineClosed
	}
	return nil
}

func (qm *ExQuantumMachine) checkCallbackContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context must not be nil")
	}
	for scope, _ := ctx.Value(callbackKey{}).(*callbackScope); scope != nil; scope = scope.parent {
		if scope.owner == qm && scope.active.Load() {
			return instrumentation.ErrReentrantCall
		}
	}
	return nil
}

type invokeTask struct {
	universe, reality string
	cancel            context.CancelFunc
}
type invokeManager struct {
	mu     sync.Mutex
	limit  int
	closed bool
	next   uint64
	tasks  map[uint64]invokeTask
	idle   chan struct{}
}

func newInvokeManager(limit int) *invokeManager {
	idle := make(chan struct{})
	close(idle)
	return &invokeManager{limit: limit, tasks: make(map[uint64]invokeTask), idle: idle}
}

func (m *invokeManager) start(ctx context.Context, universe, reality, src string, run func(context.Context)) error {
	if ctx.Err() != nil {
		// The invoke would only observe a cancelled context. Skipping it must not
		// abort the synchronous transition that already ran its callbacks.
		return nil
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return instrumentation.ErrMachineClosed
	}
	if m.limit > 0 && len(m.tasks) >= m.limit {
		m.mu.Unlock()
		return &instrumentation.ResourceLimitError{Resource: "concurrent invokes", Limit: m.limit}
	}
	if len(m.tasks) == 0 {
		m.idle = make(chan struct{})
	}
	m.next++
	id := m.next
	// Async invokes run outside the synchronous callback's lock scope. Preserve
	// cancellation and caller values without inheriting its reentry marker.
	taskCtx, cancel := context.WithCancel(context.WithValue(ctx, callbackKey{}, (*callbackScope)(nil)))
	m.tasks[id] = invokeTask{universe: universe, reality: reality, cancel: cancel}
	m.mu.Unlock()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(taskCtx, "invoke panicked", "src", src, "panic", r)
			}
			cancel()
			m.mu.Lock()
			delete(m.tasks, id)
			if len(m.tasks) == 0 {
				close(m.idle)
			}
			m.mu.Unlock()
		}()
		if taskCtx.Err() == nil {
			run(taskCtx)
		}
	}()
	return nil
}

func (m *invokeManager) cancel(universe, reality string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, task := range m.tasks {
		if task.universe == universe && (reality == "" || task.reality == reality) {
			task.cancel()
		}
	}
}
func (m *invokeManager) isClosed() bool { m.mu.Lock(); defer m.mu.Unlock(); return m.closed }
func (m *invokeManager) running() int   { m.mu.Lock(); defer m.mu.Unlock(); return len(m.tasks) }
func (m *invokeManager) close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	for _, task := range m.tasks {
		task.cancel()
	}
}
func (m *invokeManager) wait(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	idle := m.idle
	m.mu.Unlock()
	select {
	case <-idle:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (qm *ExQuantumMachine) Close() error { qm.invokes.close(); return nil }
func (qm *ExQuantumMachine) WaitInvokes(ctx context.Context) error {
	if err := qm.checkCallbackContext(ctx); err != nil {
		return err
	}
	return qm.invokes.wait(ctx)
}
func (qm *ExQuantumMachine) GetSnapshotContext(ctx context.Context) (*instrumentation.MachineSnapshot, error) {
	if err := qm.checkCallbackContext(ctx); err != nil {
		return nil, err
	}
	if err := qm.quantumMachineMtx.LockContext(ctx); err != nil {
		return nil, err
	}
	defer qm.quantumMachineMtx.Unlock()
	return qm.snapshotUnlockedWithError()
}
