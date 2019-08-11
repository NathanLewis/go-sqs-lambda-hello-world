clean:
		rm -rf ./venv
build: clean
		GOOS=linux GOARCH=amd64 go build -o main cmd/lambdaapp/lambdapp.go
		zip main.zip main

package:
		python3 --version
		python3 -m venv venv
		./venv/bin/pip install --upgrade pip
		./venv/bin/pip install -r scripts/src/lamda/requirements.txt

