package service

import (
	"net/http"
	"time"
)

var httpClient = http.Client{
	Timeout: 5 * time.Second,
}

type RequestBody struct {
	Content string `json:"content"`
}

type Response struct {
	Success bool `json:"success"`
}

func ChangeDNSRecord(ipv6 string) bool {

	/*
		Change to your own dnd changing logic
	*/

	return false
}
