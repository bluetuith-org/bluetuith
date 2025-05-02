package views

import "sync"

// viewOperation holds an operation manager instance.
type viewOperation struct {
	cancel func()
	lock   sync.Mutex

	root *Views
}

// newViewOperation returns a new operations manager.
func newViewOperation(root *Views) *viewOperation {
	return &viewOperation{root: root}
}

// startOperation sets up the cancellation handler,
// and starts the operation.
func (v *viewOperation) startOperation(dofunc, cancel func()) {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.cancel != nil {
		v.root.status.InfoMessage("Operation still in progress", false)
		return
	}

	v.cancel = cancel

	go func() {
		dofunc()
		v.cancelOperation(false)
	}()
}

// cancelOperation cancels the currently running operation.
func (v *viewOperation) cancelOperation(cancelfunc bool) {
	var cancel func()

	v.lock.Lock()
	defer v.lock.Unlock()

	if v.cancel == nil {
		return
	}

	cancel = v.cancel
	v.cancel = nil

	if cancelfunc {
		go cancel()
	}
}
