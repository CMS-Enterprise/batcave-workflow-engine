FROM golang:alpine3.19 as build

ARG VERSION
ARG GIT_COMMIT
ARG GIT_DESCRIPTION

# install build dependencies
RUN apk update && apk add git --no-cache

WORKDIR /app/src

COPY go.* .

# pre-fetch dependencies
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN mkdir -p ../bin && \
    go build -ldflags="-X 'main.cliVersion=${VERSION}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.gitDescription=${GIT_DESCRIPTION}'" -o ../bin/workflow-engine ./cmd/workflow-engine

FROM ghcr.io/cms-enterprise/batcave/omnibus:v1.5.1 as workflow-engine-base

COPY --from=build /app/bin/workflow-engine /usr/local/bin/workflow-engine

# enable the Gatecheck beta CLI
ENV GATECHECK_FF_CLI_V1_ENABLED=1

ENTRYPOINT ["workflow-engine"]

LABEL org.opencontainers.image.title="workflow-engine-docker"
LABEL org.opencontainers.image.description="A standalone CD engine for BatCAVE"
LABEL org.opencontainers.image.licenses="CC0-1.0"
LABEL io.artifacthub.package.readme-url="https://github.com/CMS-Enterprise/batcave-workflow-engine/blob/main/README.md"
LABEL io.artifacthub.package.license="CC0-1.0"

FROM workflow-engine-base as workflow-engine-podman

# Install podman CLIs
RUN apk update && apk add --no-cache podman fuse-overlayfs

COPY docker/storage.conf /etc/containers/
COPY docker/containers.conf /etc/containers/

RUN addgroup -S podman && adduser -S podman -G podman && \
    echo podman:10000:5000 > /etc/subuid && \
    echo podman:10000:5000 > /etc/subgid

COPY docker/rootless-containers.conf /home/podman/.config/containers/containers.conf

RUN mkdir -p /home/podman/.local/share/containers
RUN chown podman:podman -R /home/podman

VOLUME /var/lib/containers
VOLUME /home/podman/.local/share/containers

RUN mkdir -p /var/lib/clamav
RUN chown podman /var/lib/clamav && chown podman /etc/clamav

LABEL org.opencontainers.image.title="workflow-engine-podman"

FROM workflow-engine-base

# Install docker CLI
RUN apk update && apk add --no-cache docker-cli-buildx

LABEL org.opencontainers.image.title="workflow-engine-docker"
