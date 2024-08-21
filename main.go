package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/gofiber/fiber/v2"
)

type TwitchChannel struct {
	Channel string `query:"channel"`
}

type GetPlayerMatchmakingDataBody struct {
	PlayerId string `json:"playerId"`
}

type MatchmakingData struct {
	SequenceNumber int `json:"sequenceNumber"`
	Response       struct {
		RequestID int    `json:"requestId"`
		Type      string `json:"type"`
		Payload   struct {
			Data struct {
				PlayerID                       string `json:"playerId"`
				CasualMmr                      int    `json:"casualMmr"`
				RankedMmr                      int    `json:"rankedMmr"`
				SoloRankPoints                 int    `json:"soloRankPoints"`
				CasualMatchesPlayedCount       int    `json:"casualMatchesPlayedCount"`
				RankedMatchesPlayedCount       int    `json:"rankedMatchesPlayedCount"`
				CasualMatchesPlayedSeasonCount int    `json:"casualMatchesPlayedSeasonCount"`
				RankedMatchesPlayedSeasonCount int    `json:"rankedMatchesPlayedSeasonCount"`
				RankedPlacementMatches         []int  `json:"rankedPlacementMatches"`
				CurrentSoloRank                int    `json:"currentSoloRank"`
				HighestTeamRank                int    `json:"highestTeamRank"`
				CasualMatchesWonCount          int    `json:"casualMatchesWonCount"`
				RankedMatchesWonCount          int    `json:"rankedMatchesWonCount"`
				PriorityMatchmakingUntil       string `json:"priorityMatchmakingUntil"`
				RestrictMatchmakingUntil       string `json:"restrictMatchmakingUntil"`
				MatchmakingFlags               int    `json:"matchmakingFlags"`
				MapHistory                     string `json:"mapHistory"`
			} `json:"data"`
		} `json:"payload"`
	} `json:"response"`
}

func main() {
	//client := twitch.NewAnonymousClient()

	// JoinChannel := make(chan string)
	TWITCH_KEY := os.Getenv("TWITCH_KEY")
	client := twitch.NewClient("Santaigg", TWITCH_KEY)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println(message.User.DisplayName + " : " + message.Message)
		if message.Message == "!spectrestats" {
			player := GetPlayerMatchmakingDataBody{
				PlayerId: "8D02F2C0-69B8-4CEE-9656-2D0866B44E9B",
			}
			// marshall data to json (like json_encode)
			playerBody, err := json.Marshal(player)
			if err != nil {
				log.Fatalf("impossible to marshall player: %s", err)
			}
			resp, err := http.Post("https://collective-production.up.railway.app/dev-getPlayerMatchmakingData", "application/json", bytes.NewReader(playerBody))
			if err != nil {
				log.Fatalf("Couldn't get player matchmaking data for: %s", player.PlayerId)
			}
			defer resp.Body.Close()
			// read body
			resBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("impossible to read all body of response: %s", err)
			}
			log.Printf("res body: %s", string(resBody))
			stats := MatchmakingData{}
			json.Unmarshal(resBody, &stats)
			twitchMessage := fmt.Sprintf("Solo Season Ranked Matches: %d Solo Season Ranked Wins: %d", stats.Response.Payload.Data.RankedMatchesPlayedCount, stats.Response.Payload.Data.RankedMatchesWonCount)
			client.Reply(message.Channel, message.ID, twitchMessage)
		}
	})

	client.Join("truo", "limitediq__")

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Patches!")
	})

	app.Get("/joinTwitch", func(c *fiber.Ctx) error {
		channel := new(TwitchChannel)
		if err := c.QueryParser(channel); err != nil {
			return err
		}

		client.Join(channel.Channel)

		return c.SendString("Joined " + channel.Channel)
	})

	go func() { log.Fatal(app.Listen(":3002")) }()

	err := client.Connect()
	if err != nil {
		panic(err)
	}
}
