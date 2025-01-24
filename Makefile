check:
	staticcheck ./...
	goimports -d .
	gofmt -d -s .

fmt:
	goimports -w .

test:
	go test ./...