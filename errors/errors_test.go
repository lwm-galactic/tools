package errors

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestError(t *testing.T) {
	// 创建基础错误
	err1 := New("something went wrong")
	fmt.Println("err1", err1)

	var val = "21312"
	// 格式化创建
	err2 := Errorf("value is invalid: %v", val)
	fmt.Println("err2", err2)
	// 包装已有错误
	err3 := Wrap(err2, "failed during processing")

	fmt.Println("err3", err3)
	// 添加错误码
	err4 := WithCode(500, "internal server error")
	fmt.Println("err4", err4)
	// 获取原始错误
	original := Cause(err4)
	fmt.Println(original)
	// 打印详细堆栈
	fmt.Printf("%+v\n", err3)
}

func TestGroup(t *testing.T) {
	agg := AggregateGoroutines(
		func() error { return os.Remove("tmp1") },
		func() error { return os.Remove("tmp2") },
		func() error { return os.Remove("tmp3") },
	)

	if agg != nil {
		log.Println("Some deletions failed:", agg)
	}

}
