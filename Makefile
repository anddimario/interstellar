.PHONY: dev
dev:
	air -c .air.toml

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -v -cover ./...

.PHONY: load
load:
	siege -c 5 -t 1m http://localhost:8080

.PHONY: vuln
vuln:
	govulncheck ./...

# example: make leader name=peer1 udp_addr=localhost:9001
.PHONY: leader
leader:
	go run main.go serve -e $(udp_addr)

# example: make peer name=peer2 udp_addr=localhost:9002 bc=localhost:9001 addr=localhost:8081
.PHONY: peer
peer:
	go run main.go serve -e $(udp_addr) -b $(bc) -a $(addr)
