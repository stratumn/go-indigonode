# Launchd Installation

Make sure Alice is installed somewhere in your `$PATH`.

## Configure Logging For Launchd

Edit `alice.core.toml` so that the `log` section looks like this:

```toml
[log]

  [[log.writers]]

    # The formatter for the writer (json, text, color, journald).
    formatter = "json"

    # The log level for the writer (info, error, all).
    level = "error"

    # The type of writer (file, stdout, stderr).
    type = "stderr"

  [[log.writers]]

    # The formatter for the writer (json, text, color, journald).
    formatter = "json"

    # The log level for the writer (info, error, all).
    level = "info"

    # The type of writer (file, stdout, stderr).
    type = "stdout"
```

You should also disable the boot screen in the `core` section:

```toml
[core]

  # Whether to show the boot screen when starting the node.
  enable_boot_screen = false
```

## Setup The Alice Daemon

Install the property list file:

```bash
$ alice daemon install --core-config /path/to/alice.core.toml
```

Start the Alice service:

```bash
$ sudo launchctl load /Library/LaunchDaemons/alice.plist
```

You can stop it with:

```bash
$ sudo launchctl unload /Library/LaunchDaemons/alice.plist
```

## View Logs

To view the streaming info logs, run:

```bash
$ alice log -f /usr/local/var/log/alice.log 
```

To view the streaming error logs, run:

```bash
$ alice log -f /usr/local/var/log/alice.err 
```

