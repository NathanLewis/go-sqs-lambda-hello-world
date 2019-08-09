
build:
		GOOS=linux GOARCH=amd64 go build -o main cmd/lambdaapp/lambdapp.go
		zip main.zip main