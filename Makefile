ROOT_DIR := $(shell pwd)

help:
	@echo "可用命令:"
	@echo "  make gen   - 根据cmd/gen/clients.json同步生成代码和配置"
	@echo "  make all - 编译主程序"

all: submods build

submods:
	$(MAKE) -C mail
	$(MAKE) -C scripts install

build:
	go build -o noticat .

gen:
	@echo "🚀 正在从母本生成代码与配置..."
	go run cmd/gen/main.go -root $(ROOT_DIR)
	@echo "✅ 同步完成！"

clean:
	$(MAKE) -C mail clean
	$(MAKE) -C scripts clean
	rm -f noticat

.PHONY: all submods build clean gen
