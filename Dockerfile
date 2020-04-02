FROM golang:alpine as builder
COPY . /go/src/github.com/trecnoc/nexus-resource
ENV CGO_ENABLED 0
WORKDIR /go/src/github.com/trecnoc/nexus-resource
RUN go build -o /assets/in ./cmd/in
RUN go build -o /assets/out ./cmd/out
RUN go build -o /assets/check ./cmd/check
RUN go generate ./...
RUN set -e; for pkg in $(go list ./...); do \
		go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM alpine:latest AS resource
RUN apk add --no-cache bash tzdata ca-certificates unzip zip gzip tar
COPY --from=builder assets/ /opt/resource/
RUN chmod +x /opt/resource/*

FROM resource AS tests
ARG NEXUS_TESTING_URL
ARG NEXUS_TESTING_USERNAME
ARG NEXUS_TESTING_PASSWORD
ARG NEXUS_TESTING_REPOSITORY
COPY --from=builder /tests /go-tests
WORKDIR /go-tests
RUN set -e; for test in /go-tests/*.test; do \
		$test; \
	done

FROM resource
