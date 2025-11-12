#!/bin/sh

# 初始化项目（仅在首次使用时运行）
if [ "$1" = "init" ]; then
    hz new -mod github.com/ZampoRen/go-server-comon \
        --handler_dir api/handler \
        --model_dir api/model
    hz update -idl api/proto/user/user.proto -I api/proto
fi

# 更新代码（.hz 文件中已配置目录，无需重复指定）
hz update -idl api/proto/user/user.proto -I api/proto

