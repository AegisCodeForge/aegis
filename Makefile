clean:
	if [ -f "static.zip" ]; then rm ./static.zip; fi
	if [ -f "aegis" ]; then rm ./aegis; fi

all:
	go run ./devtools/pack-static.go ./static
	go generate
	go build

