package main

import (
	"flag"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/handler"
	"log"
	"net/http"
)

func main() {
	log.Println("服务启动")
	// 房间列表页
	http.HandleFunc("/home", handler.LoadHome)
	// 等待开始页
	http.HandleFunc("/room/", handler.LoadRoom)
	http.HandleFunc("/quitRoom/", handler.QuitRoom)
	http.HandleFunc("/startGame/", handler.StartGame)
	// 游戏中页
	http.HandleFunc("/game/", handler.LoadGame)
	http.HandleFunc("/gaming/", handler.Gaming)
	// 游戏结算页
	http.HandleFunc("/review/", handler.LoadReview)
	http.HandleFunc("/returnRoom/", handler.ReturnRoom)

	var mode string
	flag.StringVar(&mode, "mode", "dev", "dev or prod")
	flag.Parse()

	switch mode {
	case "dev":
		err := http.ListenAndServe(":62121", nil)
		if err != nil {
			log.Fatal("Error:", err)
		}
	case "prod":
		err := http.ListenAndServeTLS(":62121", "tls/bloodws.liuzhaomax.cn.pem", "tls/bloodws.liuzhaomax.cn.key", nil)
		if err != nil {
			log.Fatal("Error:", err)
		}
	}
}
