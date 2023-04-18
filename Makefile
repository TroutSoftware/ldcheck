GO := go build -trimpath

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 $(GO) -o ldcheck_linux_amd64 ./ldcheck
	GOOS=darwin GOARCH=amd64 $(GO) -o ldcheck_darwin_amd64 ./ldcheck
	GOOS=darwin GOARCH=arm64 $(GO) -o ldcheck_darwin_arm64 ./ldcheck
	GOOS=windows GOARCH=amd64 $(GO) -o ldcheck_windows_amd64 ./ldcheck
