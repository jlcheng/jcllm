check:
	staticcheck ./...
	goimports -d .
	gofmt -d -s .

fmt:
	goimports -w .
	gofmt -w .

test:
	go test ./...