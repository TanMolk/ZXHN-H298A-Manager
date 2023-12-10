package utils

import "log"

func Normal(str interface{}) {
	log.Printf("LOG: %s \n", str)
}

func Error(str interface{}) {
	log.Printf("ERROR: %s \n", str)
}
