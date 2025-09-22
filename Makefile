clean:
	if [ -f "aegis" ]; then rm ./aegis; fi

aegis-web-server:
	go run ./devtools/generate-template.go templates
	go run ./devtools/generate-footer-template.go
	go run ./devtools/embed-static.go ./static templates
	go build ./cmd/aegis

all:
	make aegis-web-server

