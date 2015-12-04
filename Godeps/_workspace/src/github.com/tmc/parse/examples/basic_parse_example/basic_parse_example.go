package main

import (
	"fmt"
	"os"

	"github.com/tmc/parse"
)

type GameScore struct {
	parse.ParseObject
	CheatMode  bool    `json:"cheatMode,omitempty"`
	PlayerName string  `json:"playerName,omitempty"`
	Score      float64 `json:"score,omitempty"`
}

func main() {
	appID := os.Getenv("APPLICATION_ID")
	apiKey := os.Getenv("REST_API_KEY")
	client, _ := parse.NewClient(appID, apiKey)

	objID, err := client.Create(&GameScore{
		CheatMode:  true,
		PlayerName: "Sean Plott",
		Score:      1337,
	})
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("Created", objID)
}
