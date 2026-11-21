# Creating Your Own Command Lists

Command Lists lets you organize commands into separate tabs using
Markdown files.

## Create a new list

Every `.md` file inside the `lists/` directory becomes a separate tab.

``` bash
nano "$(herdr plugin config-dir herdr.command-lists)/lists/Firewall.md"
```

The filename determines the tab name.

## List format

``` markdown
# Network
ip -br addr
ip route

# Firewall
firewall-cmd --state
firewall-cmd --list-all
```

-   Lines starting with `#` are section headings.
-   Blank lines add spacing.
-   Every other non-empty line is an executable command.
-   Headings cannot be selected or executed.

## Change tab order

Edit `tab-order.txt`:

``` bash
nano "$(herdr plugin config-dir herdr.command-lists)/tab-order.txt"
```

Add one filename per line:

``` text
Firewall.md
System.md
SSH.md
```

You can also reorder tabs directly inside the popup using `Ctrl+P` and
`Ctrl+N`.

Lists not specified in the file appear afterward in alphabetical order.

## Important

Changes to command lists take effect the next time the popup opens. No
rebuild or restart is required.

Commands are executed with your user permissions in the original Herdr
pane. Only add commands you trust.
