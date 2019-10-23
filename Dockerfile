##
# BUILD CONTAINER
##

FROM goreleaser/goreleaser:v0.112.2 as builder

WORKDIR /build

COPY Makefile .
RUN \
apk add --no-cache make ;\
make setup

COPY . .
RUN \
make build

##
# RELEASE CONTAINER
##

FROM busybox:1.31-glibc

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dist/s5_linux_amd64/s5 /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/s5"]
CMD [""]