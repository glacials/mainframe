#!/bin/bash

# set -euxo pipefail
set -exo pipefail

echo "Supervisor starting."
echo "Checking mainframe status..."
echo "  Installed: yes"

if pidof mainframe > /dev/null
then
  echo "  Running: yes"
else
  echo "  Running: no"

  echo "Booting mainframe..."
  tmux new -ds mainframe || echo "  Using existing tmux session"
  tmux send-keys -t mainframe mainframe C-m
  sleep 2
  if pidof mainframe > /dev/null
  then
    echo "  Success: yes"
  else
    echo "  Success: no"
    echo "  Error: Can't boot mainframe."
    exit 1
  fi
fi

echo "Checking mainframe versions..."

MAINFRAME_LOCAL=~/bin/mainframe
LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/glacials/mainframe/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
TARFILE="mainframe-$LATEST_VERSION-linux-arm.tar.gz"
ARTIFACT_URL="https://github.com/glacials/mainframe/releases/download/$LATEST_VERSION/mainframe-$LATEST_VERSION-linux-arm.tar.gz"

echo "  Current: $($MAINFRAME_LOCAL --version)"
echo "  Latest:  $LATEST_VERSION"

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
source ~/.envrc
tmux new -ds mainframe || echo "  Using existing tmux session"
tmux send-keys -t mainframe mainframe C-m

cd ~/pj/mainframe
git pull

echo "  Success: yes"
echo "Supervisor completed."
