BINARY_NAME=mario

build:
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME} main.go
# GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux main.go
# GOARCH=amd64 GOOS=windows go build -o ${BINARY_NAME}-windows main.go

run: build
	./${BINARY_NAME}

clean:
	go clean