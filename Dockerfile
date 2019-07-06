FROM golang:1.10-alpine

RUN apk add --no-cache git libgit2-dev alpine-sdk
RUN go get github.com/Masterminds/glide

WORKDIR /go/src/github.com/bondar-pavel/pitank-server

COPY ./glide.* ./
RUN glide install --skip-test -v

COPY ./*.go ./

RUN CGO_ENABLED=1 GOOS=linux go install -a -ldflags '-extldflags "-static"' .

FROM alpine:latest

# needed only if we do https request to external resources
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /root

COPY ./static ./static
COPY ./templates ./templates

COPY --from=0 /go/bin/pitank-server /root/pitank-server
CMD ["/root/pitank-server", "--port", "80"]