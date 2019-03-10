package main

import (
	"os"
	"os/signal"
	"time"
)

func main() {
	options := make(map[string]string)
	options["Threads"] = "4"
	options["Hash"] = "512"
	// options["OwnBook"] = "true"
	// options["BestBookMove"] = "true"

	e := new(ChessEngine)
	err := e.LoadEngine("engines/asmfish.exe")
	if err != nil {
		panic(err)
	}

	e.SetOptions(options)

	bot := new(KurnikBot)

	bot.Engine = e

	bot.ConnectToWebSocketServer()

	go bot.StartListening()

	bot.LoginAsGuest()

	time.Sleep(time.Second * 1)

	bot.JoinSection("#2300...")
	bot.CreateRoom()

	time.Sleep(time.Second * 1)
	bot.TakeSeat(0)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	bot.Running = false
	bot.Engine.Close()
	bot.Disconnect()
}
