package errors

import (
	"errors"
	"fmt"
	"sync"
)

// 默认未知错误码对象，用于解析失败或非 *withCode 类型的 error
var unknownCoder defaultCoder = defaultCoder{
	C:   1,                                           // 错误码编号
	Ext: "An internal server error occurred",         // 用户可见信息
	Ref: "https://github.com/tools/errors/README.md", // 文档参考链接
}

// Coder 是所有错误码必须实现的接口
type Coder interface {
	// String 返回用户可见的错误信息
	String() string

	// Reference 返回错误相关的文档地址，用于指导用户解决问题
	Reference() string

	// Code 返回错误码编号
	Code() int
}

// defaultCoder 是一个实现了 Coder 接口的基础错误码结构体
type defaultCoder struct {
	// C 表示错误码编号，例如：1001
	C int

	// Ext 是用户可见的错误描述，适合对外暴露
	Ext string

	// Ref 是错误参考文档地址，用于帮助用户排查问题
	Ref string
}

// Code 返回当前错误码的编号
func (coder defaultCoder) Code() int {
	return coder.C
}

// String 实现了 fmt.Stringer 接口，返回用户可见的错误信息
func (coder defaultCoder) String() string {
	return coder.Ext
}

// Reference 返回错误对应的文档链接
func (coder defaultCoder) Reference() string {
	return coder.Ref
}

// codes 存储注册的所有错误码，便于运行时查询
var codes = map[int]Coder{}

// codeMux 用于并发安全地操作 codes 映射表 TODO 分布式系统需要将这个转换成 etcd or redis setnx
var codeMux = &sync.Mutex{}

// Register 注册一个新的错误码。
// 如果已存在相同 Code，则会被覆盖。
func Register(coder Coder) {
	if coder.Code() == 0 {
		panic("code `0` is reserved by `github.com/tools/errors` as unknownCode error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	codes[coder.Code()] = coder
}

// MustRegister 注册一个新的错误码。
// 如果该 Code 已存在，则会触发 panic。
func MustRegister(coder Coder) {
	if coder.Code() == 0 {
		panic("code '0' is reserved by 'github.com/tools/errors' as ErrUnknown error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// ParseCoder 将任意 error 转换为 Coder 接口
// 如果 err 是 nil，返回 nil
// 如果 err 是 *withCode 类型，尝试从全局注册表中查找对应的 Coder
// 否则返回默认的 unknownCoder
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}

	var v *withCode
	if errors.As(err, &v) {
		if coder, ok := codes[v.code]; ok {
			return coder
		}
	}

	return unknownCoder
}

// IsCode 检查错误链中是否包含指定的错误码
// 支持嵌套错误（error chain）的查找
func IsCode(err error, code int) bool {
	var v *withCode
	if errors.As(err, &v) {
		if v.code == code {
			return true
		}

		if v.cause != nil {
			return IsCode(v.cause, code)
		}

		return false
	}

	return false
}

// 初始化时自动注册默认错误码（unknownCoder）
func init() {
	codes[unknownCoder.Code()] = unknownCoder
}
