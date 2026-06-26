package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	scfg "github.com/bluetuith-org/bluetooth-classic/api/config"
	"github.com/bluetuith-org/bluetooth-classic/session"
	"github.com/darkhz/bluetuith/ui/app"
	"github.com/darkhz/bluetuith/ui/config"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v2"
)

// These values are set at compile-time.
var (
	Version  = ""
	Revision = ""
)

// Run runs the commandline application.
func Run() error {
	return newApp().Run(os.Args)
}

// newApp returns a new commandline application.
func newApp() *cli.App {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Fprintf(cCtx.App.Writer, "%s (%s)\n", Version, Revision)
	}

	return &cli.App{
		Name:                   "bluetuith",
		Usage:                  "Bluetooth Manager.",
		Version:                Version + " (" + Revision + ")",
		Description:            "A Bluetooth manager for the terminal.",
		DefaultCommand:         "bluetuith",
		Copyright:              "(c) bluetuith-org.",
		Compiled:               time.Now(),
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Suggest:                true,
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:    "list-adapters",
				Aliases: []string{"l"},
				Usage:   "List available adapters.",
				Action: func(*cli.Context, bool) error {
					var sb strings.Builder

					s := session.NewSession()
					_, _, err := s.Start(nil, scfg.New())
					if err != nil {
						return err
					}
					defer s.Stop()

					adapters, err := s.Adapters()
					if err != nil {
						return err
					}

					sb.WriteString("List of adapters:")
					for _, adapter := range adapters {
						sb.WriteString("\n")
						sb.WriteString("- ")
						sb.WriteString(getAdapterDisplayName(adapter))
					}

					fmt.Println(sb.String())

					return nil
				},
			},
			&cli.StringFlag{
				Name:    "adapter",
				Aliases: []string{"a"},
				EnvVars: []string{"BLUETUITH_ADAPTER"},
				Usage:   "Specify an adapter to use. (For example, hci0)",
			},
			&cli.StringFlag{
				Name:    "receive-dir",
				Aliases: []string{"r"},
				EnvVars: []string{"BLUETUITH_RECEIVE_DIR"},
				Usage:   "Specify a directory to store received files.",
			},
			&cli.StringFlag{
				Name:    "gsm-apn",
				Aliases: []string{"m"},
				EnvVars: []string{"BLUETUITH_GSM_APN"},
				Usage:   "Specify GSM APN to connect to. (Required for DUN)",
			},
			&cli.StringFlag{
				Name:    "gsm-number",
				Aliases: []string{"b"},
				EnvVars: []string{"BLUETUITH_GSM_NUMBER"},
				Usage:   "Specify GSM number to dial. (Required for DUN)",
			},
			&cli.StringFlag{
				Name:    "adapter-states",
				Aliases: []string{"s"},
				EnvVars: []string{"BLUETUITH_ADAPTER_STATES"},
				Usage:   "Specify adapter states to enable/disable. (For example, 'powered:yes,discoverable:yes,pairable:yes,scan:no')",
			},
			&cli.StringFlag{
				Name:    "connect-bdaddr",
				Aliases: []string{"t"},
				EnvVars: []string{"BLUETUITH_CONNECT_BDADDR"},
				Usage:   "Specify device address to connect (For example, 'AA:BB:CC:DD:EE:FF')",
			},
			&cli.BoolFlag{
				Name:    "no-warning",
				Aliases: []string{"w"},
				EnvVars: []string{"BLUETUITH_NO_WARNING"},
				Usage:   "Do not display warnings when the application has initialized.",
			},
			&cli.BoolFlag{
				Name:    "no-help-display",
				Aliases: []string{"i"},
				EnvVars: []string{"BLUETUITH_NO_HELP_DISPLAY"},
				Usage:   "Do not display help keybindings in the application.",
			},
			&cli.BoolFlag{
				Name:    "confirm-on-quit",
				Aliases: []string{"c"},
				EnvVars: []string{"BLUETUITH_CONFIRM_ON_QUIT"},
				Usage:   "Ask for confirmation before quitting the application.",
			},
			&cli.BoolFlag{
				Name:    "disable-obex-services",
				Aliases: []string{"o"},
				EnvVars: []string{"BLUETUTITH_ENABLE_OBEX_SERVICES"},
				Usage:   "Specify whether to disable OBEX services (like Object Push Transfers)",
			},
			&cli.BoolFlag{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "Generate configuration.",
				Action: func(cliCtx *cli.Context, _ bool) error {
					k := koanf.New(".")

					cliCtx.Command.Name = "global"

					conf := config.NewConfig()
					if err := conf.Load(k, cliCtx); err != nil {
						return err
					}

					oldcfgparsed, err := conf.GenerateAndSave(k)
					if !oldcfgparsed {
						printWarn("the old configuration could not be parsed")
					}

					return err
				},
			},
		}, getPlatformSpecificFlags()...),
		Action: func(cliCtx *cli.Context) error {
			if cliCtx.Bool("list-adapters") || cliCtx.Bool("generate") {
				return nil
			}

			// required for koanf to merge all global flags under the root namespace.
			cliCtx.Command.Name = "global"

			k, cfg := koanf.New("."), config.NewConfig()
			if err := cfg.Load(k, cliCtx); err != nil {
				return err
			}
			if err := cfg.ValidateValues(); err != nil {
				return err
			}

			sessionCfg := scfg.New()
			populateSessionConfig(cliCtx, &sessionCfg)

			app, s := app.NewApplication(), session.NewSession()
			featureSet, _, err := s.Start(app.Authorizer(), sessionCfg)
			if err != nil {
				return err
			}
			defer s.Stop()

			if err := cfg.ValidateSessionValues(s); err != nil {
				return err
			}
			printUnsupportedFeatures(cfg, featureSet)

			return app.Start(s, featureSet, cfg)
		},
		ExitErrHandler: func(_ *cli.Context, err error) {
			if err == nil {
				return
			}

			printError(err)
		},
	}
}

// printUnsupportedFeatures prints all unsupported features of the session.
func printUnsupportedFeatures(cfg *config.Config, featureSet *appfeatures.FeatureSet) {
	if cfg.Values.NoWarning {
		return
	}

	var warn strings.Builder

	featErrors, exists := featureSet.Errors.Exists()
	if !exists {
		return
	}

	warn.WriteString("The following features are not available:")
	for feature, errors := range featErrors {
		warn.WriteString("\n")
		warn.WriteString(feature.String())
		warn.WriteString(": ")
		warn.WriteString(errors.Error())
	}

	printWarn(warn.String())
	time.Sleep(1 * time.Second)
}

func populateSessionConfig(cliCtx *cli.Context, sessionCfg *scfg.Configuration) {
	sessionCfg.EnableObexServices = true
	if cliCtx.Bool("disable-obex-services") {
		sessionCfg.EnableObexServices = false
	}

	sessionCfg.LibraryPath = cliCtx.String("alt-library-path")
	sessionCfg.SocketPath = cliCtx.String("alt-daemon-socket-path")
}

// getAdapterDisplayName returns the display name of the adapter.
func getAdapterDisplayName(adapterData bluetooth.AdapterData) string {
	if name, ok := adapterData.Name.Get(); ok {
		return name
	}

	if adapterData.UniqueName != "" {
		return adapterData.UniqueName
	}

	return adapterData.Address.String()
}
