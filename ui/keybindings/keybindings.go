package keybindings

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Key describes the application keybinding type.
type Key string

// The different application keybinding types.
const (
	KeyMenu                        Key = "Menu"
	KeySelect                      Key = "Select"
	KeyCancel                      Key = "Cancel"
	KeySuspend                     Key = "Suspend"
	KeyQuit                        Key = "Quit"
	KeySwitch                      Key = "Switch"
	KeyClose                       Key = "Close"
	KeyHelp                        Key = "Help"
	KeyAdapterChange               Key = "AdapterChange"
	KeyAdapterTogglePower          Key = "AdapterTogglePower"
	KeyAdapterToggleDiscoverable   Key = "AdapterToggleDiscoverable"
	KeyAdapterTogglePairable       Key = "AdapterTogglePairable"
	KeyAdapterToggleScan           Key = "AdapterToggleScan"
	KeyDeviceSendFiles             Key = "DeviceSendFiles"
	KeyDeviceNetwork               Key = "DeviceNetwork"
	KeyDeviceConnect               Key = "DeviceConnect"
	KeyDevicePair                  Key = "DevicePair"
	KeyDeviceTrust                 Key = "DeviceTrust"
	KeyDeviceBlock                 Key = "DeviceBlock"
	KeyDeviceAudioProfiles         Key = "DeviceAudioProfiles"
	KeyDeviceInfo                  Key = "DeviceInfo"
	KeyDeviceRemove                Key = "DeviceRemove"
	KeyPlayerShow                  Key = "PlayerShow"
	KeyPlayerHide                  Key = "PlayerHide"
	KeyFilebrowserDirForward       Key = "FilebrowserDirForward"
	KeyFilebrowserDirBack          Key = "FilebrowserDirBack"
	KeyFilebrowserSelect           Key = "FilebrowserSelect"
	KeyFilebrowserInvertSelection  Key = "FilebrowserInvertSelection"
	KeyFilebrowserSelectAll        Key = "FilebrowserSelectAll"
	KeyFilebrowserRefresh          Key = "FilebrowserRefresh"
	KeyFilebrowserToggleHidden     Key = "FilebrowserToggleHidden"
	KeyFilebrowserConfirmSelection Key = "FilebrowserConfirmSelection"
	KeyProgressView                Key = "ProgressView"
	KeyProgressTransferSuspend     Key = "ProgressTransferSuspend"
	KeyProgressTransferResume      Key = "ProgressTransferResume"
	KeyProgressTransferCancel      Key = "ProgressTransferCancel"
	KeyPlayerTogglePlay            Key = "PlayerTogglePlay"
	KeyPlayerNext                  Key = "PlayerNext"
	KeyPlayerPrevious              Key = "PlayerPrevious"
	KeyPlayerSeekForward           Key = "PlayerSeekForward"
	KeyPlayerSeekBackward          Key = "PlayerSeekBackward"
	KeyPlayerStop                  Key = "PlayerStop"
	KeyNavigateUp                  Key = "NavigateUp"
	KeyNavigateDown                Key = "NavigateDown"
	KeyNavigateRight               Key = "NavigateRight"
	KeyNavigateLeft                Key = "NavigateLeft"
	KeyNavigateTop                 Key = "NavigateTop"
	KeyNavigateBottom              Key = "NavigateBottom"
)

// Context describes the context where the keybinding is
// supposed to be applied in.
type Context string

// The different context types for keybindings.
const (
	ContextApp      Context = "App"
	ContextDevice   Context = "Device"
	ContextFiles    Context = "Files"
	ContextProgress Context = "Progress"
)

// KeyData stores the metadata for the key.
type KeyData struct {
	Title   string
	Context Context
	Kb      Keybinding
	Global  bool
}

// Keybinding stores the keybinding.
type Keybinding struct {
	Key  tcell.Key
	Rune rune
	Mod  tcell.ModMask
}

// Keybindings contains an entire list of keybindings and its associated contexts and other data.
type Keybindings struct {
	keyData        map[Key]*KeyData
	contextKeys    map[Context]map[Keybinding]Key
	navigationKeys map[Key]Keybinding
	translateKeys  map[string]string
}

// NewKeybindings returns a new keybindings configuration.
func NewKeybindings() *Keybindings {
	k := &Keybindings{}

	k.initData()
	k.initKeys()

	return k
}

