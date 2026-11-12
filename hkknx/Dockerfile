# syntax=docker/dockerfile:1
# Build a small, dependency-free path adapter for the Home Assistant sidebar.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS ingress-builder
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
WORKDIR /src
COPY ingress-proxy.go .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} GOARM=${TARGETVARIANT#v} \
    go build -trimpath -ldflags="-s -w" -o /hkknx-ingress-proxy ./ingress-proxy.go

# Keep HKKNX itself on the existing pinned upstream image.
FROM brutella/hkknx@sha256:93573884ddc56fc6d7d51406ec509c50d7861183fa0ebb11725a47804ff2bdee
COPY --from=ingress-builder /hkknx-ingress-proxy /usr/local/bin/hkknx-ingress-proxy
COPY run.sh /usr/local/bin/run.sh
RUN chmod +x /usr/local/bin/run.sh /usr/local/bin/hkknx-ingress-proxy
ENTRYPOINT ["/usr/local/bin/run.sh"]
