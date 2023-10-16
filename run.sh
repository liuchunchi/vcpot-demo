#!/bin/bash
set -e

SESSION="demo-session"
PROG_SESSION="demo-program"
PIPES_NUM=9 # do not change this value
PIPES_DIR="./run"

function command_exists() {
  command -v -- "$1" &>/dev/null
}

function ensure_bash() {
  if [ ! -x /bin/bash ]; then
    echo /bin/bash not found or not executable. 1>&2
    exit 1
  fi
}

function ensure_commands_exists() {
  [ $# -eq 0 ] && return
  if command_exists "$1"; then
    shift
    ensure_commands_exists "$@"
  else
    echo Command "$1" is missing. 1>&2
    exit 1
  fi
}

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

# note: defining this function brings a bug that this script is unable to detect whether `tmux` is installed
function tmux() {
  command tmux -f ./tmux.conf "$@"
}

function main() {
  ensure_script_dir
  ensure_bash # make sure /bin/bash exists. This hard-coded path is specified in our tmux's configuration file.
  ensure_commands_exists bash tmux mkfifo dd cat sleep rm go
  mkdir -p -- "$PIPES_DIR" || true

  # prepare FIFO files
  for i in $(seq 0 $(("$PIPES_NUM" - 1))); do
    rm -- "$PIPES_DIR/$i.fifo.in" || true
    mkfifo -- "$PIPES_DIR/$i.fifo.in"
    rm -- "$PIPES_DIR/$i.fifo.out" || true
    mkfifo -- "$PIPES_DIR/$i.fifo.out"
  done

  # prepare commands to relay FIFO files
  commands=()
  for i in $(seq 0 $(("$PIPES_NUM" - 1))); do
    commands+=("echo [IO-$i]; cat -- $PIPES_DIR/$i.fifo.out & dd -- of=$PIPES_DIR/$i.fifo.in conv=sync; echo [IO-$i] terminated; sleep infinity")
  done

  # set up demonstration window
  tmux new-session -d -s "$SESSION" "${commands[0]}"
  tmux split-window -p 66 -h -t "$SESSION" "${commands[1]}"
  tmux split-window -p 50 -h -t "$SESSION" "${commands[2]}"
  for i in $(seq 0 2); do
    tmux select-pane -t "$SESSION:0.$((3 * "$i"))"
    tmux split-window -p 66 -v -t "$SESSION" "${commands["$i" + 3]}"
    tmux split-window -p 50 -v -t "$SESSION" "${commands["$i" + 6]}"
  done
  tmux select-pane -t "$SESSION:0.0"

  # run the program
  tmux new-session -s "$PROG_SESSION" -d "/usr/local/go/bin/go run main.go; echo Program terminated; sleep infinity"

  # show demonstration window
  tmux attach-session -t "$SESSION"
}

main
