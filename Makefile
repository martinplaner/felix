HASGOVVV := $(shell command -v govvv 2> /dev/null)
GO = go
ifdef HASGOVVV
    GO = govvv
endif

default:
	$(GO) build -v

test:
	go test -v -cover ./...

docker:
	docker build . -t martinplaner/felix

docker-run:
	docker run -it --rm -v $(shell pwd)/config.yml:/config.yml -v /data -p 6554:6554 martinplaner/felix
