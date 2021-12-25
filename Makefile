IMAGE_NAME=pitank-server
REGISTRY=pbondar
VERSION=0.2.0

image:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(VERSION) .

image-amd64:
	docker build --platform linux/amd64 -t $(REGISTRY)/$(IMAGE_NAME):$(VERSION) .

push:
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)

build:
	go build -o ./bin/pitank_server cmd/main.go

fmt:
	gofmt -w cmd pkg