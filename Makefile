GO = go
EXT=

PLAKAR	= plakar
VERSION	= v1.1.0

all: build

build:
	${GO} build -v -o github-importer${EXT} ./plugin/importer

package: build
	rm -f github_${VERSION}_*.ptar
	${PLAKAR} pkg create ./manifest.yaml ${VERSION}

uninstall:
	-${PLAKAR} pkg rm github

install: package
	${PLAKAR} pkg add ./github_${VERSION}_*.ptar

reinstall: uninstall install

test:
	${GO} test -v ./...

check: test

clean:
	rm -f github-importer${EXT} github_${version}*.ptar

.PHONY: all build package uninstall install reinstall test check clean
