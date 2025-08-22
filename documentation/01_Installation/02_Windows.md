# Requirements
The `haraltd` daemon is required to be downloaded and started.
Download it from [here](https://github.com/bluetuith-org/haraltd).
Once downloaded and extracted from the archive:
- Double-click on the `haraltd.exe` executable
- A SmartScreen window will pop-up, since the executable is currently not signed. Press "Allow anyways".
- A UAC popup will show, press "Yes"
- The daemon will launch, a notification will pop-up and an icon will be displayed in the taskbar.
- To stop the daemon, right-click the same icon, and press the "Stop" option.

Alternatively, to start it from Powershell or CMD.exe, type:
```ps
<path-to-downloaded-executable>\haraltd server start
```

# Installation
To install and run **bluetuith**:
- Download a release matching the architecture of the operating system, extract the archive
- Open Powershell or CMD, and run: `bluetuith`