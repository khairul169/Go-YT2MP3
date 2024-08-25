all: build-ui backend

build-ui:
	cd ui && npm run build

backend:
	CGO_ENABLED=0 go build -o main -tags="prod" main.go

clean:
	rm -f main && rm -rf ui/dist
