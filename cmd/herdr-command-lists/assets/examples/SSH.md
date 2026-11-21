# Edit List
nano "$(herdr plugin config-dir herdr.command-lists)/lists/SSH.md"

# Client
ssh -V
ssh-add -l

# SSH Directory
ls -la ~/.ssh

# Client Config
test -f ~/.ssh/config && sed -n '1,160p' ~/.ssh/config || true

# Known Hosts
test -f ~/.ssh/known_hosts && wc -l ~/.ssh/known_hosts || true
