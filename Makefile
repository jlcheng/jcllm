check:
	staticcheck ./...
	goimports -d .

fmt:
	goimports -w .

test:
	go test ./...