package tools_lib

import (
	"fmt"
	"github.com/lwm-galactic/logger"
	"os"
	"path"
	"strings"
	"sync"
)

type Func func()
type FuncNode struct {
	Name  string // 函数名（命令名）
	Usage string // 用法描述
	Func  Func   // 实际要执行的函数
}

// 存储所有注册的函数信息（名称、用法、函数本身）
var fList []*FuncNode

// 互斥锁，用于保证并发安全地注册函数
var mu sync.Mutex

// UsageTail 使用说明的附加文本，显示在帮助信息末尾
var UsageTail string

// DefaultMethod 默认执行的方法名称（如果用户未指定）
var DefaultMethod string

var optMap = make(map[string]string)

// Register
// 将一个函数注册为可用命令。
// 加锁防止并发写入冲突。
// 把命令添加到全局命令列表 fList 中。
func Register(name, usage string, f Func) {
	mu.Lock()
	defer mu.Unlock()
	fList = append(fList, &FuncNode{Name: name, Usage: usage, Func: f})
}

// 打印完整的命令行使用帮助信息，并退出程序。
func showUsage() {
	fmt.Println()
	for _, f := range fList {
		fmt.Printf("\t-f %s %s\n", f.Name, f.Usage)
	}
	if UsageTail != "" {
		fmt.Printf("\n%s\n", UsageTail)
	}
	if DefaultMethod != "" {
		fmt.Println()
		fmt.Printf("default action: %s\n", DefaultMethod)
	}
	fmt.Println()
	os.Exit(1)
}

func Run() {
	// 遍历命令行参数（跳过第一个参数，即程序自身）
	for i := 1; i < len(os.Args); i++ {
		// 如果用户输入了 -h 或 --help，则显示使用说明并退出
		if os.Args[i] == "-h" || os.Args[i] == "--help" {
			showUsage()
		}

		// 判断当前参数是否以 '-' 开头，表示是一个选项
		if strings.HasPrefix(os.Args[i], "-") {
			// 去掉前缀 '-', 获取选项名称
			opt := strings.Trim(os.Args[i], "-")

			// 查看下一个参数是否存在
			if i+1 < len(os.Args) {
				// 如果下一个参数也以 '-' 开头，说明当前选项没有值，设为空字符串
				if strings.HasPrefix(os.Args[i+1], "-") {
					optMap[opt] = ""
				} else {
					// 否则将下一个参数作为该选项的值保存
					optMap[opt] = os.Args[i+1]
				}
			} else {
				// 如果已经是最后一个参数，没有值，设为空字符串
				optMap[opt] = ""
			}
		}
	}

	// 检查是否传入了 "-f" 参数（用于指定要运行的函数/命令）
	f, ok := optMap["f"]

	// 如果没有传 "-f"，并且设置了默认方法，则使用默认方法
	if !ok {
		if DefaultMethod != "" {
			f = DefaultMethod
		} else {
			// 如果既没有 "-f" 又没有默认方法，打印错误并显示用法
			fmt.Printf("\nmissed -f opt\n\n")
			showUsage()
		}
	}

	// 获取程序名，用于设置日志模块名称
	name := path.Base(strings.Replace(os.Args[0], "\\", "/", -1))
	if strings.HasPrefix(name, "./") {
		name = name[2:] // 去掉 "./" 前缀
	}
	logger.SetModName(name)

	// 查找注册过的函数中是否有匹配的名字（忽略大小写）
	var fu Func
	for _, x := range fList {
		if strings.ToLower(x.Name) == strings.ToLower(f) {
			fu = x.Func
			break
		}
	}

	// 如果找不到对应的函数，提示错误并显示帮助信息
	if fu == nil {
		fmt.Printf("\nnot found func %s\n\n", f)
		showUsage()
	}

	// 注册系统信号处理（例如 Ctrl+C），用于优雅退出
	regExitSignals()

	// 调用找到的函数，开始执行具体功能
	fu()
}
