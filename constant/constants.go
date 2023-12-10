package constant

import "os"

var GateWay string

func Init(defaultValue string) {
	GateWay = os.Getenv("ZTE_GATEWAY")

	if GateWay == "" {
		GateWay = defaultValue
	}
}
