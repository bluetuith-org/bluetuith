# Linux
- Ensure that the bluetooth service is up and running, and it is visible to DBus before launching the application. With systemd you can find out the status using the following command: `systemctl status bluetooth.service`.
- Check whether the OBEX service is running as well using `systemctl status obexd` and if not running, start it using `systemctl start obexd`.
- Only one transfer (either of send or receive) can happen on an adapter. Attempts to start another transfer while a transfer is already running (for example, trying to send files to a device when a transfer is already in progress) will be silently ignored.

# Windows
- Check whether the `haraltd` daemon is running, and if not, run the following command from the folder where `haraltd` was downloaded:
`haraltd server start`
- The daemon will currently not start if no adapter are present in the system.