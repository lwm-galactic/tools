# 工具说明文档



## tools_builder



### 示例

**输入**

```go
// tool.go
package main

import "fmt"

// usage: Say hello world
func Hello() {
    fmt.Println("Hello World")
}

// usage: Show current version
func Version() {
    fmt.Println("v1.0.0")
}
```

**输出**

```go
package main

import (
    "github.com/lwm-galactic/tools/tools_builder/tools_lib"
)

func wrapperHello() {
    Hello()
}

func wrapperVersion() {
    Version()
}

func main() {
    tools_lib.Register("Hello", `Say hello world`, wrapperHello)
    tools_lib.Register("Version", `Show current version`, wrapperVersion)
    tools_lib.Run()
}
```



## snow_flake

### 🧊 雪花算法简介

Snowflake 是 Twitter 开源的一种分布式 ID 生成算法，其核心思想是将一个 64 位的整数划分为几个部分：

| 名称      | 位数   | 含义                               |
| --------- | ------ | ---------------------------------- |
| sign      | 1bit   | 固定为0，不使用                    |
| timestamp | 41bits | 时间戳（毫秒），相对于某一时间起点 |
| machineId | 10bits | 工作节点ID（最多支持1024个节点）   |
| sequence  | 12bits | 同一毫秒内的序列号（最多4095个）   |

### 雪花算法存在的问题

#### **时间回拨问题（Time Backwards）**

- 如果服务器时间被 NTP 同步或手动调整，导致时间倒退，Snowflake 可能会生成重复的 ID。