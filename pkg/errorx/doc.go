// Package errorx 提供了一个带状态码的错误处理包
//
// 特性：
//   - 支持错误码定义和注册
//   - 自动生成堆栈跟踪信息
//   - 支持错误包装和链式错误处理
//   - 支持错误消息中的占位符替换
//   - 支持额外信息（Extra）附加
//   - 支持稳定性标记（IsAffectStability）
//   - 兼容标准库 errors 包的 Unwrap、Is、As 方法
//
// 基本使用：
//
//	// 1. 注册错误码（通常在初始化时调用）
//	import "github.com/ZampoRen/go-server-comon/pkg/errorx/code"
//
//	code.Register(1001, "用户不存在")
//	code.Register(1002, "用户名为空", code.WithAffectStability(false))
//
//	// 2. 创建错误
//	import "github.com/ZampoRen/go-server-comon/pkg/errorx"
//
//	err := errorx.New(1001)
//	err := errorx.New(1002, errorx.KV("username", "test"))
//
//	// 3. 包装错误
//	err := errorx.WrapByCode(originalErr, 1001)
//	err := errorx.Wrapf(originalErr, "处理用户数据失败: %s", "详细信息")
//
//	// 4. 获取错误信息
//	if se, ok := err.(errorx.StatusError); ok {
//		code := se.Code()           // 获取错误码
//		msg := se.Msg()             // 获取错误消息
//		extra := se.Extra()         // 获取额外信息
//		affect := se.IsAffectStability() // 是否影响稳定性
//	}
//
//	// 5. 获取不带堆栈的错误消息
//	msg := errorx.ErrorWithoutStack(err)
//
// 错误码注册：
//
// 使用 code 包注册错误码定义：
//
//	import "github.com/ZampoRen/go-server-comon/pkg/errorx/code"
//
//	// 注册基本错误码
//	code.Register(1001, "用户不存在")
//
//	// 注册错误码并设置不影响稳定性
//	code.Register(1002, "参数验证失败", code.WithAffectStability(false))
//
//	// 设置默认错误码（用于未定义的错误码）
//	code.SetDefaultErrorCode(9999)
//
// 错误创建：
//
//	// 通过错误码创建错误（会自动生成堆栈跟踪）
//	err := errorx.New(1001)
//
//	// 使用占位符替换（错误消息中可以使用 {key} 作为占位符）
//	code.Register(1001, "用户 {username} 不存在")
//	err := errorx.New(1001, errorx.KV("username", "alice"))
//
//	// 使用格式化占位符
//	err := errorx.New(1001, errorx.KVf("username", "用户: %s", "alice"))
//
//	// 添加额外信息
//	err := errorx.New(1001, errorx.Extra("request_id", "12345"))
//
// 错误包装：
//
//	// 用错误码包装现有错误
//	err := errorx.WrapByCode(originalErr, 1001)
//
//	// 用格式化消息包装错误
//	err := errorx.Wrapf(originalErr, "处理用户数据失败: %s", "详细信息")
//
//	// 组合使用
//	err := errorx.WrapByCode(originalErr, 1001,
//		errorx.KV("username", "alice"),
//		errorx.Extra("request_id", "12345"),
//	)
//
// 错误信息获取：
//
//	// 检查是否为 StatusError
//	if se, ok := err.(errorx.StatusError); ok {
//		code := se.Code()                    // int32: 错误码
//		msg := se.Msg()                      // string: 错误消息
//		extra := se.Extra()                  // map[string]string: 额外信息
//		affect := se.IsAffectStability()     // bool: 是否影响稳定性
//	}
//
//	// 使用标准库的 errors 包方法
//	if errors.Is(err, targetErr) {
//		// 错误匹配
//	}
//
//	var se errorx.StatusError
//	if errors.As(err, &se) {
//		// 错误类型转换
//	}
//
//	// 获取不带堆栈的错误消息（用于日志记录）
//	msg := errorx.ErrorWithoutStack(err)
//
// 稳定性标记：
//
// 错误可以标记为是否影响系统稳定性：
//
//   - true: 会影响系统稳定性，会在接口错误率中体现
//
//   - false: 不会影响稳定性，例如参数验证错误
//
//     code.Register(1001, "用户不存在") // 默认影响稳定性
//     code.Register(1002, "参数验证失败", code.WithAffectStability(false))
//
// 占位符替换：
//
// 错误消息中可以使用 {key} 作为占位符，通过 KV 或 KVf 选项进行替换：
//
//	code.Register(1001, "用户 {username} 不存在，ID: {user_id}")
//
//	err := errorx.New(1001,
//		errorx.KV("username", "alice"),
//		errorx.KV("user_id", "123"),
//	)
//
//	// 输出: "用户 alice 不存在，ID: 123"
//
// 额外信息：
//
// 使用 Extra 选项可以添加额外的键值对信息，这些信息不会替换消息中的占位符：
//
//	err := errorx.New(1001,
//		errorx.KV("username", "alice"),           // 替换占位符
//		errorx.Extra("request_id", "12345"),      // 添加额外信息
//		errorx.Extra("trace_id", "abc-123"),
//	)
//
//	if se, ok := err.(errorx.StatusError); ok {
//		extra := se.Extra()
//		// extra["request_id"] = "12345"
//		// extra["trace_id"] = "abc-123"
//	}
//
// 堆栈跟踪：
//
// 所有通过 New、WrapByCode、Wrapf 创建的错误都会自动包含堆栈跟踪信息。
// 使用 ErrorWithoutStack 可以获取不带堆栈的错误消息，适合用于日志记录。
//
//	err := errorx.New(1001)
//	fullMsg := err.Error()              // 包含堆栈信息
//	simpleMsg := errorx.ErrorWithoutStack(err) // 不包含堆栈信息
//
// 示例：
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/ZampoRen/go-server-comon/pkg/errorx"
//		"github.com/ZampoRen/go-server-comon/pkg/errorx/code"
//	)
//
//	func init() {
//		// 注册错误码
//		code.Register(1001, "用户不存在")
//		code.Register(1002, "用户 {username} 已被禁用")
//		code.Register(1003, "参数验证失败", code.WithAffectStability(false))
//	}
//
//	func main() {
//		// 创建错误
//		err := errorx.New(1001)
//		fmt.Println(err)
//
//		// 使用占位符
//		err = errorx.New(1002, errorx.KV("username", "alice"))
//		fmt.Println(errorx.ErrorWithoutStack(err))
//
//		// 包装错误
//		originalErr := fmt.Errorf("数据库连接失败")
//		err = errorx.WrapByCode(originalErr, 1001)
//		fmt.Println(err)
//
//		// 获取错误信息
//		if se, ok := err.(errorx.StatusError); ok {
//			fmt.Printf("错误码: %d\n", se.Code())
//			fmt.Printf("错误消息: %s\n", se.Msg())
//			fmt.Printf("影响稳定性: %v\n", se.IsAffectStability())
//		}
//	}
package errorx
