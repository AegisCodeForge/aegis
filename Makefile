clean:
	if [ -f "static.zip" ]; then rm ./static.zip; fi
	if [ -f "gitus" ]; then rm ./gitus; fi

all:
	go run ./devtools/pack-static.go ./static
	go generate
	go build

