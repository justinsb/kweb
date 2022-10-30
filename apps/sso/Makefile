GCP_PROJECT_ID ?= $(shell gcloud config get project)
IMAGE_REPO ?= gcr.io/$(GCP_PROJECT_ID)
IMAGE_TAG ?= latest

.PHONY: push
push:
	cd ../..; docker buildx build --push --tag "$(IMAGE_REPO)/kweb-sso:$(IMAGE_TAG)" -f apps/sso/Dockerfile .

.PHONY: apply
apply: push
	cat ./k8s/manifest.yaml | sed -e s~image-repo~${IMAGE_REPO}~g | kubectl apply -f -

.PHONY: bounce
bounce:
	kubectl delete pod -n kweb-sso-system -l app=kweb-sso
