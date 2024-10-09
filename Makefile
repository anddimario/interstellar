.PHONY: dev
dev:
	air -c .air.toml

# https://hub.docker.com/r/hashicorp/http-echo/
# example: make be port=8081 msg="Hello, World!"
.PHONY: be
be:
	docker run -p $(port):$(port) hashicorp/http-echo -listen=:$(port) -text="$(msg)"

# .PHONY: test
# test:
# 	go test ./...

# .PHONY: coverage
# coverage:
# 	go test -v -cover ./...
