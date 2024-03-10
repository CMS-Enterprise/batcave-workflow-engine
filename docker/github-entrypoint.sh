#! /usr/bin/env sh

set -e

chgrp -R podman "$GITHUB_WORKSPACE"
chmod -R g+w "$GITHUB_WORKSPACE"

# Store podman containers in a persistent location
sed -i -e "s|^#\\? *rootless_storage_path *=.*$|rootless_storage_path=\"$GITHUB_WORKSPACE/.containers/storage\"|" /etc/containers/storage.conf

# It is important that each argument be quoted because empty strings are significant
QUOTED_ARGS=""
for arg in "$@"; do
  QUOTED_ARGS="$QUOTED_ARGS \"$arg\""
done

# su requires the command to be a single argument, not a series
su podman -s "$(which sh)" -c "workflow-engine $QUOTED_ARGS"
