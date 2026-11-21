# Edit List
nano "$(herdr plugin config-dir herdr.command-lists)/lists/Firewall.md"

# UFW (Ubuntu / Debian)
ufw status verbose

# firewalld (Fedora / RHEL)
firewall-cmd --state
firewall-cmd --get-active-zones
firewall-cmd --list-all

# nftables (Linux)
nft list ruleset

# PF (macOS)
pfctl -s info
pfctl -s rules