// Data returns the key data associated with
// the provided keyID and operation name.
func (k *Keybindings) Data(key Key) *KeyData {
	return k.keyData[key]
}

// Key returns the operation name for the provided keyID
// and the keyboard event.
func (k *Keybindings) Key(event *tcell.EventKey, keyContexts ...Context) Key {
	ch := event.Rune()
	if event.Key() != tcell.KeyRune {
		ch = ' '
	}

	mod := event.Modifiers()
	if unicode.IsUpper(ch) && mod&tcell.ModShift != 0 {
		mod &^= tcell.ModShift
	}

	kb := Keybinding{event.Key(), ch, mod}

	if key, ok := k.checkContexts(kb, keyContexts); ok {
		return key
	}

	if key, ok := k.checkContexts(kb, []Context{
		ContextApp,
		ContextDevice,
	}); ok {
		return key
	}

	return ""
}

// Name formats and returns the key's name.
func (k *Keybindings) Name(kb Keybinding) string {
	if kb.Key == tcell.KeyRune {
		keyname := string(kb.Rune)
		if kb.Rune == ' ' {
			keyname = "Space"
		}

		if kb.Mod&tcell.ModAlt != 0 {
			keyname = "Alt+" + keyname
		}

		return keyname
	}

	return tcell.NewEventKey(kb.Key, kb.Rune, kb.Mod).Name()
}

// IsNavigation checks whether the provided key is a navigation key.
func (k *Keybindings) IsNavigation(pressed Key, event *tcell.EventKey) (*tcell.EventKey, bool) {
	kb := Keybinding{event.Key(), event.Rune(), event.Modifiers()}
	if kb.Key != tcell.KeyRune {
		kb.Rune = ' '
	}

	n, ok := k.navigationKeys[pressed]
	if !ok || n == kb {
		return nil, false
	}

	return tcell.NewEventKey(n.Key, n.Rune, n.Mod), true
}

// Validate validates the keybindings from the configuration.
func (k *Keybindings) Validate(kbMap map[string]string) error {
	if len(kbMap) == 0 {
		return nil
	}

	keyNames := make(map[string]tcell.Key)
	for key, names := range tcell.KeyNames {
		keyNames[names] = key
	}

	for keyType, key := range kbMap {
		k.checkBindings(keyType, key, keyNames)
	}

	keyErrors := make(map[Keybinding]string)

	for keyType, keydata := range k.keyData {
		for existing, data := range k.keyData {
			if data.Kb == keydata.Kb && data.Title != keydata.Title {
				if data.Context == keydata.Context || data.Global || keydata.Global {
					goto KeyError
				}

				continue

			KeyError:
				if _, ok := keyErrors[keydata.Kb]; !ok {
					keyErrors[keydata.Kb] = fmt.Sprintf("- %s will override %s (%s)", keyType, existing, k.Name(keydata.Kb))
				}
			}
		}
	}

	if len(keyErrors) > 0 {
		err := "Config: The following keybindings will conflict:\n"
		for _, ke := range keyErrors {
			err += ke + "\n"
		}

		return errors.New(strings.TrimRight(err, "\n"))
	}

	return nil
}

// checkContexts checks whether a keybinding exists within the provided keybinding context.
func (k *Keybindings) checkContexts(kb Keybinding, contexts []Context) (Key, bool) {
	for _, context := range contexts {
		if operation, ok := k.contextKeys[context][kb]; ok {
			return operation, true
		}

	}

	return "", false
}

