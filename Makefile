.PHONY: clean evictor

evictor:
	go build -o bin/evictor ./cmd/evictor

clean:
	rm -rf bin
