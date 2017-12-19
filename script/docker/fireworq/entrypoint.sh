#!/bin/sh

DIR=$(cd "$(dirname "$0")"; pwd)
ROOT=$(cd "$DIR/../../.."; pwd)
cd "$ROOT"

. "$DIR/env.sh"
make build

export FIREWORQ_QUEUE_DEFAULT="${FIREWORQ_QUEUE_DEFAULT:-default}"
exec ./fireworq
