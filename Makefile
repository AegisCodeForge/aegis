clean:
	if [ -f "gitus" ]; then rm ./gitus; fi

gitus-web-server:
	go run ./devtools/generate-template.go templates
	go run ./devtools/generate-footer-template.go
	go run ./devtools/embed-static.go ./static templates
	go build ./cmd/gitus

all:
	make gitus-web-server

