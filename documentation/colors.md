# Colors support

DNSControl is using colors on both its console and notification outputs, to
denote changes, such as during preview or push events.

This is on by default.

## Disable colors

There are cases that we might need to disable colored output, such as in certain
CI environments or particular notification mechanisms that don't support ANSI color
codes.

In order to do so, a global `--no-colors` command option is provided, which when
set, will disable colors globally.
