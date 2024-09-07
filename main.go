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
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/nicklaw5/helix/v2"
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

type GetPlayerIdentityFromTwitchId struct {
	PlayerId string `json:"playerId"`
}

type GetPlayerRank struct {
	PlayerId string `json:"playerId"`
}

type PlayerRank struct {
	SoloRank int `json:"soloRank"`
	TeamRank int `json:"teamRank"`
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
	TWITCH_CLIENT_ID := os.Getenv("TWITCH_CLIENT_ID")
	TWITCH_SECRET := os.Getenv("TWITCH_SECRET")
	irc_client := twitch.NewClient("Santaigg", TWITCH_KEY)
	twitch_client, err := helix.NewClient(&helix.Options{
		ClientID:     TWITCH_CLIENT_ID,
		ClientSecret: TWITCH_SECRET,
	})
	if err != nil {
		log.Panicf("Issue logging into helix twitch client... %s", err)
	}

	resp, err := twitch_client.RequestAppAccessToken([]string{"user:read:email"})
	if err != nil {
		log.Panicf("Issue getting app access token from helix twitch client... %s", err)
	}

	fmt.Printf("%+v\n", resp)

	// Set the access token on the client
	twitch_client.SetAppAccessToken(resp.Data.AccessToken)

	var lastCrewDump time.Time
	irc_client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println(message.Channel + " :: " + message.User.DisplayName + "( " + message.User.ID + " )" + " : " + message.Message)
		if strings.ToLower(message.User.Name) == "santaigg" || strings.ToLower(message.User.DisplayName) == "santaigg" {
			return
		}

		if strings.Contains(message.Message, "!mycrewstats") {
			var playerId string
			playerIdReq := getPlayerIdFromTwitchId(message.User.ID)

			if playerIdReq != "ERROR" && playerIdReq != "" {
				playerId = playerIdReq
			} else {
				irc_client.Reply(message.Channel, message.ID, "You must link your Spectre Divide account to Twitch!")
				return
			}

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
				if stats.CrewGlobalRank < stats.CrewDivisionRank {
					stats.CrewGlobalRank = stats.CrewDivisionRank
				}
				twitchMessage := fmt.Sprintf("[Crew Division Rank]: %d [Player Crew Score]: %s [Total Crew Score]: %s [Global Crew Rank]: %d/%d Crews", stats.CrewDivisionRank, stats.PlayerCrewScore, stats.CrewTotalScore, stats.CrewGlobalRank, stats.CrewTotalCrews)
				irc_client.Reply(message.Channel, message.ID, twitchMessage)
			} else {
				twitchMessage := fmt.Sprintf("[Crew Division Rank]: %d [Player Crew Score]: %s [Total Crew Score]: %s", stats.CrewDivisionRank, stats.PlayerCrewScore, stats.CrewTotalScore)
				irc_client.Reply(message.Channel, message.ID, twitchMessage)
			}
		}

		if strings.Contains(message.Message, "!crewstats") {
			if hasTimePassed(lastCrewDump, 60*time.Second) {
				go func() {
					lastCrewDump = time.Now()
					_, initErr := http.Get("https://collective-production.up.railway.app/dumpAllCrewsFromDivisionsInDb")
					if initErr != nil {
						log.Println("Issue hitting dump ALL crews endpoint...")
					}
				}()
			}
			var playerId string
			playerIdReq := getPlayerIdFromChannel(message.Channel, twitch_client)
			if playerIdReq != "ERROR" && playerIdReq != "" {
				playerId = playerIdReq
			} else {
				messageReply := fmt.Sprintf("@%s must link their Spectre Divide account to Twitch!", message.Channel)
				irc_client.Reply(message.Channel, message.ID, messageReply)
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
				if stats.CrewGlobalRank < stats.CrewDivisionRank {
					stats.CrewGlobalRank = stats.CrewDivisionRank
				}
				twitchMessage := fmt.Sprintf("[Crew Division Rank]: %d [Player Crew Score]: %s [Total Crew Score]: %s [Global Crew Rank]: %d/%d Crews", stats.CrewDivisionRank, stats.PlayerCrewScore, stats.CrewTotalScore, stats.CrewGlobalRank, stats.CrewTotalCrews)
				irc_client.Reply(message.Channel, message.ID, twitchMessage)
			} else {
				twitchMessage := fmt.Sprintf("[Crew Division Rank]: %d [Player Crew Score]: %s [Total Crew Score]: %s", stats.CrewDivisionRank, stats.PlayerCrewScore, stats.CrewTotalScore)
				irc_client.Reply(message.Channel, message.ID, twitchMessage)
			}
		}

		if strings.Contains(message.Message, "!rank") {
			var playerId string
			playerIdReq := getPlayerIdFromChannel(message.Channel, twitch_client)
			if playerIdReq != "ERROR" && playerIdReq != "" {
				playerId = playerIdReq
			} else {
				messageReply := fmt.Sprintf("@%s must link their Spectre Divide account to Twitch!", message.Channel)
				irc_client.Reply(message.Channel, message.ID, messageReply)
				return
			}

			respUrl := "https://collective-production.up.railway.app/getPlayerRankData/" + playerId
			resp, err := http.Get(respUrl)
			if err != nil {
				log.Fatalf("Couldn't get player rank data for playerId: %s", playerId)
			}

			defer resp.Body.Close()
			playerRankBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("Issue while reading body of resp: %s", err)
				return
			}
			playerRankData := PlayerRank{}

			playerRankDataUnmarshalErr := json.Unmarshal(playerRankBody, &playerRankData)
			if playerRankDataUnmarshalErr != nil {
				log.Fatalf("Issue while unmarshalling json response from player rank data endpoint. %s", err)
			}

			twitchMessage := fmt.Sprintf("[Current Solo Rank]: %s  [Highest Team Rank]: %s", getSoloRankFromRankNumber(playerRankData.SoloRank), getTeamRankFromRankNumber(playerRankData.TeamRank))
			irc_client.Reply(message.Channel, message.ID, twitchMessage)
			return
		}

		if strings.Contains(message.Message, "!myrank") {
			var playerId string
			playerIdReq := getPlayerIdFromTwitchId(message.User.ID)

			if playerIdReq != "ERROR" && playerIdReq != "" {
				playerId = playerIdReq
			} else {
				irc_client.Reply(message.Channel, message.ID, "You must link your Spectre Divide account to Twitch!")
				return
			}

			respUrl := "https://collective-production.up.railway.app/getPlayerRankData/" + playerId
			resp, err := http.Get(respUrl)
			if err != nil {
				log.Fatalf("Couldn't get player rank data for playerId: %s", playerId)
			}

			defer resp.Body.Close()
			playerRankBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("Issue while reading body of resp: %s", err)
				return
			}
			playerRankData := PlayerRank{}

			playerRankDataUnmarshalErr := json.Unmarshal(playerRankBody, &playerRankData)
			if playerRankDataUnmarshalErr != nil {
				log.Fatalf("Issue while unmarshalling json response from player rank data endpoint. %s", err)
			}

			twitchMessage := fmt.Sprintf("[Current Solo Rank]: %s  [Highest Team Rank]: %s", getSoloRankFromRankNumber(playerRankData.SoloRank), getTeamRankFromRankNumber(playerRankData.TeamRank))
			irc_client.Reply(message.Channel, message.ID, twitchMessage)
			return
		}

		if strings.Contains(message.Message, "!spectrestats") || strings.Contains(message.Message, "!santaigg") {
			twitchMessage := fmt.Sprintf("Use !crewstats to get %s's crew stats & !mycrewstats to get your own crew stats. !rank for %s's rank, and !myrank for yours.", message.Channel, message.Channel)
			irc_client.Reply(message.Channel, message.ID, twitchMessage)
		}
	})

	irc_client.Join("truo", "limitediq__", "staycationtg", "ethos", "bugzvii", "steazecs", "moepork", "relyks", "shroud", "iitztimmy", "pieman", "shrood", "omegatooyew", "bixle", "equustv")

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Patches!")
	})

	app.Get("/joinTwitch", func(c *fiber.Ctx) error {
		channel := new(TwitchChannel)
		if err := c.QueryParser(channel); err != nil {
			return err
		}

		irc_client.Join(channel.Channel)

		return c.SendString("Joined " + channel.Channel)
	})

	go func() { log.Fatal(app.Listen(":3002")) }()

	err = irc_client.Connect()
	if err != nil {
		panic(err)
	}
}

