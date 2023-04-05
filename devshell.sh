#!/usr/bin/env bash

set -e
set -o pipefail

if [ $# != 0 ] ; then
	echo "Usage: $0"
	echo "Starts a local development shell where the build can be run as close as possible to the CI"
	exit 1
fi

DOCKER_PLATFORM="linux/amd64"

GID="$(id -g)"
GROUP="$(getent group $(getent passwd $USER | cut -d: -f4) | cut -d: -f1)"
GO_VERSION="$(awk '/^go /{print $2}' go.mod)"

GIT_ROOT="$(cd $(dirname $0) && git rev-parse --show-toplevel)"
if ! test -d "$GIT_ROOT"/.cache ; then
	mkdir "$GIT_ROOT"/.cache
fi

NAME="rrb-$$"

function kill_container() {
	docker kill --signal SIGKILL "${NAME}" &>/dev/null || true
}

trap kill_container EXIT

docker run \
	--name "${NAME}" \
	--platform ${DOCKER_PLATFORM} \
	--rm \
	--tty \
	--interactive \
	--volume ${GIT_ROOT}:${HOME}/rrb \
	--volume ${GIT_ROOT}/.cache:${HOME}/.cache \
	--volume ${HOME}/rrb/.cache \
	--workdir ${HOME}/rrb \
	golang:${GO_VERSION}-bullseye \
	/bin/bash -c "$(cat <<EOF
set -e

# User
passwd -d root > /dev/null
addgroup --gid ${GID} ${GROUP} > /dev/null
useradd --home-dir ${HOME} --gid ${GID} --no-create-home --shell /bin/bash --uid ${UID} ${USER}
ln -s ${HOME}/rrb/.bashrc ${HOME}/.bashrc
chown ${UID}:${GID} ${HOME}

# Shell
exec su --group ${GROUP} --pty ${USER} sh -c "
	set -e

	export BINDIR=\\\$(make BINDIR)
	PATH=\\\$BINDIR:\\\$PATH
	export GOBIN=\\\$(make GOBIN)
	PATH=\\\$GOBIN:\\\$PATH
	export GOCACHE=\\\$(make GOCACHE)
	export GOMODCACHE=\\\$(make GOMODCACHE)
  
	echo Available make targets:
	make help
	exec bash -i"
EOF
)"