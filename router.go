package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

const serverURL = "http://localhost:8080/heartbeat" // 请更改为你的服务器地址

func main() {
	go sendHeartbeats()

	// 为了让主程序不退出，使用一个无限循环
	select {}
}

func sendHeartbeats() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一个心跳

	for {
		select {
		case <-ticker.C:
			sendHeartbeat()
		}
	}
}

func sendHeartbeat() {
	// 这里只是一个例子，你可以根据你的需求调整这个心跳数据结构
	heartbeatData := `{
  "name": "service-3",
  "host": "localhost",
  "port": 8000
}`
	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer([]byte(heartbeatData)))
	if err != nil {
		fmt.Println("Failed to send heartbeat:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Server returned non-OK status:", resp.Status)
	} else {
		fmt.Println("Heartbeat sent successfully!")
	}
}
