# 📘 Go Validate 工具包学习笔记

## 一、简介

`go-playground/validator` 是一个非常流行且功能强大的 **结构体字段校验库** ，用于对 Go 中的结构体进行数据合法性检查。它支持丰富的内置规则（如非空、长度、邮箱格式等），也支持自定义校验规则与多语言错误提示。

GitHub 地址：https://github.com/go-playground/validator

## 二、安装方式

```bash
go get github.com/go-playground/validator/v10
```



## 三、基本使用示例

### 1. 定义结构体并添加校验 tag

```go
type User struct {
    Name  string `validate:"required,min=2,max=10"`
    Email string `validate:"required,email"`
    Age   uint   `validate:"gte=0,lte=150"`
}
```



### 2. 创建 validator 实例并执行校验

```go
import "github.com/go-playground/validator/v10"

validate := validator.New()

user := User{
    Name:  "",
    Email: "invalid-email",
}

err := validate.Struct(user)
if err != nil {
    fmt.Println(err)
}
```



### 3. 输出结果示例

```text
Key: 'User.Name' Error:Field validation for 'Name' failed on the 'required' tag
Key: 'User.Email' Error:Field validation for 'Email' failed on the 'email' tag
```



## 四、常用 tag 规则说明

| TAG              | 描述                     |
| ---------------- | ------------------------ |
| `required`       | 字段必须存在且不为空     |
| `omitempty`      | 如果字段为空，则跳过校验 |
| `min=5`,`max=10` | 最小值/最大值限制        |
| `gt=10`,`lt=20`  | 大于/小于指定值          |
| `email`          | 必须是合法的邮箱格式     |
| `url`            | 必须是一个合法的 URL     |
| `len=5`          | 长度必须为 5             |
| `alphanum`       | 只能包含字母数字         |
| `uuid`           | 必须是合法的 UUID 格式   |

> 更多 tag 支持请参考官方文档。 

## 五、进阶用法

### 1. 自定义校验函数

你可以注册自己的校验逻辑来处理特定业务需求：

```go
func validateCustom(fl validator.FieldLevel) bool {
    return fl.Field().String() == "valid"
}

validate.RegisterValidation("custom", validateCustom)
```

然后在结构体中使用：

```go
type MyStruct struct {
    Field string `validate:"custom"`
}
```



### 2. 获取结构化错误信息

```go
if err != nil {
    if _, ok := err.(*validator.InvalidValidationError); ok {
        log.Fatal(err)
    }

    for _, err := range err.(validator.ValidationErrors) {
        fmt.Printf("Field: %s, Tag: %s, Value: %v\n", err.Field(), err.Tag(), err.Value())
    }
}
```

输出示例：

```go
Field: Name, Tag: required, Value: 
Field: Email, Tag: email, Value: invalid-email
```



## 六、国际化支持（i18n）

`validator` 支持通过 `universal-translator` 进行错误信息翻译。

### 示例：启用中文错误提示

```go
import (
    zh "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    zh_trans "github.com/go-playground/validator/v10/translations/zh"
)

zhCn := zh.New()
uni := ut.New(zhCn, zhCn)
trans, _ := uni.GetTranslator("zh")

validate := validator.New()
_ = zh_trans.RegisterDefaultTranslations(validate, trans)

err := validate.Struct(myStruct)
if err != nil {
    validationErrs := err.(validator.ValidationErrors)
    for _, e := range validationErrs {
        fmt.Println(e.Translate(trans))
    }
}
```

输出示例：

```text
Name: 长度必须至少为 2 个字符
Email: 必须是一个有效的电子邮件地址
```



## 七、封装建议（适用于项目中统一使用）

推荐将 `validator` 封装成一个统一的校验工具类，便于复用和管理错误提示、自定义规则等。

### 封装常用数据结构校验

[地址]: https://github.com/lwm-galactic/tools/validation	"”数据结构校验示例"



#### 主要功能模块：

| 模块                                      | 功能                                                 | 校验规则                                                     |
| ----------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------ |
| `IsQualifiedName`                         | 校验是否是合法的“qualified name”（带命名空间的名称） | 可以包含字母数字、下划线 `_`、点 `.` 和连字符 `-` 必须以字母或数字开头和结尾 最大长度限制为 **63 字符** 支持带前缀形式：`<prefix>/<name>`，其中前缀也必须是合法的 DNS 子域名格式 |
| `IsValidLabelValue`                       | 校验 label 值是否合法                                | 可为空字符串 或者符合 QualifiedName 的格式 最大长度**63 字符** |
| `IsDNS1123Label`                          | 校验 DNS-1123 标签格式                               | 小写字母、数字、连字符 `-` 不能以 `-` 开头或结尾 最长 **63 字符** |
| `IsDNS1123Subdomain`                      | 校验子域名格式                                       | 多个 DNS-1123 标签用 `.` 连接组成 总长度不超过 **253 字符**  |
| `IsValidPortNum`                          | 校验端口号是否在 1~65535 范围内                      | **1 ~ 65535**                                                |
| `IsValidIP`                               | 校验是否为合法 IP 地址                               | IPv4 或 IPv6 地址                                            |
| `IsValidIPv4Address`,`IsValidIPv6Address` | 分别校验 IPv4/IPv6 地址                              | IsValidIPv4Address 确保是 IPv4 地址（不含 IPv6）\| IsValidIPv6Address 确保是 IPv6 地址（不含 IPv4） |
| `IsValidPercent`                          | 校验是否是百分比格式（如 "90%"）                     | 数字 + `%` 结尾                                              |
| `IsValidPassword`                         | 校验密码是否符合复杂度要求                           | 至少包含一个大写字母 至少包含一个小写字母 至少包含一个数字 至少包含一个特殊符号（标点或符号类 Unicode） 长度在 **8 ~ 16 字符之间** |

