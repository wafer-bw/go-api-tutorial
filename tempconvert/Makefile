protoc:
	rm -f contract/*.pb.go
	protoc contract/contract.proto --go_out=contract
	mv contract/tempconvert/contract/* contract
	rm -rf contract/tempconvert
.PHONY: protoc

test:
	go test -v -coverprofile=cover.out ./...
.PHONY: test

run:
	go run main.go
.PHONY: run
