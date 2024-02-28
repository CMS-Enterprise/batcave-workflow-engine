# TODO: semgrep

FROM golang:alpine as build

ARG GRYPE_VERSION=v0.73.4
ARG SYFT_VERSION=v0.98.0
ARG GITLEAKS_VERSION=v8.18.1
ARG COSIGN_VERSION=v2.2.2
ARG CRANE_VERSION=v0.17.0
ARG RELEASE_CLI_VERSION=v0.16.0
ARG GATECHECK_VERSION=v0.3.0
ARG S3UPLOAD_VERSION=v1.0.4
ARG ORAS_VERSION=v1.1.0

RUN apk --no-cache add ca-certificates git openssh make

WORKDIR /app

# Layer on purpose for caching improvements
RUN git clone --branch ${GRYPE_VERSION} --depth=1 --single-branch https://github.com/anchore/grype /app/grype
RUN cd /app/grype && \
    go build -ldflags="-w -s -extldflags '-static' -X 'main.version=${GRYPE_VERSION}' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.gitDescription=$(git log -1 --pretty=%B)'" -o /usr/local/bin ./cmd/grype
    
RUN git clone --branch ${SYFT_VERSION} --depth=1 --single-branch https://github.com/anchore/syft /app/syft 
RUN cd /app/syft && \
    go build -ldflags="-w -s -extldflags '-static' -X 'main.version=${SYFT_VERSION}' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.gitDescription=$(git log -1 --pretty=%B)'" -o /usr/local/bin ./cmd/syft
    
RUN git clone --branch ${GITLEAKS_VERSION} --depth=1 --single-branch https://github.com/zricethezav/gitleaks /app/gitleaks
RUN cd /app/gitleaks && \
    go build -ldflags="-s -w -X=github.com/zricethezav/gitleaks/v8/cmd.Version=${GITLEAKS_VERSION}" -o /usr/local/bin .

RUN git clone --branch ${COSIGN_VERSION} --depth=1 --single-branch https://github.com/sigstore/cosign /app/cosign
RUN cd /app/cosign && \
    make cosign && \
    mv cosign /usr/local/bin
   
RUN git clone --branch ${CRANE_VERSION} --depth=1 --single-branch https://github.com/google/go-containerregistry /app/go-containerregistry
RUN cd go-containerregistry && \
    go build -ldflags="-s -w -X github.com/google/go-containerregistry/cmd/crane/cmd.Version=${CRANE_VERSION}" -o /usr/local/bin ./cmd/crane

RUN git clone --branch ${RELEASE_CLI_VERSION} --depth=1 --single-branch https://gitlab.com/gitlab-org/release-cli /app/release-cli
RUN cd release-cli && \
    make build && \
    mv ./bin/release-cli /usr/local/bin
    
RUN git clone --branch ${GATECHECK_VERSION} --depth=1 --single-branch https://github.com/gatecheckdev/gatecheck /app/gatecheck
RUN cd gatecheck && \
    go build -ldflags="-s -w" -o /usr/local/bin ./cmd/gatecheck

RUN git clone --branch ${S3UPLOAD_VERSION} --depth=1 --single-branch https://github.com/bacchusjackson/go-s3-upload /app/go-s3-upload
RUN cd go-s3-upload && \
    go build -ldflags="-s -w" -o /usr/local/bin/s3upload .

RUN git clone --branch ${ORAS_VERSION} --depth=1 --single-branch https://github.com/oras-project/oras /app/oras
RUN cd oras && \
    make build-linux-amd64 && \
    mv bin/linux/amd64/oras /usr/local/bin/oras

FROM rust:alpine as build-just

RUN apk add musl-dev
RUN cargo install just

FROM alpine:latest as final-base

RUN apk --no-cache add curl jq sqlite-libs git ca-certificates tzdata clamav

WORKDIR /app

LABEL org.opencontainers.image.title="omnibus"
LABEL org.opencontainers.image.description="A collection of CI/CD tools for batCAVE"
LABEL org.opencontainers.image.vendor="nightwing"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL io.artifacthub.package.readme-url="https://code.batcave.internal.cms.gov/devops-pipelines/pipeline-tools/omnibus/-/blob/main/README.md"
LABEL io.artifacthub.package.license="Apache-2.0"

# Final image in a CI environment, assumes binaries are located in ./bin
FROM final-base as final-ci

COPY ./bin/grype /usr/local/bin/grype
COPY ./bin/syft /usr/local/bin/syft
COPY ./bin/gitleaks /usr/local/bin/gitleaks
COPY ./bin/cosign /usr/local/bin/cosign
COPY ./bin/crane /usr/local/bin/crane
COPY ./bin/release-cli /usr/local/bin/release-cli
COPY ./bin/gatecheck /usr/local/bin/gatecheck
COPY ./bin/s3upload /usr/local/bin/s3upload
COPY ./bin/just /usr/local/bin/just
COPY ./bin/oras /usr/local/bin/oras

RUN cp /usr/bin/freshclam /usr/local/bin/freshclam
RUN cp /usr/bin/clamscan /usr/local/bin/clamscan

USER omnibus

# Final image if building locally and build dependencies are needed
FROM final-base

COPY --from=build-just /usr/local/cargo/bin/just /usr/local/bin/just

COPY --from=build /usr/local/bin/grype /usr/local/bin/grype
COPY --from=build /usr/local/bin/syft /usr/local/bin/syft
COPY --from=build /usr/local/bin/gitleaks /usr/local/bin/gitleaks
COPY --from=build /usr/local/bin/cosign /usr/local/bin/cosign
COPY --from=build /usr/local/bin/crane /usr/local/bin/crane
COPY --from=build /usr/local/bin/release-cli /usr/local/bin/release-cli
COPY --from=build /usr/local/bin/gatecheck /usr/local/bin/gatecheck
COPY --from=build /usr/local/bin/s3upload /usr/local/bin/s3upload
COPY --from=build /usr/local/bin/oras /usr/local/bin/oras

RUN cp /usr/bin/freshclam /usr/local/bin/freshclam
RUN cp /usr/bin/clamscan /usr/local/bin/clamscan
