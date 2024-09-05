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

type GetPlayerCrewData struct {
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

type PlayerCrewData struct {
	PlayerID            string `json:"playerId"`
	PlayerDisplayName   string `json:"playerDisplayName"`
	PlayerDiscriminator string `json:"playerDiscriminator"`
	PlayerCrewScore     string `json:"playerCrewScore"`
	CrewID              string `json:"crewId"`
	CrewTotalScore      string `json:"crewTotalScore"`
	CrewDivisionType    string `json:"crewDivisionType"`
	CrewDivisionRank    int    `json:"crewDivisionRank"`
	CrewGlobalRank      int    `json:"crewGlobalRank"`
	CrewTotalCrews      int    `json:"crewTotalCrews"`
}

func main() {
	//client := twitch.NewAnonymousClient()

	// JoinChannel := make(chan string)
	TWITCH_KEY := os.Getenv("TWITCH_KEY")
	client := twitch.NewClient("Santaigg", TWITCH_KEY)
	// soloRanks := map[int]string{
	// 	1:  "Bronze 1",
	// 	2:  "Bronze 2",
	// 	3:  "Bronze 3",
	// 	4:  "Bronze 4",
	// 	5:  "Silver 1",
	// 	6:  "Silver 2",
	// 	7:  "Silver 3",
	// 	8:  "Silver 4",
	// 	9:  "Gold 1",
	// 	10: "Gold 2",
	// 	11: "Gold 3",
	// 	12: "Gold 4",
	// 	13: "Platinum 1",
	// 	14: "Platinum 2",
	// 	15: "Platinum 3",
	// 	16: "Platinum 4",
	// 	17: "Emerald 1",
	// 	18: "Emerald 2",
	// 	19: "Emerald 3",
	// 	20: "Emerald 4",
	// 	21: "Ruby 1",
	// 	22: "Ruby 2",
	// 	23: "Ruby 3",
	// 	24: "Ruby 4",
	// 	25: "Diamond 1",
	// 	26: "Diamond 2",
	// 	27: "Diamond 3",
	// 	28: "Diamond 4",
	// 	29: "Champion"}
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println(message.Channel + " :: " + message.User.DisplayName + " : " + message.Message)
		if strings.Contains(message.Message, "!crewstats") {
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
			if message.Channel == "bugzvii" {
				playerId = "30C3E8E8-B5A4-4461-B77C-567B9B3C762D"
			}
			if message.Channel == "steazecs" {
				playerId = "BCD9F729-DA28-4802-8CF6-DE831B852D62"
			}
			if message.Channel == "limitediq__" {
				playerId = strings.Split(message.Message, ":")[1]
				playerId = strings.TrimSpace(playerId)
			}
			if message.Channel == "moepork" {
				playerId = "39F848C1-A9A5-42DF-81AA-033191455DAA"
			}
			if message.Channel == "relyks" {
				playerId = "DC5D1993-5B94-4F0C-8F57-DB51B0DAE7F1"
			}
			if message.Channel == "shroud" {
				playerId = "CE4C88F7-7D66-417F-A3F5-01D0F9F52B90"
			}
			if message.Channel == "iitztimmy" {
				playerId = "1d36bff3-1ac5-422f-bb21-f6524e0b83a0"
			}
			if message.Channel == "pieman" {
				playerId = "a666813a-5cc1-48ac-bcdf-ac937bda38bf"
			}

			// Previous GetPlayerMatchmakingDataBody ****

			// player := GetPlayerMatchmakingDataBody{
			// 	PlayerId: playerId,
			// }
			// // marshall data to json (like json_encode)
			// playerBody, err := json.Marshal(player)
			// if err != nil {
			// 	log.Fatalf("impossible to marshall player: %s", err)
			// }
			// resp, err := http.Post("https://collective-production.up.railway.app/dev-getPlayerMatchmakingData", "application/json", bytes.NewReader(playerBody))
			// if err != nil {
			// 	log.Fatalf("Couldn't get player matchmaking data for: %s", player.PlayerId)
			// }
			// defer resp.Body.Close()
			// // read body
			// resBody, err := io.ReadAll(resp.Body)
			// if err != nil {
			// 	log.Fatalf("impossible to read all body of response: %s", err)
			// }
			// log.Printf("res body: %s", string(resBody))
			// stats := MatchmakingData{}
			// json.Unmarshal(resBody, &stats)
			// rank, rankPrs := soloRanks[stats.Response.Payload.Data.CurrentSoloRank]
			// if rankPrs {
			// 	twitchMessage := fmt.Sprintf("[Solo Rank]: %s [Solo Season Ranked Wins]: %d/%d games (%.1f%%)", rank, stats.Response.Payload.Data.RankedMatchesWonCount, stats.Response.Payload.Data.RankedMatchesPlayedCount, (float64(stats.Response.Payload.Data.RankedMatchesWonCount) / float64(stats.Response.Payload.Data.RankedMatchesPlayedCount) * 100))
			// 	client.Reply(message.Channel, message.ID, twitchMessage)
			// } else {
			// 	twitchMessage := fmt.Sprintf("[Solo Rank]: Unranked [Solo Season Ranked Wins]: %d/%d games (%.1f%%)", stats.Response.Payload.Data.RankedMatchesWonCount, stats.Response.Payload.Data.RankedMatchesPlayedCount, (float64(stats.Response.Payload.Data.RankedMatchesWonCount) / float64(stats.Response.Payload.Data.RankedMatchesPlayedCount) * 100))
			// 	client.Reply(message.Channel, message.ID, twitchMessage)
			// }
			player := GetPlayerCrewData{
				PlayerId: playerId,
			}
			// marshall data to json (like json_encode)
			playerBody, err := json.Marshal(player)
			if err != nil {
				log.Fatalf("impossible to marshall player: %s", err)
			}
			resp, err := http.Post("https://collective-production.up.railway.app/getPlayerCrewData", "application/json", bytes.NewReader(playerBody))
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
			stats := PlayerCrewData{}
			json.Unmarshal(resBody, &stats)
			if stats.CrewGlobalRank > 0 && stats.CrewTotalCrews > 0 {
				twitchMessage := fmt.Sprintf("[Crew Division Rank]: %d [Player Crew Score]: %s [Total Crew Score]: %s [Global Crew Rank]: %d/%d Crews", stats.CrewDivisionRank, stats.PlayerCrewScore, stats.CrewTotalScore, stats.CrewGlobalRank, stats.CrewTotalCrews)
				client.Reply(message.Channel, message.ID, twitchMessage)
			} else {
				twitchMessage := fmt.Sprintf("[Crew Division Rank]: %d [Player Crew Score]: %s [Total Crew Score]: %s", stats.CrewDivisionRank, stats.PlayerCrewScore, stats.CrewTotalScore)
				client.Reply(message.Channel, message.ID, twitchMessage)
			}
		}

		if strings.Contains(message.Message, "!spectrestats") {
			twitchMessage := fmt.Sprintf("Use !crewstats to get %s's crew stats. !spectrestats will be re-implemented when ranked drops on Sept 10th.", message.Channel)
			client.Reply(message.Channel, message.ID, twitchMessage)
		}
	})

	client.Join("truo", "limitediq__", "staycationtg", "ethos", "bugzvii", "steazecs", "moepork", "relyks", "shroud", "iitztimmy", "pieman")

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
