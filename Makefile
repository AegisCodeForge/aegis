clean:
	if [ -f "aegis" ]; then rm ./aegis; fi

all:
	go run ./devtools/generate-footer-template.go
	go run ./devtools/embed-static.go ./static templates
	go generate
	go build

