# Build the manager binary
FROM golang:1.20 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY controllers/ controllers/
COPY internal/controller internal/controller

# Build
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -a -o manager cmd/main.go

FROM registry.access.redhat.com/ubi8/ubi-micro:8.7

ENV HOME=/opt/helm \
    USER_NAME=helm \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

# Copy necessary files with the right permissions
COPY --chown=${USER_UID}:0 watches.yaml ${HOME}/watches.yaml
COPY --chown=${USER_UID}:0 helm-charts  ${HOME}/helm-charts

# Copy manager binary
COPY --from=builder /workspace/manager .

USER ${USER_UID}

WORKDIR ${HOME}

ENTRYPOINT ["/manager"]