// checkBindings validates the provided keybinding.
//
//gocyclo:ignore
func (k *Keybindings) checkBindings(keyType, key string, keyNames map[string]tcell.Key) error {
	var runes []rune
	var keys []tcell.Key

	if _, ok := k.keyData[Key(keyType)]; !ok {
		return fmt.Errorf("config: Invalid key type %s", keyType)
	}

	keybinding := Keybinding{
		Key:  tcell.KeyRune,
		Rune: ' ',
		Mod:  tcell.ModNone,
	}

	tokens := strings.FieldsFunc(key, func(c rune) bool {
		return unicode.IsSpace(c) || c == '+'
	})

	for _, token := range tokens {
		length := runewidth.StringWidth(token)
		if length > 1 {
			token = cases.Title(language.Und, cases.NoLower).String(token)
		} else if length == 1 {
			c, _ := utf8.DecodeRuneInString(token)

			keybinding.Rune = rune(c)
			runes = append(runes, keybinding.Rune)

			continue
		}

		if translated, ok := k.translateKeys[token]; ok {
			token = translated
		}

		switch token {
		case "Ctrl":
			keybinding.Mod |= tcell.ModCtrl

		case "Alt":
			keybinding.Mod |= tcell.ModAlt

		case "Shift":
			keybinding.Mod |= tcell.ModShift

		case "Space", "Plus":
			keybinding.Rune = ' '
			if token == "Plus" {
				keybinding.Rune = '+'
			}

			runes = append(runes, keybinding.Rune)

		default:
			if key, ok := keyNames[token]; ok {
				keybinding.Key = key
				keybinding.Rune = ' '
				keys = append(keys, keybinding.Key)
			}
		}
	}

	if keys != nil && runes != nil || len(runes) > 1 || len(keys) > 1 {
		return fmt.Errorf("config: More than one key entered for %s (%s)", keyType, key)
	}

	if keybinding.Mod&tcell.ModShift != 0 {
		keybinding.Rune = unicode.ToUpper(keybinding.Rune)

		if unicode.IsLetter(keybinding.Rune) {
			keybinding.Mod &^= tcell.ModShift
		}
	}

	if keybinding.Mod&tcell.ModCtrl != 0 {
		var modKey string

		switch {
		case len(keys) > 0:
			if key, ok := tcell.KeyNames[keybinding.Key]; ok {
				modKey = key
			}

		case len(runes) > 0:
			if keybinding.Rune == ' ' {
				modKey = "Space"
			} else {
				modKey = string(unicode.ToUpper(keybinding.Rune))
			}
		}

		if modKey != "" {
			modKey = "Ctrl-" + modKey
			if key, ok := keyNames[modKey]; ok {
				keybinding.Key = key
				keybinding.Rune = ' '
				keys = append(keys, keybinding.Key)
			}
		}
	}

	if keys == nil && runes == nil {
		return fmt.Errorf("config: No key specified or invalid keybinding for %s (%s)", keyType, key)
	}

	k.keyData[Key(keyType)].Kb = keybinding

	return nil
}

// initKeys initializes and stores the key types and contexts.
func (k *Keybindings) initKeys() {
	k.contextKeys = make(map[Context]map[Keybinding]Key)
	for keyName, key := range k.keyData {
		if k.contextKeys[key.Context] == nil {
			k.contextKeys[key.Context] = make(map[Keybinding]Key)
		}

		k.contextKeys[key.Context][key.Kb] = keyName
	}

	k.navigationKeys = map[Key]Keybinding{
		KeyNavigateUp:     {tcell.KeyUp, ' ', tcell.ModNone},
		KeyNavigateDown:   {tcell.KeyDown, ' ', tcell.ModNone},
		KeyNavigateRight:  {tcell.KeyRight, ' ', tcell.ModNone},
		KeyNavigateLeft:   {tcell.KeyLeft, ' ', tcell.ModNone},
		KeyNavigateTop:    {tcell.KeyPgUp, ' ', tcell.ModNone},
		KeyNavigateBottom: {tcell.KeyPgDn, ' ', tcell.ModNone},
	}

	k.translateKeys = map[string]string{
		"Pgup":      "PgUp",
		"Pgdn":      "PgDn",
		"Pageup":    "PgUp",
		"Pagedown":  "PgDn",
		"Upright":   "UpRight",
		"Downright": "DownRight",
		"Upleft":    "UpLeft",
		"Downleft":  "DownLeft",
		"Prtsc":     "Print",
		"Backspace": "Backspace2",
	}
}

