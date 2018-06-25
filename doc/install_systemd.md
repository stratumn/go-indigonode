# Systemd Installation

Make sure Indigo Node is installed somewhere in your `$PATH`.

## Configure Logging For Journald

Edit `indigonode.core.toml` so that the `log` section looks like this:

```toml
[log]

  [[log.writers]]

    # The formatter for the writer (json, text, color, journald).
    formatter = "journald"

    # The log level for the writer (info, error, all).
    level = "all"

    # The type of writer (file, stdout, stderr).
    type = "stderr"
```

You should also disable the boot screen in the `core` section:

```toml
[core]

  # Whether to show the boot screen when starting the node.
  enable_boot_screen = false
```

## Setup The Indigo Node Daemon

Install the systemd unit file:

```bash
$ indigo-node daemon install --core-config /path/to/indigonode.core.toml
```

Tell systemd to reload the unit files:

```bash
$ systemctl daemon-reload
```

Start the Indigo Node service:

```bash
$ systemctl start indigo-node
```

You can stop it with:

```bash
$ systemctl stop indigo-node
```

## View Logs

To view all the streaming logs, run:

```bash
$ journalctl -f -u indigo-node
```

To view only errors, run (errors have a priority of three):

```bash
$ journalctl -f -u indigo-node -p 3
```
