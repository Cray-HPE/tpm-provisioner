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
########### Build ##########
FROM artifactory.algol60.net/docker.io/library/golang:alpine AS build

RUN apk add --no-cache git build-base openssl-dev

RUN mkdir -p /build
COPY cmd pkg third_party go.mod conf  tests  Dockerfile go.sum  Makefile /build/
RUN cd /build && go build ./...
RUN cd /build && go test -v ./...
RUN	cd /build && go build -o /usr/local/bin/tpm-provisioner-server ./cmd/server

########## Runtime ##########
FROM gcr.io/distroless/static:nonroot AS runtime
USER nonroot

COPY --from=build /usr/local/bin/tpm-provisioner-server /usr/local/bin/tpm-provisioner-server
COPY ./.version /version
EXPOSE 8080/tcp
ENTRYPOINT ["/usr/local/bin/tpm-provisioner-server"]
