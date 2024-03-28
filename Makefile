all:
	go build ./cmd/mmdb
	ln -sf mmdb lookup
