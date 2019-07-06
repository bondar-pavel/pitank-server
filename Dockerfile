FROM golang:1.10-alpine

RUN go get github.com/Masterminds/glide

WORKDIR /go/src/github.com/bondar-pavel/pitank-server

COPY ./glide.* ./
RUN glide install --skip-test -v

COPY ./static ./static
COPY ./templates ./templates
COPY ./*.go ./

RUN CGO_ENABLED=1 GOOS=linux go install -a -ldflags '-extldflags "-static"' ./pitank_server

FROM alpine:latest

# needed only if we do https request to external resources
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=0 /go/bin/pitank_server ./
CMD ["/root/pitank_server", "--port", "80"]