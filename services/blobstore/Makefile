GCP_PROJECT_ID ?= $(shell gcloud config get project)
IMAGE_REPO ?= gcr.io/$(GCP_PROJECT_ID)
IMAGE_TAG ?= latest

.PHONY: push
push:
	KO_DOCKER_REPO=${IMAGE_REPO}/kweb-blobstore-server ko build  --bare --tags ${IMAGE_TAG} ./cmd/blobstore-server/

.PHONY: apply
apply: push
	cat ./k8s/manifest.yaml | sed -e s~image-repo~${IMAGE_REPO}~g | kubectl apply -f -

.PHONY: bounce
bounce:
	kubectl delete pod -n kweb-blobstore -l app=kweb-blobstore-server
