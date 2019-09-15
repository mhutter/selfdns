IMAGE ?= mhutter/selfdns:amd64
DOCKERFILE ?= Dockerfile

test:
	go test -v -race -cover ./...

image:
	docker build -t $(IMAGE) -f $(DOCKERFILE) .

run:
	docker run \
		--env-file=.env \
		-v "$(PWD)/gcloud-sa.json:/gcloud-sa.json:ro" \
		--rm $(IMAGE)
