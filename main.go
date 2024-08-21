package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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
	soloRanks := map[int]string{
		1:  "Bronze 1",
		2:  "Bronze 2",
		3:  "Bronze 3",
		4:  "Bronze 4",
		5:  "Silver 1",
		6:  "Silver 2",
		7:  "Silver 3",
		8:  "Silver 4",
		9:  "Gold 1",
		10: "Gold 2",
		11: "Gold 3",
		12: "Gold 4",
		13: "Platinum 1",
		14: "Platinum 2",
		15: "Platinum 3",
		16: "Platinum 4",
		17: "Emerald 1",
		18: "Emerald 2",
		19: "Emerald 3",
		20: "Emerald 4",
		21: "Ruby 1",
		22: "Ruby 2",
		23: "Ruby 3",
		24: "Ruby 4",
		25: "Diamond 1",
		26: "Diamond 2",
		27: "Diamond 3",
		28: "Diamond 4",
		29: "Champion"}
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println(message.Channel + " :: " + message.User.DisplayName + " : " + message.Message)
		if message.Message == "!spectrestats" {
			var playerId string
			if message.Channel == "ethos" {
				playerId = "E27C1FD1-4EEB-483D-952D-A7C904869509"
			}
			if message.Channel == "truo" {
				playerId = "8D02F2C0-69B8-4CEE-9656-2D0866B44E9B"
			}
			if message.Channel == "staycationtg" {
				playerId = "F0CD9516-6DFB-4235-8E04-32D6B820754C"
			}
			if message.Channel == "limitediq__" {
				playerId = strings.Split(message.Message, ":")[1]
			}
			player := GetPlayerMatchmakingDataBody{
				PlayerId: playerId,
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
			rank, rankPrs := soloRanks[stats.Response.Payload.Data.CurrentSoloRank]
			if rankPrs {
				twitchMessage := fmt.Sprintf("[Solo Rank]: %s [Solo Season Ranked Matches]: %d [Solo Season Ranked Wins]: %d", rank, stats.Response.Payload.Data.RankedMatchesPlayedCount, stats.Response.Payload.Data.RankedMatchesWonCount)
				client.Reply(message.Channel, message.ID, twitchMessage)
			} else {
				twitchMessage := fmt.Sprintf("Solo Rank: Unranked Solo Season Ranked Matches: %d Solo Season Ranked Wins: %d", stats.Response.Payload.Data.RankedMatchesPlayedCount, stats.Response.Payload.Data.RankedMatchesWonCount)
				client.Reply(message.Channel, message.ID, twitchMessage)
			}
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
