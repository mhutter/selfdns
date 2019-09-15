IMAGE ?= mhutter/selfdns
DOCKERFILE ?= Dockerfile

test:
	go test -v -race -cover ./...

images: image-amd image-arm
image-amd: Dockerfile
	docker build -t $(IMAGE):amd64 -f $< .
image-arm: Dockerfile.arm
	docker build -t $(IMAGE):arm -f $< .
manifest:
	docker manifest create \
		--amend \
		$(IMAGE) \
		$(IMAGE):amd64 \
		$(IMAGE):arm
	docker manifest annotate \
		--arch arm \
		$(IMAGE) \
		$(IMAGE):arm
	docker manifest push $(IMAGE)

run:
	docker run \
		--env-file=.env \
		-v "$(PWD)/gcloud-sa.json:/gcloud-sa.json:ro" \
		--rm $(IMAGE)

Dockerfile.arm: Dockerfile
	sed 's/amd64/arm/g' $< > $@

.PHONY: test image run
