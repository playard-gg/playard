.PHONY: dev dev-server dev-web build build-server build-web run clean

dev-server:
	cd server && go run ./cmd/server

dev-web:
	cd web && npm run dev

dev:
	@trap 'kill 0' EXIT; \
	$(MAKE) dev-server & \
	$(MAKE) dev-web & \
	wait

build-server:
	cd server && go build -o bin/server ./cmd/server

build-web:
	cd web && npm run build

build: build-server build-web

run: build
	@trap 'kill 0' EXIT; \
	./server/bin/server & \
	cd web && npm run preview & \
	wait

clean:
	rm -rf server/bin web/dist
