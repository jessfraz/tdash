FROM golang:alpine as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . ${GOPATH}/src/github.com/jessfraz/tdash

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd ${GOPATH}/src/github.com/jessfraz/tdash \
	&& make static \
	&& mv tdash /usr/bin/tdash \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM scratch

COPY --from=builder /usr/bin/tdash /usr/bin/tdash
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "tdash" ]
CMD [ "--help" ]
