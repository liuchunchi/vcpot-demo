#!/bin/bash
set -e

SESSION="demo-session"
PROG_SESSION="demo-program"
PIPES_NUM=9 # do not change this value
PIPES_DIR="./run"

function get_script_dir() {
  (cd "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
}

function ensure_script_dir() {
  local script_dir
  script_dir="$(get_script_dir)"
  local working_dir
  working_dir="$(pwd)"
  if [ "$script_dir" != "$working_dir" ]; then
    echo Run this script in "$script_dir". 1>&2
    exit 1
  fi
}

function tmux() {
  command tmux -f ./tmux.conf "$@"
}

function main() {
  ensure_script_dir

  tmux kill-session -t "$SESSION" || true
  tmux kill-session -t "$PROG_SESSION" || true

  for i in $(seq 0 $(("$PIPES_NUM" - 1))); do
    rm -- "$PIPES_DIR/$i.fifo.in" || true
    rm -- "$PIPES_DIR/$i.fifo.out" || true
  done
}

main