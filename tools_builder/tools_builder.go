package tools_builder

import (
	"flag"
	"fmt"
	"github.com/lwm-galactic/logger"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

type funcNode struct {
	Name  string
	Usage string
}

// 打印必要参数提示,并退出
func showUsage() {
	fmt.Printf("\n%s -f <tools>.go\n\n", os.Args[0])
	os.Exit(1)
}

// 定义常量，用于识别注释中的用途描述关键字
const usageTxt = "usage:"

// flag 包是 Go 标准库中用于解析命令行参数 的工具包。
// 它允许你定义命令行标志（flags），并自动处理用户输入，将字符串、整数、布尔值等类型的参数解析到程序变量中。
func main() {
	logger.SetModName("tools_builder")
	// 解析参数 -f 吗 -f 参数是一个go源码文件名
	fnPtr := flag.String("f", "", "tools source file")
	flag.Parse()
	if *fnPtr == "" {
		showUsage()
	}

	fn := *fnPtr
	// token 是 Go 标准库中的一个包，用于处理源码文件中的“词法单元”（tokens），也就是源代码中最小有意义的语法单位。
	fileSet := token.NewFileSet()
	// go/parser 是 Go 标准库中的一个包，用于将 Go 源代码文件解析为抽象语法树（AST）。
	// fileSet 管理多个源文件的位置信息（来自 token包）
	// 要解析的 Go 源文件名（如 "main.go"）
	// 可选参数：可以传入源码字符串或io.Reader；这里为nil表示从文件读取
	// parser.ParseComments 解析模式：表示要保留注释
	f, err := parser.ParseFile(fileSet, fn, nil, parser.ParseComments)
	if err != nil { // 解析源码文件错误
		logger.Fatalf("parse file error, fn %s, err %s", fn, err)
	}

	pkgName := f.Name
	if pkgName == nil || pkgName.Name != "main" { // 源码包名必须是main方法
		logger.Fatalf("package name must be 'main'")
	}

	// chdir命令用于更改当前工作目录，切换到源码的工作目录
	dir := filepath.Dir(fn)
	if dir != "." {
		err := os.Chdir(dir)
		if err != nil {
			fmt.Printf("chdir to %s failed, %s", dir, err)
			os.Exit(1)
		}
	}
	base := filepath.Base(fn)
	toolsName := base
	if strings.HasSuffix(base, ".go") {
		toolsName = base[:len(base)-3]
	}
	outPath := fmt.Sprintf("%s_autogen.go", toolsName) // 文件重命名
	logger.Infof("out path %s", outPath)               // 输出位置
	// 将源码的 方法拷贝出来
	funcList := make([]*funcNode, 0)
	// 遍历文件中的所有声明（如：变量、函数、类型等）
	for _, decl := range f.Decls {
		// 判断当前声明是否为函数声明
		switch t := decl.(type) {
		case *ast.FuncDecl:
			// 获取函数名
			funcName := t.Name.Name

			// 如果函数名为空，则跳过
			if funcName == "" {
				continue
			}

			// 只保留导出函数（首字母大写）
			if !unicode.IsUpper(rune(funcName[0])) {
				logger.Infof("- skip %s, first char not upper", funcName)
				continue
			}

			// 排除方法（即带有接收者的函数）
			if t.Recv != nil {
				logger.Infof("- skip %s, only support raw function", funcName)
				continue
			}

			// 排除有参数的函数
			if len(t.Type.Params.List) > 0 {
				logger.Infof("- skip %s, params not empty", funcName)
				continue
			}

			// 初始化 usage 描述字段
			usage := ""

			// 如果函数有文档注释，则尝试从中提取 usage 信息
			if t.Doc != nil {
				for _, comment := range t.Doc.List {
					txt := comment.Text

					// 查找 "usage:" 关键字
					p := strings.Index(txt, usageTxt)
					if p >= 0 {
						// 提取关键字后的文本内容，并去除前后空格
						usage = strings.TrimSpace(txt[p+len(usageTxt):])
					}
				}
			}

			// 记录当前找到的函数及其描述信息
			logger.Infof("func name %s, usage %s", funcName, usage)

			// 添加到函数列表中
			funcList = append(funcList, &funcNode{
				Name:  funcName,
				Usage: usage,
			})
		}
	}

	if len(funcList) == 0 {
		logger.Info("func list empty, skip")
		return
	}

	// 重新拼接文件
	buf := ""
	imp := `package main
import (
"github.com/lwm-galactic/tools/tools_builder/tools_lib"
)
`
	buf += imp
	wrapperTemplate := `
func wrapper%s() {
%s()
}
`
	mainTemplate := `
func main() {
%s
tools_lib.Run()
}
`
	for _, f := range funcList {
		buf += fmt.Sprintf(wrapperTemplate, f.Name, f.Name)
	}
	regList := make([]string, 0)
	for _, f := range funcList {
		regList = append(
			regList,
			fmt.Sprintf("\ttools_lib.Register(\"%s\", `%s`, wrapper%s)",
				f.Name, f.Usage, f.Name))
	}
	regBlock := strings.Join(regList, "\n")
	buf += fmt.Sprintf(mainTemplate, regBlock)

	os.WriteFile(outPath, []byte(buf), 0644)
	logger.Infof("gen %s success", outPath)
	binName := toolsName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	// build
	cmd := exec.Command(
		"go", "build", "-o", binName,
		fmt.Sprintf("%s.go", toolsName),
		fmt.Sprintf("%s_autogen.go", toolsName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		println("build error", err.Error(), string(out))
		os.Exit(1)
	}
	print(string(out))
	logger.Info("build success")
}
