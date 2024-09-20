
lint::
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0 -v run ./...

test:
	go test -coverpkg=./... -race -coverprofile=cover.out.tmp -covermode atomic -v ./...
	cat cover.out.tmp | grep -v "cmd/b2bgw" | grep -v "_mock.go" > coverage.txt
	go tool cover -func coverage.txt
	rm cover.out.tmp coverage.txt

build-image:
	docker build -f Dockerfile . \
		  --platform linux/amd64 \
          --tag yas3-front:local