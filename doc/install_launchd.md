# Launchd Installation

Make sure Indigo Node is installed somewhere in your `$PATH`.

## Configure Logging For Launchd

Edit `indigo_node.core.toml` so that the `log` section looks like this:

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

## Setup The Indigo Node Daemon

Install the property list file:

```bash
$ indigo-node daemon install --core-config /path/to/indigo_node.core.toml
```

Start the Indigo Node service:

```bash
$ sudo launchctl load /Library/LaunchDaemons/indigo-node.plist
```

You can stop it with:

```bash
$ sudo launchctl unload /Library/LaunchDaemons/indigo-node.plist
```

## View Logs

To view the streaming info logs, run:

```bash
$ indigo-node log -f /usr/local/var/log/indigo-node.log
```

To view the streaming error logs, run:

```bash
$ indigo-node log -f /usr/local/var/log/indigo-node.err
```
