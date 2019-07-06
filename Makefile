IMAGE_NAME=pitank-server
REGISTRY=pbondar
VERSION=0.1.0

image:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(VERSION) .

push:
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)

build:
	go build -o pitank_server

fmt:
	gofmt -w *.go