// initData initializes and stores the keybindings configuration.
func (k *Keybindings) initData() {
	k.keyData = map[Key]*KeyData{
		KeySwitch: {
			Title:   "Switch",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyTab, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyClose: {
			Title:   "Close",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyEscape, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyQuit: {
			Title:   "Quit",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRune, 'Q', tcell.ModNone},
			Global:  true,
		},
		KeyMenu: {
			Title:   "Menu",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRune, 'm', tcell.ModAlt},
		},
		KeySelect: {
			Title:   "Select",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyEnter, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyCancel: {
			Title:   "Cancel",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyCtrlX, ' ', tcell.ModCtrl},
			Global:  true,
		},
		KeySuspend: {
			Title:   "Suspend",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyCtrlZ, ' ', tcell.ModCtrl},
			Global:  true,
		},
		KeyHelp: {
			Title:   "Help",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRune, '?', tcell.ModShift},
			Global:  true,
		},
		KeyNavigateUp: {
			Title:   "Navigate Up",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyUp, ' ', tcell.ModNone},
		},
		KeyNavigateDown: {
			Title:   "Navigate Down",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyDown, ' ', tcell.ModNone},
		},
		KeyNavigateRight: {
			Title:   "Navigate Right",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyNavigateLeft: {
			Title:   "Navigate Left",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyNavigateTop: {
			Title:   "Navigate Top",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyPgUp, ' ', tcell.ModNone},
		},
		KeyNavigateBottom: {
			Title:   "Navigate Bottom",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyPgDn, ' ', tcell.ModNone},
		},
		KeyAdapterTogglePower: {
			Title:   "Power",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'o', tcell.ModNone},
		},
		KeyAdapterToggleDiscoverable: {
			Title:   "Discoverable",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'S', tcell.ModNone},
		},
		KeyAdapterTogglePairable: {
			Title:   "Pairable",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'P', tcell.ModNone},
		},
		KeyAdapterToggleScan: {
			Title:   "Scan",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 's', tcell.ModNone},
		},
		KeyAdapterChange: {
			Title:   "Change",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'a', tcell.ModNone},
		},
		KeyDeviceConnect: {
			Title:   "Connect",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'c', tcell.ModNone},
		},
		KeyDevicePair: {
			Title:   "Pair",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'p', tcell.ModNone},
		},
		KeyDeviceTrust: {
			Title:   "Trust",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 't', tcell.ModNone},
		},
		KeyDeviceBlock: {
			Title:   "Block",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'b', tcell.ModNone},
		},
		KeyDeviceSendFiles: {
			Title:   "Send",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'f', tcell.ModNone},
		},
		KeyDeviceNetwork: {
			Title:   "Network Options",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'n', tcell.ModNone},
		},
		KeyDeviceAudioProfiles: {
			Title:   "Audio Profiles",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'A', tcell.ModNone},
		},
		KeyDeviceInfo: {
			Title:   "Info",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'i', tcell.ModNone},
		},
		KeyDeviceRemove: {
			Title:   "Remove",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'd', tcell.ModNone},
		},
		KeyPlayerShow: {
			Title:   "Show Media Player",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'm', tcell.ModNone},
		},
		KeyPlayerHide: {
			Title:   "Hide Media Player",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'M', tcell.ModNone},
		},
		KeyPlayerTogglePlay: {
			Title:   "Play/Pause",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, ' ', tcell.ModNone},
		},
		KeyPlayerNext: {
			Title:   "Next",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, '>', tcell.ModNone},
		},
		KeyPlayerPrevious: {
			Title:   "Previous",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, '<', tcell.ModNone},
		},
		KeyPlayerSeekForward: {
			Title:   "Seek Forward",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyPlayerSeekBackward: {
			Title:   "Seek Backward",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyPlayerStop: {
			Title:   "Stop",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, ']', tcell.ModNone},
		},
		KeyFilebrowserConfirmSelection: {
			Title:   "Confirm Selection",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyCtrlS, ' ', tcell.ModCtrl},
		},
		KeyFilebrowserDirForward: {
			Title:   "Go Forward",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyFilebrowserDirBack: {
			Title:   "Go Back",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyFilebrowserSelect: {
			Title:   "Select",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, ' ', tcell.ModNone},
		},
		KeyFilebrowserInvertSelection: {
			Title:   "Invert Selection",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'a', tcell.ModNone},
		},
		KeyFilebrowserSelectAll: {
			Title:   "Select All",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'A', tcell.ModNone},
		},
		KeyFilebrowserRefresh: {
			Title:   "Refresh",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyCtrlR, ' ', tcell.ModCtrl},
		},
		KeyFilebrowserToggleHidden: {
			Title:   "Hidden",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'h', tcell.ModCtrl},
		},
		KeyProgressTransferResume: {
			Title:   "Resume Transfer",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'g', tcell.ModNone},
		},
		KeyProgressTransferCancel: {
			Title:   "Cancel Transfer",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'x', tcell.ModNone},
		},
		KeyProgressView: {
			Title:   "View Downloads",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'v', tcell.ModNone},
		},
		KeyProgressTransferSuspend: {
			Title:   "Suspend Transfer",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'z', tcell.ModNone},
		},
	}
}
