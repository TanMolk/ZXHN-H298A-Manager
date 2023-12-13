package utils

import (
	"fmt"
	"log"
)

func Normal(str interface{}) {
	log.Printf("LOG: %s \n", str)
}

func Error(str interface{}) {
	log.Printf("ERROR: %s \n", str)
}

func RecoverHandler() {
	if r := recover(); r != nil {
		// 发生了 panic，r 包含了 panic 的信息
		Error(fmt.Sprintf("Recovered from panic: %s", r))
		// 可以在这里进行一些处理，如记录日志等
	}
}
