IMAGE_NAME=pitank_server
VERSION=0.1.0

image:
	docker build -t $(IMAGE_NAME):$(VERSION) .

push:
	docker push $(IMAGE_NAME):$(VERSION)

build:
	go build -o pitank_server

fmt:
	gofmt -w *.go