### 返回统一错误格式封装

 **结构化字段错误处理库** ，用于在 Go 项目中进行 **字段级（field-level）的错误校验和报告** 。它常用于 API 接口参数校验、配置文件验证、Kubernetes 风格的资源校验等场景。

#### 📦 整体功能概述

| 类型                                  | 功能                               |
| ------------------------------------- | ---------------------------------- |
| `Error`                               | 表示一个字段级别的错误             |
| `ErrorType`                           | 错误类型，如 Required、Invalid 等  |
| `ErrorList`                           | 多个 Error 的集合                  |
| `NotFound`,`Required`,`Invalid`等函数 | 快捷构造器，创建特定类型的字段错误 |
| `ToAggregate`,`Filter`等方法          | 对多个错误进行聚合和过滤           |

#### 🔧 核心结构体详解

##### `Error`：字段错误信息结构体

```go
type Error struct {
    Type     ErrorType 
    Field    string
    BadValue interface{}
    Detail   string
}
```

> - `Type`: 错误类型，如 `ErrorTypeRequired`、`ErrorTypeInvalid`
> - `Field`: 出错的字段路径（如 `"User.Address.City"`）
> - `BadValue`: 出错的具体值
> - `Detail`: 更详细的错误说明

##### `ErrorType`：错误类型枚举

常见类型如下：

| 类型                    | 含义                                 |
| ----------------------- | ------------------------------------ |
| `ErrorTypeRequired`     | 必填字段为空                         |
| `ErrorTypeInvalid`      | 值不合法（格式错误、超出长度等）     |
| `ErrorTypeNotSupported` | 不支持的值（比如枚举值不在白名单内） |
| `ErrorTypeForbidden`    | 被禁止的值（权限不足或策略限制）     |
| `ErrorTypeTooLong`      | 字段太长                             |
| `ErrorTypeTooMany`      | 列表项太多                           |
| `ErrorTypeNotFound`     | 找不到该值                           |
| `ErrorTypeInternal`     | 内部错误                             |

#### ✅ 构造错误的方法（工厂函数）

这些函数用来快速创建特定类型的错误对象：

| 方法                                            | 示例                                                         |
| ----------------------------------------------- | ------------------------------------------------------------ |
| `Required(field *Path, detail string)`          | `"User.Name": Required value`                                |
| `Invalid(field *Path, value, detail)`           | `"User.Age": Invalid value: "abc"`                           |
| `NotFound(field *Path, value)`                  | `"User.Role": Not found`                                     |
| `Forbidden(field *Path, detail)`                | `"User.Role": Forbidden`                                     |
| `TooLong(field *Path, value, maxLength)`        | `"Description": Too long: must have at most 255 bytes`       |
| `TooMany(field *Path, actual, max)`             | `"Tags": Too many: must have at most 10 items`               |
| `NotSupported(field *Path, value, validValues)` | `"Image": Unsupported value: 'alpine', supported values: 'ubuntu', 'centos'` |
| `InternalError(field *Path, err)`               | `"Config": Internal error: failed to parse JSON`             |

## 八、常见问题与注意事项

| 问题                           | 解决方法                                     |
| ------------------------------ | -------------------------------------------- |
| 如何忽略某些字段的校验？       | 使用`omitempty`tag                           |
| 结构体嵌套如何处理？           | 默认支持嵌套结构体                           |
| 校验 slice 或 map？            | 使用`dive`tag，如`validate:"dive,required"`  |
| 如何获取字段名？               | 使用`fe.Field()`获取字段名                   |
| 如何区分字段和结构体标签错误？ | 检查`fe.StructNamespace()`和`fe.Namespace()` |
| 如何避免 panic？               | 注意不要传入`nil`或非结构体类型              |

## 九、总结

| 功能           | 支持情况   |
| -------------- | ---------- |
| 内置校验规则   | ✅ 非常丰富 |
| 自定义规则     | ✅ 支持     |
| 错误结构化输出 | ✅ 支持     |
| 多语言支持     | ✅ 支持     |
| 嵌套结构体校验 | ✅ 支持     |
| Slice/Map 校验 | ✅ 支持     |
| Gin/Echo 集成  | ✅ 支持     |

## 十、参考资料

- GitHub 主页：https://github.com/go-playground/validator
- 官方文档：https://pkg.go.dev/github.com/go-playground/validator/v10
- i18n 翻译支持：https://github.com/go-playground/universal-translator