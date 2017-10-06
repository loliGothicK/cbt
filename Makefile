ARCH := 386 amd64
OS := linux darwin windows

preinstall: 
	go get golang.org/x/tools/cmd/goyacc
	go get github.com/mitchellh/gox
	go get github.com/jstemmer/go-junit-report
	go get github.com/haya14busa/goverage
	go get golang.org/x/tools/cmd/cover
	go get -u github.com/golang/lint/golint
	go get github.com/goreleaser/goreleaser

status:
	dep status

install:
	dep ensure

update:
	dep ensure -update

lint: 
	golint ./...

build:
	go generate ./...

test:
	go test ./...

package: build
	gox -os="$(OS)" -arch="$(ARCH)" -output "dist/{{.OS}}_{{.Arch}}/{{.Dir}}"