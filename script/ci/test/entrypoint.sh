#!/bin/sh

DIR=$(cd "$(dirname "$0")"; pwd)
ROOT=$(cd "$DIR/../../.."; pwd)
cd "$ROOT"

. "$ROOT/script/docker/fireworq/env.sh"

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
gosu docker make lint
lint_status=$?
gosu docker make test TEST_OUTPUT=/workspace
test_status=$?
gosu docker make cover TEST_OUTPUT=/workspace

[ $lint_status = 0 ] || exit $lint_status
[ $test_status = 0 ] || exit $test_status
