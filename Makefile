# Get the latest commit branch, hash, and date
TAG=$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
BRANCH=$(if $(TAG),$(TAG),$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
HASH=$(shell git rev-parse --short=7 HEAD 2>/dev/null)
TIMESTAMP=$(shell git log -1 --format=%ct HEAD 2>/dev/null | xargs -I{} date -u -r {} +%Y%m%dT%H%M%S)
GIT_REV=$(shell printf "%s-%s-%s" "$(BRANCH)" "$(HASH)" "$(TIMESTAMP)")
REV=$(if $(filter --,$(GIT_REV)),latest,$(GIT_REV)) # fallback to latest if not in git repo


all: test build

build:
	cd cmd/spot && go build -ldflags "-X main.revision=$(REV) -s -w" -o ../../.bin/spot.$(BRANCH)
	cd cmd/secrets && go build -ldflags "-X main.revision=$(REV) -s -w" -o ../../.bin/spot-secrets.$(BRANCH)
	cp .bin/spot.$(BRANCH) .bin/spot
	cp .bin/spot-secrets.$(BRANCH) .bin/spot-secrets

release:
	@echo release to .bin
	goreleaser --snapshot --skip-publish --clean
	ls -l .bin

test:
	go clean -testcache
	go test -race -coverprofile=coverage.out ./...
	grep -v "_mock.go" coverage.out | grep -v mocks > coverage_no_mocks.out
	go tool cover -func=coverage_no_mocks.out
	rm coverage.out coverage_no_mocks.out

version:
	@echo "branch: $(BRANCH), hash: $(HASH), timestamp: $(TIMESTAMP)"
	@echo "revision: $(REV)"

site:
	@rm -f  site/public/*
	@docker rm -f spot-site
	docker build -f Dockerfile.site --progress=plain -t spot.site .
	docker run -d --name=spot-site spot.site
	sleep 3
	docker cp "spot-site":/srv/site/ site/public
	docker rm -f spot-site

.PHONY: build release test site