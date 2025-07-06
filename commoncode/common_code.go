package commoncode

// 业务错误码设计 100001 六位数字
// 10 服务号 | 00 模块号 | 01 错误码|
// 10 所有通用服务 共享错误码 如(解析错误,结构体绑定错误等) | 00 | 错误码 |
// 每一个服务创建时 需要在 配置中心查看 服务错误码是否使用

// common: basic 基础模块
const (
	// Success 成功
	Success int = iota + 100001
	// ErrorBind 绑定结构体错误
	ErrorBind
	// ErrIllegalPage : 非法的页面
	ErrIllegalPage
	// 	ErrIllegalPageSize : 非法的页大小
	ErrIllegalPageSize
)

// common: authorization 认证模块
const (
	// ErrTokenExpired : token过期
	ErrTokenExpired int = iota + 100101
	// ErrTokenInvalid : 无效的token
	ErrTokenInvalid
	// ErrInvalidAuthHeader :无效的 认证头
	ErrInvalidAuthHeader
	// ErrMissingHeader 缺少请求头
	ErrMissingHeader
)

// common: database errors 数据库模块.
const (
	// ErrDatabase : Database error.
	ErrDatabase int = iota + 100201
)

// common: encode/decode errors 编解码模块.
const (
	// ErrEncodingFailed : Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301

	// ErrDecodingFailed : Decoding failed due to an error with the data.
	ErrDecodingFailed

	// ErrInvalidJSON : Data is not valid JSON.
	ErrInvalidJSON

	// ErrEncodingJSON : JSON data could not be encoded.
	ErrEncodingJSON

	// ErrDecodingJSON  JSON data could not be decoded.
	ErrDecodingJSON

	// ErrInvalidYaml : Data is not valid Yaml.
	ErrInvalidYaml

	// ErrEncodingYaml : Yaml data could not be encoded.
	ErrEncodingYaml

	// ErrDecodingYaml : Yaml data could not be decoded.
	ErrDecodingYaml
)
