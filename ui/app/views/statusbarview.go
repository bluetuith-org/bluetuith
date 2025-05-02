package views

import (
	"context"
	"errors"
	"time"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

const (
	statusInputPage    viewName = "input"
	statusMessagesPage viewName = "messages"
)

type statusBarView struct {
	// MessageBox is an area to display messages.
	MessageBox *tview.TextView

	// Help is an area to display help keybindings.
	Help *tview.TextView

	// InputField is an area to interact with messages.
	InputField *tview.InputField

	sctx    context.Context
	scancel context.CancelFunc
	msgchan chan message

	*Views

	*tview.Pages
}

type message struct {
	text    string
	persist bool
}

func (s *statusBarView) Initialize() error {
	s.Pages = tview.NewPages()
	s.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	s.InputField = tview.NewInputField()
	s.InputField.SetLabelColor(theme.GetColor(theme.ThemeText))
	s.InputField.SetFieldTextColor(theme.GetColor(theme.ThemeText))
	s.InputField.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	s.InputField.SetFieldBackgroundColor(theme.GetColor(theme.ThemeBackground))

	s.MessageBox = tview.NewTextView()
	s.MessageBox.SetDynamicColors(true)
	s.MessageBox.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	s.Help = tview.NewTextView()
	s.Help.SetDynamicColors(true)
	s.Help.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	s.AddPage(statusInputPage.String(), s.InputField, true, true)
	s.AddPage(statusMessagesPage.String(), s.MessageBox, true, true)
	s.SwitchToPage(statusMessagesPage.String())

	s.msgchan = make(chan message, 10)
	s.sctx, s.scancel = context.WithCancel(context.Background())

	go s.startStatus()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(s.Pages, 1, 0, false)

	s.layout.AddItem(flex, flex.GetItemCount(), 0, false)

	return nil
}

func (s *statusBarView) SetRootView(root *Views) {
	s.Views = root
}

func (s *statusBarView) Release() {
	s.scancel()
}

// SetInput sets the inputfield label and returns the input text.
func (s *statusBarView) SetInput(label string, multichar ...struct{}) string {
	return s.waitForInput(context.Background(), label, multichar...)
}

func (s *statusBarView) waitForInput(ctx context.Context, label string, multichar ...struct{}) string {
	input := make(chan string)

	go func() {
		exited := make(chan struct{}, 1)
		exit := func() {
			s.SwitchToPage(statusMessagesPage.String())

			_, item := s.pages.GetFrontPage()
			s.app.FocusPrimitive(item)

			exited <- struct{}{}
		}

		s.app.InstantDraw(func() {
			s.InputField.SetText("")
			s.InputField.SetLabel("[::b]" + label + " ")

			if multichar != nil {
				s.InputField.SetAcceptanceFunc(nil)
				s.InputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch s.kb.Key(event) {
					case keybindings.KeySelect:
						input <- s.InputField.GetText()

						exit()

					case keybindings.KeyClose:
						input <- ""

						exit()
					}

					return event
				})
			} else {
				s.InputField.SetAcceptanceFunc(tview.InputFieldMaxLength(1))
				s.InputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					input <- string(event.Rune())
					exit()

					return event
				})
			}

			s.SwitchToPage(statusInputPage.String())
			s.app.FocusPrimitive(s.InputField)
		})

		select {
		case <-ctx.Done():
			s.app.InstantDraw(func() {
				s.InputField.SetText("")
				exit()
			})

		case <-exited:
		}
	}()

	return <-input
}

// InfoMessage sends an info message to the status bar.
func (s *statusBarView) InfoMessage(text string, persist bool) {
	if s.msgchan == nil {
		return
	}

	select {
	case s.msgchan <- message{theme.ColorWrap(theme.ThemeStatusInfo, text), persist}:
		return

	default:
	}
}

// ErrorMessage sends an error message to the status bar.
func (s *statusBarView) ErrorMessage(err error) {
	if s.msgchan == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		return
	}

	select {
	case s.msgchan <- message{theme.ColorWrap(theme.ThemeStatusError, "Error: "+err.Error()), false}:
		return

	default:
	}
}

// startStatus starts the message event loop
func (s *statusBarView) startStatus() {
	var text string
	var cleared bool

	t := time.NewTicker(2 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-s.sctx.Done():
			return

		case msg, ok := <-s.msgchan:
			if !ok {
				return
			}

			t.Reset(2 * time.Second)

			cleared = false

			if msg.persist {
				text = msg.text
			}

			if !msg.persist && text != "" {
				text = ""
			}

			s.app.InstantDraw(func() {
				s.MessageBox.SetText(msg.text)
			})

		case <-t.C:
			if cleared {
				continue
			}

			cleared = true

			s.app.InstantDraw(func() {
				s.MessageBox.SetText(text)
			})
		}
	}
}
