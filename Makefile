#
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
BUILD_METADATA ?= "1~development~$(shell git rev-parse --short HEAD)"
CHART_METADATA_IMAGE ?= artifactory.algol60.net/csm-docker/stable/chart-metadata
HELM_DOCS_IMAGE ?= artifactory.algol60.net/docker.io/jnorwood/helm-docs:v1.5.0
HELM_IMAGE ?= artifactory.algol60.net/docker.io/alpine/helm:3.7.1
NAME ?= tpm-provisioner
RPM_BUILD_DIR ?= $(PWD)/dist/rpmbuild
RPM_NAME ?= tpm-provisioner-client
RPM_VERSION ?= $(shell cat .version)
RPM_SOURCE_NAME ?= ${RPM_NAME}-${RPM_VERSION}
RPM_SOURCE_PATH := ${RPM_BUILD_DIR}/SOURCES/${RPM_SOURCE_NAME}.tar.bz2
SPEC_FILE ?= ${SPEC_NAME}.spec
SPEC_NAME ?= tpm-provisioner
YQ_IMAGE ?= artifactory.algol60.net/docker.io/mikefarah/yq:4
export VERSION ?= $(shell cat .version)-local
export WORKSPACE = $(shell pwd)
export DOCKER_IMAGE ?= ${NAME}:${VERSION}

ifeq ($(GOARCH),)
	ifeq "$(ARCH)" "aarch64"
		export GOARCH=arm64
	else ifeq "$(ARCH)" "x86_64"
		export GOARCH=amd64
	endif
endif

DIR := ${CURDIR}
build_dir := $(DIR)/.build/$(GOARCH)
go_version = $(shell grep '^go' go.mod | awk '{print $$2}')
go_dir := $(build_dir)/go/$(go_version)
go_bin_dir = $(go_dir)/bin
go_path := PATH="$(go_bin_dir):$(PATH)"
go_url = https://storage.googleapis.com/golang/go$(go_version).linux-$(GOARCH).tar.gz



all: test dockerimage chart
chart: chart-lint dep-up chart-package
rpm: rpm_prepare rpm_package_source rpm_build_source rpm_build

dockerimage:
		docker build --no-cache --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .

build:
	go build -o ./bin/tpm-provisioner-client ./cmd/client
	go build -o ./bin/tpm-provisioner-server ./cmd/server
	go build -o ./bin/tpm-getEK ./cmd/getEK
	go build -o ./bin/tpm-blob-clear ./cmd/blob-clear
	go build -o ./bin/tpm-blob-store ./cmd/blob-store
	go build -o ./bin/tpm-blob-retrieve ./cmd/blob-retrieve

build-jenkins:
	$(go_bin_dir)/go build -o ./bin/tpm-provisioner-client ./cmd/client
	$(go_bin_dir)/go build -o ./bin/tpm-provisioner-server ./cmd/server
	$(go_bin_dir)/go build -o ./bin/tpm-getEK ./cmd/getEK
	$(go_bin_dir)/go build -o ./bin/tpm-blob-clear ./cmd/blob-clear
	$(go_bin_dir)/go build -o ./bin/tpm-blob-store ./cmd/blob-store
	$(go_bin_dir)/go build -o ./bin/tpm-blob-retrieve ./cmd/blob-retrieve

build-linux:
	GOOS=linux go build -o ./bin/tpm-provisioner-client ./cmd/client
	GOOS=linux go build -o ./bin/tpm-provisioner-server ./cmd/server
	GOOS=linux go build -o ./bin/tpm-getEK ./cmd/getEK
	GOOS=linux go build -o ./bin/tpm-blob-clear ./cmd/blob-clear
	GOOS=linux go build -o ./bin/tpm-blob-store ./cmd/blob-store
	GOOS=linux go build -o ./bin/tpm-blob-retrieve ./cmd/blob-retrieve

test:
	go test -v ./...

helm:
	docker run --rm \
		--user $(shell id -u):$(shell id -g) \
		--mount type=bind,src="$(shell pwd)",dst=/src \
		-w /src \
		-e HELM_CACHE_HOME=/src/.helm/cache \
		-e HELM_CONFIG_HOME=/src/.helm/config \
		-e HELM_DATA_HOME=/src/.helm/data \
		$(HELM_IMAGE) \
		$(CMD)

chart-lint:
	CMD="lint charts/tpm-provisioner"              $(MAKE) helm

dep-up:
	CMD="dep up charts/tpm-provisioner"              $(MAKE) helm

chart-package:
ifdef CHART_VERSIONS
	CMD="package charts/tpm-provisioner              --version $(word 1, $(CHART_VERSIONS)) -d packages" $(MAKE) helm
else
	CMD="package charts/* -d packages" $(MAKE) helm
endif

extracted-images:
	CMD="template release $(CHART) --dry-run --replace --dependency-update" $(MAKE) -s helm \
	| docker run --rm -i $(YQ_IMAGE) e -N '.. | .image? | select(.)' -

annotated-images:
	CMD="show chart $(CHART)" $(MAKE) -s helm \
	| docker run --rm -i $(YQ_IMAGE) e -N '.annotations."artifacthub.io/images"' - \
	| docker run --rm -i $(YQ_IMAGE) e -N '.. | .image? | select(.)' -

images:
	{ CHART=charts/tpm-provisioner              $(MAKE) -s extracted-images annotated-images; \
	} | sort -u

snyk:
	$(MAKE) -s images | xargs --verbose -n 1 snyk container test

gen-docs:
	docker run --rm \
		--user $(shell id -u):$(shell id -g) \
		--mount type=bind,src="$(shell pwd)",dst=/src \
		-w /src \
		$(HELM_DOCS_IMAGE) \
		helm-docs --chart-search-root=charts

clean:
	$(RM) -r .helm packages charts/tpm-provisioner/charts

rpm_prepare:
	rm -rf $(RPM_BUILD_DIR)
	mkdir -p $(RPM_BUILD_DIR)/SPECS $(RPM_BUILD_DIR)/SOURCES
	cp $(SPEC_FILE) $(RPM_BUILD_DIR)/SPECS/

rpm_package_source:
	tar --transform 'flags=r;s,^,/$(RPM_SOURCE_NAME)/,' --exclude .git --exclude dist -cvjf $(RPM_SOURCE_PATH) .

rpm_build_source:
	BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ts $(RPM_SOURCE_PATH) --define "_topdir $(RPM_BUILD_DIR)"

rpm_build:
	BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ba $(SPEC_FILE) --define "_topdir $(RPM_BUILD_DIR)"

install_go:
	rm -rf $(dir $(go_dir))
	mkdir -p $(go_dir)
	curl -sSfL $(go_url) | tar xz -C $(go_dir) --strip-components=1
