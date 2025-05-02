package views

import (
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/tview"
	"go.uber.org/atomic"
)

// viewPages holds a pages manager for multiple views.
type viewPages struct {
	page        atomic.String
	pageContext atomic.String

	*tview.Pages
}

// newViewPages returns a new viewPages.
func newViewPages() *viewPages {
	p := &viewPages{
		Pages: tview.NewPages(),
	}

	p.page.Store(devicePage.String())
	p.pageContext.Store(string(keybindings.ContextApp))

	return p
}

// currentPage returns the currently focused page.
func (v *viewPages) currentPage(set ...string) string {
	var pg string

	if set != nil {
		pg = set[0]
		v.page.Store(pg)
	} else {
		pg = v.page.Load()
	}

	return pg
}

// currentContext gets or sets the current keybinding/page context of the currently focused page.
func (v *viewPages) currentContext(set ...keybindings.Context) keybindings.Context {
	var c keybindings.Context

	if set != nil {
		c = set[0]
		v.pageContext.Store(string(c))
	} else {
		c = keybindings.Context(v.pageContext.Load())
	}

	return c
}
