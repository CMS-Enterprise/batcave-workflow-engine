#! /usr/bin/env sh

set -e

# It is important that each argument be quoted because empty strings are significant
QUOTED_ARGS=""
for arg in "$@"; do
  QUOTED_ARGS="$QUOTED_ARGS \"$arg\""
done

# su requires the command to be a single argument, not a series
su podman -s "$(which sh)" -c "workflow-engine $QUOTED_ARGS"
