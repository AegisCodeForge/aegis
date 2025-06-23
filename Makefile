clean:
	if [ -f "static.zip" ]; then rm ./static.zip; fi
	if [ -f "aegis" ]; then rm ./aegis; fi

all:
	go run ./devtools/embed-static.go ./static templates
	go generate
	go build

