# Launchd Installation

Make sure Stratumn Node is installed somewhere in your `$PATH`.

## Configure Logging For Launchd

Edit `stratumn_node.core.toml` so that the `log` section looks like this:

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

## Setup The Stratumn Node Daemon

Install the property list file:

```bash
stratumn-node daemon install --core-config /path/to/stratumn_node.core.toml
```

Start the Stratumn Node service:

```bash
sudo launchctl load /Library/LaunchDaemons/stratumn-node.plist
```

You can stop it with:

```bash
sudo launchctl unload /Library/LaunchDaemons/stratumn-node.plist
```

## View Logs

To view the streaming info logs, run:

```bash
stratumn-node log -f /usr/local/var/log/stratumn-node.log
```

To view the streaming error logs, run:

```bash
stratumn-node log -f /usr/local/var/log/stratumn-node.err
```
