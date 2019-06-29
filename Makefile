PREFIX = pavelgonchukov/
NAME = k8s-graph-controller

build:
	env CGO_ENABLED=0 go build -o ./bin/$(NAME)
clean:
	rm ./bin/$(NAME)

dep:
	go mod download

docker-build:
	docker build -t $(PREFIX)$(NAME) .
