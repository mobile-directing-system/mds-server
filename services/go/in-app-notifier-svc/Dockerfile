FROM golang:1.19 as builder
WORKDIR /go/src/github.com/mobile-directing-system/mds-server

# Copy shared library.
COPY /services/go/shared ./services/go/shared

# Download dependencies.
COPY /services/go/in-app-notifier-svc/go.mod ./services/go/in-app-notifier-svc/go.mod
COPY /services/go/in-app-notifier-svc/go.sum ./services/go/in-app-notifier-svc/go.sum

WORKDIR /go/src/github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc
RUN go mod download

# Perform the actual build.

COPY /services/go/in-app-notifier-svc ./

# `skaffold debug` sets SKAFFOLD_GO_GCFLAGS to disable compiler optimizations
ARG SKAFFOLD_GO_GCFLAGS
RUN go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -o /app ./main.go

FROM debian
# Define GOTRACEBACK to mark this container as using the Go language runtime
# for `skaffold debug` (https://skaffold.dev/docs/workflows/debug/).
ENV GOTRACEBACK=single
EXPOSE 8080
CMD ["/app"]
COPY --from=builder /app .
