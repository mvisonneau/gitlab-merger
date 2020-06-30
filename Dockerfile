##
# BUILD CONTAINER
##

FROM goreleaser/goreleaser:v0.138.0 as builder

WORKDIR /build

COPY . .
RUN \
  apk add --no-cache make ca-certificates ;\
  make build-linux-amd64

##
# RELEASE CONTAINER
##

FROM busybox:1.32.0-glibc

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dist/gitlab-merger_linux_amd64/gitlab-merger /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/gitlab-merger"]
CMD [""]
