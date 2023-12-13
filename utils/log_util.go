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
		Error(fmt.Sprintf("Recovered from panic: %s", r))
	}
}
