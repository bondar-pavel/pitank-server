FROM golang:1.16-alpine

RUN apk add --no-cache git libgit2-dev alpine-sdk

WORKDIR /go/src/github.com/bondar-pavel/pitank-server

COPY ./go.* ./
RUN go mod download

COPY ./pkg ./pkg
COPY ./cmd ./cmd

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./bin/pitank-server ./cmd/main.go

FROM alpine:latest

# needed only if we do https request to external resources
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /root

COPY ./static ./static
COPY ./templates ./templates

COPY --from=0 /go/src/github.com/bondar-pavel/pitank-server/bin/pitank-server /root/pitank-server
CMD ["/root/pitank-server", "--port", "80"]