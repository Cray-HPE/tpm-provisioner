#!/bin/bash

grep image Chart.yaml | awk '{print $2}' | while read -r image; do
	docker pull "$image"
	kind load docker-image "$image"
done
