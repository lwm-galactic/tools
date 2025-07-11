package registry

import (
	"context"
	"crypto/tls"

	"time"
)

// Options 配置中心的配置.
type Options struct {
	// 必填项 - etcd 集群节点地址
	Endpoints []string `json:"endpoints" mapstructure:"endpoints"`

	// 连接超时时间
	DialTimeout time.Duration `json:"dial-timeout" mapstructure:"dial_timeout"`

	// 安全相关配置
	DialKeepAliveTime    time.Duration `json:"dial-keep-alive-time" mapstructure:"dial-keep-alive-time"`
	DialKeepAliveTimeout time.Duration `json:"dial-keep-alive-timeout" mapstructure:"dial-keep-alive-timeout"`

	// TLS 配置
	TLS *tls.Config

	// 用户名密码认证
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`

	// 连接池配置
	MaxCallSendMsgSize int `json:"max-send-bytes" mapstructure:"max-send-bytes"` // 单个请求最大大小(默认 2MB)
	MaxCallRecvMsgSize int `json:"max-recv-bytes" mapstructure:"max-send-bytes"` // 单个响应最大大小

	// 重试策略
	ReconnectWait time.Duration // 重连等待时间
	BackoffWait   time.Duration // 退避等待时间

	// 上下文配置
	Context context.Context // 控制客户端生命周期的上下文
}

// Option 配置函数类型
type Option func(*Options)

// NewOptions 创建配置实例
func NewOptions(opts ...Option) *Options {
	// 设置默认值
	options := &Options{
		Endpoints:            []string{"localhost:2379"},
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    30 * time.Second,
		DialKeepAliveTimeout: 10 * time.Second,
		MaxCallSendMsgSize:   2 * 1024 * 1024, // 2MB
		MaxCallRecvMsgSize:   4 * 1024 * 1024, // 4MB
		ReconnectWait:        200 * time.Millisecond,
		BackoffWait:          1 * time.Second,
		Context:              context.Background(),
	}

	// 应用配置函数
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithEndpoints 设置 etcd 集群地址
func WithEndpoints(endpoints []string) Option {
	return func(o *Options) {
		o.Endpoints = endpoints
	}
}

// WithDialTimeout 设置连接超时时间
func WithDialTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = timeout
	}
}

// WithTLS 设置 TLS 配置
func WithTLS(tlsConfig *tls.Config) Option {
	return func(o *Options) {
		o.TLS = tlsConfig
	}
}

// WithAuth 设置用户名密码认证
func WithAuth(username, password string) Option {
	return func(o *Options) {
		o.Username = username
		o.Password = password
	}
}

// WithKeepAlive 设置保活参数
func WithKeepAlive(timeout, keepAliveTime time.Duration) Option {
	return func(o *Options) {
		o.DialKeepAliveTimeout = timeout
		o.DialKeepAliveTime = keepAliveTime
	}
}

// WithMessageSize 设置消息大小限制
func WithMessageSize(sendSize, recvSize int) Option {
	return func(o *Options) {
		o.MaxCallSendMsgSize = sendSize
		o.MaxCallRecvMsgSize = recvSize
	}
}

// WithRetryPolicy 设置重试策略
func WithRetryPolicy(reconnectWait, backoffWait time.Duration) Option {
	return func(o *Options) {
		o.ReconnectWait = reconnectWait
		o.BackoffWait = backoffWait
	}
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}
