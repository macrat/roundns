FROM golang:latest AS builder

ARG arch
ARG armv
WORKDIR /usr/src/roundns

COPY . .

RUN go get -d && \
    GOARCH=${arch} \
    GOARM=${armv} \
    go build \
    -tags netgo \
    -installsuffix netgo \
    --ldflags '-extldflags -static' \
    -o /roundns \
    .


FROM scratch

WORKDIR /
COPY --from=builder /roundns /roundns

EXPOSE 53/udp
EXPOSE 9553/tcp
ENTRYPOINT ["/roundns"]
