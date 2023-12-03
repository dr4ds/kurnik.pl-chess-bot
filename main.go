package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"main/kurnik"
	"main/uci"
	"main/utils"
)

var settingsFilePath = flag.String("settings", "settings.json", "path to settings file")

func LoadBotSettings(path string) (kurnik.BotSettings, error) {
	s := kurnik.BotSettings{}

	err := utils.LoadJsonFile(path, &s)

	return s, err
}

func SaveBotSettings(bs kurnik.BotSettings) error {
	b, err := json.Marshal(bs)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(*settingsFilePath, b, os.ModePerm)
	return err
}

func CreateBotFromSettings(bs kurnik.BotSettings) *kurnik.KurnikBot {
	e := new(uci.ChessEngine)
	err := e.LoadEngine(bs.EnginePath)
	if err != nil {
		panic(err)
	}

	err = e.SetOptions(bs.EngineOptions)
	if err != nil {
		panic(err)
	}

	bot := new(kurnik.KurnikBot)
	bot.Engine = e
	bot.BotSettings = bs

	return bot
}

func main() {
	flag.Parse()

	settings, err := LoadBotSettings(*settingsFilePath)
	if err != nil {
		panic(err)
	}

	bot := CreateBotFromSettings(settings)
	bot.ConnectToWebSocketServer()

	go bot.StartListening()

	// TODO: remove after web dashboard is done
	if bot.BotSettings.Account.Login != "" && bot.BotSettings.Account.Password != "" {
		bot.Login(bot.BotSettings.Account.Login, bot.BotSettings.Account.Password)
	} else {
		bot.LoginAsGuest()
	}
	time.Sleep(time.Second * 1)

	bot.JoinSection("#100...")
	time.Sleep(time.Second * 1)
	bot.CreateRoom()

	time.Sleep(time.Second * 1)
	bot.TakeSeat(0)

	bot.StartWebServer()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	bot.Exit()
}
