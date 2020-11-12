.PHONY: clean

bin/evictor:
	go build -o bin/evictor ./cmd/evictor

clean:
	rm -rf bin
