FROM alpine:latest as build
MAINTAINER Anthony O'Brien

ARG VERSION
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

ENTRYPOINT [ "/usr/bin/ghtrim" ]

WORKDIR /go/src/github.com/asobrien/ghtrim

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		go \
		git \
		gcc \
		libc-dev \
		libgcc

COPY . .

RUN set -x \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -a -tags "netgo cgo static_build" \
		-ldflags="-X main.VERSION=${VERSION} -w -extldflags -static" \
		-o /usr/bin/ghtrim . \
	&& echo "Build complete."


FROM scratch
MAINTAINER Anthony O'Brien

ENTRYPOINT [ "/ghtrim" ]

ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca.cert
COPY --from=build /usr/bin/ghtrim /ghtrim
