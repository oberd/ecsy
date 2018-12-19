FROM alpine:3.8

RUN apk add --no-cache \
    ca-certificates=20171114-r3

COPY dist/ecsy-*-linux /usr/local/bin/ecsy

ENTRYPOINT ["/usr/local/bin/ecsy"]