func hasTimePassed(lastSet time.Time, duration time.Duration) bool {
	return time.Since(lastSet) >= duration
}

func getSoloRankFromRankNumber(rankNumber int) string {
	soloRanks := map[int]string{
		0:  "Unranked",
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

	return soloRanks[rankNumber]
}

func getTeamRankFromRankNumber(rankNumber int) string {
	teamRanks := map[int]string{
		0:  "Unranked",
		1:  "Undiscovered 1",
		2:  "Undiscovered 2",
		3:  "Undiscovered 3",
		4:  "Undiscovered 4",
		5:  "Prospect 1",
		6:  "Prospect 2",
		7:  "Prospect 3",
		8:  "Prospect 4",
		9:  "Talent 1",
		10: "Talent 2",
		11: "Talent 3",
		12: "Talent 4",
		13: "Professional 1",
		14: "Professional 2",
		15: "Professional 3",
		16: "Professional 4",
		17: "Elite 1",
		18: "Elite 2",
		19: "Elite 3",
		20: "Elite 4",
		21: "International 1",
		22: "International 2",
		23: "International 3",
		24: "International 4",
		25: "Superstar 1",
		26: "Superstar 2",
		27: "Superstar 3",
		28: "Superstar 4",
		29: "World Class 1",
		30: "World Class 2",
		31: "World Class 3",
		32: "World Class 4",
		33: "Champion"}

	return teamRanks[rankNumber]
}

func getPlayerIdFromChannel(channel string, client *helix.Client) string {
	var playerId string
	if channel == "ethos" {
		playerId = "E27C1FD1-4EEB-483D-952D-A7C904869509"
	}
	if channel == "truo" {
		playerId = "8D02F2C0-69B8-4CEE-9656-2D0866B44E9B"
	}
	if channel == "staycationtg" {
		playerId = "F0CD9516-6DFB-4235-8E04-32D6B820754C"
	}
	if channel == "bugzvii" {
		playerId = "30C3E8E8-B5A4-4461-B77C-567B9B3C762D"
	}
	if channel == "steazecs" {
		playerId = "BCD9F729-DA28-4802-8CF6-DE831B852D62"
	}
	if channel == "moepork" {
		playerId = "39F848C1-A9A5-42DF-81AA-033191455DAA"
	}
	if channel == "relyks" {
		playerId = "DC5D1993-5B94-4F0C-8F57-DB51B0DAE7F1"
	}
	if channel == "shroud" {
		playerId = "CE4C88F7-7D66-417F-A3F5-01D0F9F52B90"
	}
	if channel == "iitztimmy" {
		playerId = "1d36bff3-1ac5-422f-bb21-f6524e0b83a0"
	}
	if channel == "pieman" {
		playerId = "a666813a-5cc1-48ac-bcdf-ac937bda38bf"
	}
	if channel == "omegatooyew" {
		playerId = "8edc5a72-933c-412a-af09-51f587099e89"
	}
	if channel == "shrood" {
		playerId = "CE4C88F7-7D66-417F-A3F5-01D0F9F52B90"
	}
	if channel == "bixle" {
		playerId = "acc53a25-d944-4853-bcbf-f03526885008"
	}

	if channel == "just9n" {
		playerId = "1dbe12d4-b170-4d1b-bea0-63be2b72f410"
	}

	if playerId != "" {
		return playerId
	}

	channelId := getTwitchIdFromChannel(channel, client)
	if channelId == "" {
		return ""
	}

	userUrl := "https://collective-production.up.railway.app/getPlayerIdentityFromTwitchId/" + channelId
	playerIdReq, playerIdReqErr := http.Get(userUrl)
	if playerIdReqErr != nil {
		log.Fatalf("Couldn't get player-id from twitch-id for: %s", channelId)
		return ""
	}
	defer playerIdReq.Body.Close()
	// read body
	playerIdReqBody, err := io.ReadAll(playerIdReq.Body)
	if err != nil {
		log.Fatalf("impossible to read all body of response: %s", err)
		return ""
	}

	playerIdentity := GetPlayerIdentityFromTwitchId{}
	playerIdReqUnmarshalErr := json.Unmarshal(playerIdReqBody, &playerIdentity)
	if playerIdReqUnmarshalErr != nil {
		log.Fatalf("Error while unmarshaling getPlayerIdentityFromTwitchId: %s", playerIdReqUnmarshalErr)
		return ""
	}

	return playerIdentity.PlayerId
}

func getPlayerIdFromTwitchId(userId string) string {
	userUrl := "https://collective-production.up.railway.app/getPlayerIdentityFromTwitchId/" + userId
	playerIdReq, playerIdReqErr := http.Get(userUrl)
	if playerIdReqErr != nil {
		log.Fatalf("Couldn't get player-id from twitch-id for: %s", userId)
		return ""
	}
	defer playerIdReq.Body.Close()
	// read body
	playerIdReqBody, err := io.ReadAll(playerIdReq.Body)
	if err != nil {
		log.Fatalf("impossible to read all body of response: %s", err)
		return ""
	}

	playerIdentity := GetPlayerIdentityFromTwitchId{}
	playerIdReqUnmarshalErr := json.Unmarshal(playerIdReqBody, &playerIdentity)
	if playerIdReqUnmarshalErr != nil {
		log.Fatalf("Error while unmarshaling getPlayerIdentityFromTwitchId: %s", playerIdReqUnmarshalErr)
		return ""
	}

	return playerIdentity.PlayerId
}

func getTwitchIdFromChannel(channel string, client *helix.Client) string {
	resp, err := client.GetUsers(&helix.UsersParams{
		Logins: []string{channel},
	})

	if err != nil {
		log.Fatalf("Issue getting twitch id from channel... %s", err)
	}

	if len(resp.Data.Users) > 0 {
		return resp.Data.Users[0].ID
	}
	return ""
}
