.PHONY: protoc
protoc:
	protoc  -I. --go_out=. --go_opt=paths=source_relative \
		./components/kube/kube.proto
	protoc -I./components/kube -I. --descriptor_set_out=components/all.pb \
		--go_out=. --go_opt=paths=source_relative \
		./components/login/pb/state.proto  \
		./components/users/pb/user.proto
