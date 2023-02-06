#!/bin/bash

# set -euxo pipefail # Echo commands
# set -euo pipefail  # Don't echo commands
set -euo pipefail

BINARY_FILENAME=mainframe # NOT a path
TMUX_SESSION_NAME=mainframe
TMUX_WINDOW_NAME=mainframe

function boot () {
  echo "Booting mainframe..."
  THROWAWAY_WINDOW_NAME=deleteme # We'll delete this after setting env vars because it won't inherit them
  tmux new-session -d -n $THROWAWAY_WINDOW_NAME -s $TMUX_SESSION_NAME > /dev/null 2>&1 || echo "  Using existing tmux session: $TMUX_SESSION_NAME"
  tmux new-window -n $THROWAWAY_WINDOW_NAME > /dev/null 2>&1 || echo "  Using existing tmux window: $TMUX_WINDOW_NAME"
  cat .envrc | while read line
  do
    IFS="=" read key val <<< $line
    tmux set-environment -t mainframe $key $val
  done
  TMUX_WINDOW_NAME=mainframe
  tmux new-window -d -n $TMUX_WINDOW_NAME
  tmux unlink-window -k -t $THROWAWAY_WINDOW_NAME > /dev/null 2>&1 || echo "  Throwaway session: gone"
  tmux send-keys -t $TMUX_WINDOW_NAME mainframe C-m

  sleep 2
  if ps aux | grep -v grep | grep mainframe > /dev/null
  then
    echo "  Success: yes"
  else
    echo "  Success: no"
    echo "  Error: Can't boot mainframe."
    exit 1
  fi
}

echo "Mainframe supervisor starting."
echo "Checking status..."
echo "  Installed: yes"

if ps aux | grep -v grep | grep -c mainframe > /dev/null
then
  echo "  Running:   yes"
else
  echo "  Running:   no"
fi

boot

echo "Checking version..."

uname=$(uname -s | tr "[:upper:]" "[:lower:]")
if [[ $uname == linux* ]]; then
  platform="linux"
elif [[ $uname == darwin* ]]; then
  platform="darwin"
elif [[ $uname == msys* ]]; then
  platform="windows"
else
  echo "  Error: Unknown platform."
  exit 1
fi

arch=$(uname -m) # RPi gives "arm"; macOS Apple Silicon gives "arm64"

MAINFRAME_LOCAL=~/bin/mainframe
LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/glacials/mainframe/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
TARFILE="mainframe-$LATEST_VERSION-$platform-$arch.tar.gz"
ARTIFACT_URL="https://github.com/glacials/mainframe/releases/download/$LATEST_VERSION/$TARFILE"

echo "  Current:        $($MAINFRAME_LOCAL --version)"
echo "  Latest:         $LATEST_VERSION"

if [[ "$($MAINFRAME_LOCAL --version)" == "$LATEST_VERSION" ]]
then
  echo "  Upgrade needed: no"
  echo "Supervisor completed."
  exit 0
else
  echo "  Upgrade needed: yes"
fi

echo "Upgrading mainframe..."

mkdir -p ~/bin

rm -f $TARFILE # Just in case
wget $ARTIFACT_URL
tar xzf $TARFILE # Extracts a binary called "mainframe"
rm -f $TARFILE

mv mainframe $MAINFRAME_LOCAL
pkill mainframe # If this fails, mainframe wasn't running but we expected it to
boot
cd ~/pj/mainframe
git pull

echo "  Success: yes"
echo "Mainframe supervisor completed. To attach:"
echo ""
echo "    tmux attach -t mainframe"
