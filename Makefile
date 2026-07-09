BINARY := msn-converter
APP_ID := com.rrc4du.msnconverter

.PHONY: build run test tidy clean windows windows-mingw

## build: build native binary for the host OS
build:
	go build -o $(BINARY) .

## run: build and run locally
run:
	go run .

## test: run all tests
test:
	go test ./...

## tidy: sync go.mod/go.sum
tidy:
	go mod tidy

## windows: cross-build Windows .exe via fyne-cross (needs Docker running)
windows:
	go run github.com/fyne-io/fyne-cross@latest windows -arch=amd64 -app-id $(APP_ID)

## windows-mingw: cross-build Windows .exe on host via mingw-w64 (needs `brew install mingw-w64` + Icon.png)
windows-mingw:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
		go run fyne.io/tools/cmd/fyne@latest package -os windows -icon Icon.png

## clean: remove build artifacts
clean:
	rm -rf $(BINARY) $(BINARY).exe fyne-cross/
