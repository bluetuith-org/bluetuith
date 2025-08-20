[![Go Report Card](https://goreportcard.com/badge/github.com/darkhz/bluetuith)](https://goreportcard.com/report/github.com/darkhz/bluetuith) [![Packaging status](https://repology.org/badge/tiny-repos/bluetuith.svg)](https://repology.org/project/bluetuith/versions)

![demo](demo/demo.gif)

# bluetuith
bluetuith is a TUI-based bluetooth connection manager, which can interact with bluetooth adapters and devices.
It aims to be a replacement to most bluetooth managers, like blueman.

This is only available on Linux.

This project is currently in the alpha stage.

## Project status
This project has currently been confirmed to be sponsored by the [NLnet](https://nlnet.nl/project/bluetuith/) foundation.
The draft is complete, and the MoU has been signed. The work is now in progress.

Although this repo seems to be currently inactive, please bear in mind that we are actively working on new features, namely:
- Cross-platform support (Windows, MacOS, FreeBSD)
    - Shims[1] for Windows and MacOS
    - Cross platform daemon[2] with a unified API, for any bluetooth app to function across OSes.

- Updating and adding more UI features.
- Extensively refactoring the documentation.

[1]:
A shim is a lightweight application which can wrap native APIs and provide an API to invoke various bluetooth functions.
For every function, such as connection or pairing, the caller will invoke a separate process and execute the required function.
The caller will be responsible for handling events and exit codes passed by the shim.

[2]:
A cross-platform daemon with a uniform protocol (currently MQTT) will be developed to facilitate communicating with the shim, and handle
invoking bluetooth functions and communication with clients, using a proper pub/sub like mechanism (like DBus for example).

#### Updates
All development has now moved to [bluetuith-org](https://github.com/bluetuith-org).<br />
This project will be moved to the organisation later.

- A new Windows-based shim has been released at [bluetuith-shim-windows](https://github.com/bluetuith-org/bluetuith-shim-windows).

[![Packaging status](https://repology.org/badge/vertical-allrepos/bluetuith.svg)](https://repology.org/project/bluetuith/versions)

## Features
- Transfer and receive files via OBEX.
- Perform pairing with authentication.
- Connect to/disconnect from different devices.
- Interact with bluetooth adapters, toggle power and discovery states
- Connect to or manage Bluetooth based networking/tethering (PANU/DUN)
- Remotely control media playback on the connected device
- Mouse support

## Documentation
The documentation is now hosted [here](https://darkhz.github.io/bluetuith)

The wiki is out-of-date.
