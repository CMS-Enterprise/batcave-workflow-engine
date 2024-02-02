FROM artifactory.cloud.cms.gov/docker/golang:alpine3.19 as build

WORKDIR /app/src

COPY go.* .

# pre-fetch dependencies
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN mkdir -p ../bin && \
    go build -o ../bin/workflow-engine ./cmd/workflow-engine

FROM artifactory.cloud.cms.gov/batcave-docker/devops-pipelines/pipeline-tools/omnibus:v1.0.0

# Install docker and podman CLIs
RUN apk update && apk add --no-cache docker-cli-buildx podman

COPY --from=build /app/bin/workflow-engine /usr/local/bin/workflow-engine

ENTRYPOINT ["workflow-engine"]

LABEL org.opencontainers.image.title="workflow-engine"
LABEL org.opencontainers.image.description="A standalone CD engine for BatCAVE"
LABEL io.artifacthub.package.readme-url="https://github.com/CMS-Enterprise/batcave-workflow-engine/blob/main/README.md"
