#!/bin/bash
LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/glacials/mainframe/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
TARFILE="mainframe-$LATEST_VERSION-linux-arm.tar.gz"
ARTIFACT_URL="https://github.com/glacials/mainframe/releases/download/$LATEST_VERSION/mainframe-$LATEST_VERSION-linux-arm.tar.gz"

if [[ $(mainframe --version) == "$LATEST_VERSION" ]]
then
  echo "Already running latest version; exiting"
  exit
fi

mkdir -p ~/bin

rm -f $TARFILE # Just in case
wget $ARTIFACT_URL
tar xzf $TARFILE # Extracts a binary called "mainframe"
rm -f $TARFILE

mv mainframe ~/bin
pkill mainframe
source ~/.envrc
~/bin/mainframe & > /var/log/mainframe.log

cd ~/pj/mainframe
git pull
