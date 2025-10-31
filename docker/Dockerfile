FROM alpine:3.22

ARG EXECUTABLE

RUN apk add --no-cache ca-certificates \
                       e2fsprogs \
                       findmnt \
                       xfsprogs \
                       xfsprogs-extra \
                       blkid \
                       e2fsprogs-extra

ADD ${EXECUTABLE} /bin/

ENTRYPOINT ["/bin/${EXECUTABLE}"]