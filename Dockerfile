FROM golang:alpine3.19 as build

WORKDIR /app/src

COPY go.* .

# pre-fetch dependencies
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN mkdir -p ../bin && \
    go build -o ../bin/workflow-engine ./cmd/workflow-engine

FROM ghcr.io/cms-enterprise/batcave/omnibus:v1.1.0-rc3

# Install docker and podman CLIs
RUN apk update && apk add --no-cache docker-cli-buildx podman fuse-overlayfs

COPY docker/storage.conf /etc/containers/
COPY docker/containers.conf /etc/containers/

COPY --from=build /app/bin/workflow-engine /usr/local/bin/workflow-engine

RUN addgroup -S podman && adduser -S podman -G podman && \
    echo podman:10000:5000 > /etc/subuid && \
    echo podman:10000:5000 > /etc/subgid

COPY docker/rootless-containers.conf /home/podman/.config/containers/containers.conf

RUN mkdir -p /home/podman/.local/share/containers
RUN chown podman:podman -R /home/podman

VOLUME /var/lib/containers
VOLUME /home/podman/.local/share/containers

ENTRYPOINT ["workflow-engine"]

LABEL org.opencontainers.image.title="workflow-engine"
LABEL org.opencontainers.image.description="A standalone CD engine for BatCAVE"
LABEL io.artifacthub.package.readme-url="https://github.com/CMS-Enterprise/batcave-workflow-engine/blob/main/README.md"
