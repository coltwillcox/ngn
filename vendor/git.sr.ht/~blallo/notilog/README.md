# Notilog

This is a small daemon to log notifications.

*NOTE*: The implementation of the storage used by default is in-memory, so the daemon will lose all the backlog as soon as it is restarted.

It is meant to be used with a [client](./cmd/notilogctl).

## Build

There is a make target.

```
make build
```

If you have problems with the dependencies, you might need to

```
export GOPRIVATE=git.sr.ht
```

## Install

There is a simple makefile target `install` that assumes that you are on a archlinux-like distro.
It builds and installs the daemon and cli to `/usr/bin` and a systemd unit to `/usr/lib/systemd/user/notilogd.service`.

To enable the unit

```
systemctl --user enable notilogd.service
```

## Usage

Both the daemon and the cli read the configuration from 4 places, with the following order:

  - default values backed into the code
  - configuration file
  - environment variables
  - flags on the command line

The configuration file has a simple `key=value` structure, where the key is in snake_case.
The corresponding environment variable is SCREAMING_SNAKE_CASE.
The flags are in kebab-case.

The configuration is, by default, read from `$XDG_CONFIG_DIR/notilog` and is `notilogd.conf` for the daemon and `notilog.conf` for the cli.

The daemon has a somehow simple configuration space:

```
Usage of notilogd:
  -config string
        Path to the config file
  -json-logs
        Format logs as json
  -log-level string
        Set logger to the specified level
  -socket-path string
        Path to the control socket
  -sqlite-storage string
        Path to the sqlite db to save the notifications to (if this is set, this storage will be used in place of the in-memory one)
```

`log-level` may be any of (case insensitive) `trace`, `debug`, `info`, `warning`, `error`.

The cli has 4 subcommands:

```
Usage: notilogctl [globalOpts] SUBCMD

globalOpts:
        -config Path to the config file
        -log-level Set the log level
        -socket-path Path to the control socket

Subcmds:
        get
        prune
        restart
        stop
        help

Type notilogctl [SUBCMD] --help for more help
```

Use `get` to retrieve the logged notifications:

```
$ notilogctl get
notify-send | notification | one   | 2024-03-16T20:05:14+01:00
notify-send | notification | two   | 2024-03-16T20:05:17+01:00
notify-send | notification | three | 2024-03-16T20:05:20+01:00
```

You can limit the retrieved notifications with `--tail <n>`, filter by program with `--program <prog-name>` and by time interval since now with `--since <duration>` (the duration here follows golang parsing rules).

Use `prune` to empty the notification store, `restart` to restart the daemon and `stop` to make it exit.
