package views

import (
	"sync"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/tview"
	"go.uber.org/atomic"
)

type viewPages struct {
	page        atomic.String
	pageContext atomic.String
	mu          sync.Mutex

	*tview.Pages
}

func newViewPages() *viewPages {
	p := &viewPages{
		Pages: tview.NewPages(),
	}

	p.page.Store(devicePage.String())
	p.pageContext.Store(string(keybindings.ContextApp))

	return p
}

func (v *viewPages) currentPage(set ...string) string {
	v.mu.Lock()
	defer v.mu.Unlock()

	var pg string
	if set != nil {
		pg = set[0]
		v.page.Store(pg)
	} else {
		pg = v.page.Load()
	}

	return pg
}

func (v *viewPages) currentContext(set ...keybindings.Context) keybindings.Context {
	v.mu.Lock()
	defer v.mu.Unlock()

	var c keybindings.Context
	if set != nil {
		c = set[0]
		v.pageContext.Store(string(c))
	} else {
		c = keybindings.Context(v.pageContext.Load())
	}

	return c
}
