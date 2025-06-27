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

