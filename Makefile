.PHONY: all build clean run check cover lint docker help
BIN_FILE=onedss
all: check build
build:
	@go build -tags release -ldflags "-s -w" -o "${BIN_FILE}"
clean:
	@go clean
	rm --force "${BIN_FILE}"
test:
	@go test
check:
	@go fmt ./
	@go vet ./
cover:
	@go test -coverprofile "${BIN_FILE}"
	@go tool cover -html="${BIN_FILE}"
run:
	./"${BIN_FILE}"
lint:
	golangci-lint run --enable-all
help:
	@echo "make 格式化go代码 并编译生成二进制文件"
	@echo "make build 编译go代码生成二进制文件"
	@echo "make clean 清理中间目标文件"
	@echo "make test 执行测试case"
	@echo "make check 格式化go代码"
	@echo "make cover 检查测试覆盖率"
	@echo "make run 直接运行程序"
	@echo "make lint 执行代码检查"