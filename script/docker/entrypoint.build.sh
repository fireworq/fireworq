#!/bin/sh

set -xe

DIR=$(cd "$(dirname "$0")"; pwd)
ROOT=$(cd "$DIR/../.."; pwd)
cd "$ROOT"

[ -n "$GOOS" ] && [ -n "$GOARCH" ] && go install std

USER_ID=${USER_ID:-10000}
GROUP_ID=${GROUP_ID:-$USER_ID}

echo "Starting with UID: ${USER_ID}, GID: ${GROUP_ID}" >&2
groupadd -g $GROUP_ID --non-unique docker
useradd -u $USER_ID -g $GROUP_ID --non-unique docker

export HOME=/home/docker
mkdir -p $HOME
chown -R $USER_ID:$GROUP_ID $HOME
chown -R $USER_ID:$GROUP_ID /go
sync

gosu docker make clean
gosu docker make release BUILD_OUTPUT=/workspace PRERELEASE=''
gosu docker cp AUTHORS /workspace/
gosu docker cp CREDITS /workspace/
