package main

import (
	"encoding/json"
	"fmt"
	constants "github.com/tanmolk/ZXHN-H298A-Manager/constant"
	"github.com/tanmolk/ZXHN-H298A-Manager/service"
	"github.com/tanmolk/ZXHN-H298A-Manager/utils"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	//Get admin page url
	constants.Init("http://192.168.1.1")

	//get password
	psw := os.Getenv("ZTE_ADMIN_PSW")
	if psw == "" {
		log.Fatal("env ZTE_ADMIN_PSW can't be empty")
	}

	//get running portStr
	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		portStr = "8080"
	}

	//get execute interval
	interval := os.Getenv("EXECUTE_INTERVAL")
	if interval == "" {
		interval = strconv.Itoa(60 * 5)
	}
	var interValNum int
	interValNum, _ = strconv.Atoi(interval)
	interValNum /= 2
	period := int64(interValNum) * time.Second.Nanoseconds()
	duration := time.Duration(period)

	// create http client
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
	}

	//create service
	zteService := service.NewZteService(
		constants.GateWay,
		"admin",
		psw,
		client,
	)

	//start server to receive signal
	var mutex sync.Mutex
	ch := make(chan bool)

	executeState := false

	//start execute thread
	go func() {
		for {
			mutex.Lock()
			<-ch

			//login
			err = zteService.Login()
			if err == nil {
				zteService.UpdateIPV6()
			} else {
				utils.Error("Login Fail")
			}

			mutex.Unlock()
		}
	}()

	//first sync
	ch <- true

	//start cron thread
	go func() {
		for {
			time.Sleep(duration)
			if executeState {
				executeState = false
				continue
			}
			executeState = true
			ch <- true
		}
	}()

	http.HandleFunc("/execute", func(writer http.ResponseWriter, request *http.Request) {
		ch <- true
		writer.Header().Set("Content-Type", "application/json")
		marshal, _ := json.Marshal(true)
		_, _ = writer.Write(marshal)

	})

	// start http server and run
	utils.Normal(fmt.Sprintf("Server is running on port %s", portStr))
	err = http.ListenAndServe(fmt.Sprintf(`:%s`, portStr), nil)

	if err != nil {
		log.Fatal(err)
	}
}
