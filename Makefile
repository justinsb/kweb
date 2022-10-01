.PHONY: generate
generate: go-mod-tidy go-generate protoc-generate format

.PHONY: format
format: go-fmt protoc-fmt

.PHONY: go-mod-tidy
go-mod-tidy:
	go mod tidy

.PHONY: go-generate
go-generate:
	go generate ./...

.PHONY: go-fmt
go-fmt:
	go fmt ./...

.PHONY: protoc-generate
protoc-generate:
	cd dev/build/protobuf; docker buildx build --tag protobuf --load .
	docker run -v `pwd`:/workspace protobuf protoc -I. --go_out=. --go_opt=paths=source_relative \
		./components/kube/kube.proto
	docker run -v `pwd`:/workspace protobuf protoc -I./components/kube -I. --descriptor_set_out=components/all.pb \
		--go_out=. --go_opt=paths=source_relative \
		./components/login/pb/state.proto  \
		./components/github/pb/types.proto \
		./components/users/pb/user.proto

.PHONY: protoc-fmt
protoc-fmt:
	find -name "*.proto" | xargs clang-format -i

.PHONY: apply
apply:
	 kubectl apply --server-side -f components/users/kbapi/config/
	 kubectl apply --server-side -f components/github/kbapi/config/
