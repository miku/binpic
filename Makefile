binpic: cmd/binpic/main.go
	go build -o $@ $<

.PHONY: clean
clean:
	rm -f binpic

