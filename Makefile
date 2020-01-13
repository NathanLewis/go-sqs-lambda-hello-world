clean:
		rm -rf ./venv
build: 
		GOOS=linux GOARCH=amd64 go build -o main *.go
		zip main.zip main

package:
		python3 --version
		python3 -m venv venv
		./venv/bin/pip install --upgrade pip
		./venv/bin/pip install -r scripts/python/src/requirements.txt

