# Systemd Installation

Make sure Alice is installed somewhere in your `$PATH`.

## Configure Logging For Journald

Edit `alice.core.toml` so that the `log` section looks like this:

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

## Setup The Alice Daemon

Install the systemd unit file:

```bash
$ alice daemon install --core-config /path/to/alice.core.toml
```

Tell systemd to reload the unit files:

```bash
$ systemctl daemon-reload
```

Start the Alice service:

```bash
$ systemctl start alice
```

You can stop it with:

```bash
$ systemctl stop alice
```

## View Logs

To view all the streaming logs, run:

```bash
$ journalctl -f -u alice
```

To view only errors, run (errors have a priority of three):

```bash
$ journalctl -f -u alice -p 3
```

