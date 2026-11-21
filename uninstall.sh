#!/bin/sh
set -eu
if [ "$#" -gt 1 ] || { [ "$#" -eq 1 ] && [ "$1" != "--keep-source" ] && [ "$1" != "-h" ] && [ "$1" != "--help" ]; }; then
    echo 'Usage: ./uninstall.sh [--keep-source]' >&2
    exit 2
fi
if [ "${1:-}" = '-h' ] || [ "${1:-}" = '--help' ]; then
    echo 'Usage: ./uninstall.sh [--keep-source]'
    exit 0
fi
source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
herdr_bin=${HERDR_BIN_PATH:-herdr}
if [ -x "$source_dir/command-lists" ]; then
    "$source_dir/command-lists" shortcut-config remove >/dev/null
fi
"$herdr_bin" plugin unlink herdr.command-lists >/dev/null
"$herdr_bin" server reload-config >/dev/null 2>&1 || true
printf '%s\n' 'Command Lists unlinked. Your .md lists and tab-order.txt have been kept.'
printf '%s\n' 'Source files are also kept; delete the source directory manually if desired.'
