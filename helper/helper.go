package helper

import "fmt"

func CheckParameter(param string) {
	if len(param) == 0 {
		panic(fmt.Sprintf("参数错误"))
	}
}
