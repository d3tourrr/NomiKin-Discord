package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "regexp"
    "strings"
    "time"

    "github.com/bwmarrin/discordgo"
)

func (companion *Companion) AmIPrimary(m *discordgo.MessageCreate) bool {
    sendPrimary := true
    if RoomPrimaries[m.ChannelID] != companion.CompanionId {
        // We're not the primary
        VerboseLog("%v is not primary for %v - %v is", companion.CompanionId, m.ChannelID, RoomPrimaries[m.ChannelID])
        sendPrimary = false
    } else {
        VerboseLog("%v is primary for %v", companion.CompanionId, m.ChannelID)
    }
    return sendPrimary
}

func (c *Companion) GetRoomMembers(roomId string) []string {
    roomInfo, err := c.NomiKin.RoomExists(&roomId)
    if err != nil {
        log.Printf("Error checking if room exists: %v\n", err)
        return []string{}
    }

    if roomInfo == nil || roomInfo.Nomis == nil {
        log.Printf("Room info or room members are nil for room %s\n", roomId)
        return []string{}
    }

    var retMembers []string
    for _, m := range roomInfo.Nomis {
        if m.Uuid != "" {
            retMembers = append(retMembers, m.Uuid)
        }
    }

    VerboseLog("Room members for %v are: %v", roomId, retMembers)
    return retMembers
}

func (c *Companion) CheckRoomStatus(roomId string) string {
    endpoint := "https://api.nomi.ai/v1/rooms/" + roomId
    repl, err := c.NomiKin.ApiCall(endpoint, "GET", nil)
    if err != nil {
        log.Printf("Error calling Nomi %v API: %v\n", c.CompanionId, err)
    }

    var resp map[string]interface{}
    err = json.Unmarshal(repl, &resp)
    if err != nil {
        log.Printf("%v failed to decode JSON: %v\n", c.CompanionId, err)
    }

    status, ok := resp["status"].(string)
    if !ok {
        log.Printf("%v status field is missing\n", c.CompanionId)
    }

    return status
}

func (c *Companion) WaitForRoom(roomId string) bool {
    waitFor := 45
    waited := 0
    for {
        if waited > waitFor {
            VerboseLog("companion.WaitForRoom took longer than 45 seconds - Nomi: %v - Room: %v", c.CompanionId, roomId)
            return false
        }
        if c.CheckRoomStatus(roomId) == "Default" {
            return true
        }
        time.Sleep(time.Second * 1)
        waited++
    }
}

func (companion *Companion) ResponseNeeded(m *discordgo.MessageCreate) (bool, string) {
    respondToThis := false
    verboseReason := ""

    // Is the companion mentioned or is this a reply to their message?
    if companion.RespondPing {
        for _, user := range m.Mentions {
            if user.ID == companion.DiscordSession.State.User.ID {
                respondToThis = true
                verboseReason = "Pinged/RepliedTo"
                break
            }
        }
    }

    // Is the companion mentioned by role
    // Doesn't work in DMs, no need to check if the bot is also mentioned specifically
    if companion.RespondRole && !respondToThis && m.GuildID != "" {
        // Check this every time in case the bot is added to or removed from roles, not in DMs
        botID := companion.DiscordSession.State.User.ID

        botMember, err := companion.DiscordSession.GuildMember(m.GuildID, botID)
        if err != nil {
            log.Printf("Error retrieving bot member: %v\n", err)
            return false, verboseReason
        }

        for _, roleID := range botMember.Roles {
            roleMention := fmt.Sprintf("<@&%s>", roleID)
            if strings.Contains(m.Content, roleMention) {
                respondToThis = true
                verboseReason = "RolePing"
                break
            }
        }
    }

    // Does this message contain one of the reponse keywords?
    if companion.Keywords != "" && !respondToThis {
        responseKeywords := companion.Keywords
        words := strings.Split(responseKeywords, ",")
        charCleaner := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
        messageWords := strings.Fields(strings.ToLower(charCleaner.ReplaceAllString(m.Message.Content, "")))

        for _, word := range words {
            trimmedWord := strings.ToLower(strings.TrimSpace(word))
            for _, messageWord := range messageWords {
                if trimmedWord == messageWord {
                    respondToThis = true
                    verboseReason = "Keyword:" + trimmedWord
                    break
                }
            }
            if respondToThis {
                break
            }
        }
    }

    // Is this a DM?
    if m.GuildID == "" {
        // If this is a DM, respond if RESPOND_TO_DIRECT_MESSAGE is true, ignore if it's false,
        switch companion.RespondDM {
            case true:
            respondToThis = true
            verboseReason = "DM"
            case false:
            respondToThis = false
        }
    }

    // Is this a Nomi Room? Random chance to respond
    if companion.CompanionType == "NOMI" && companion.ChatStyle == "ROOMS" && !respondToThis && m.GuildID != "" {
        rand.Seed(time.Now().UnixNano())
        randomValue := rand.Float64() * 100
        if randomValue < float64(companion.RoomObjects[m.ChannelID].RandomResponseChance) {
            respondToThis = true
            verboseReason = "RandomResponseChance"
            VerboseLog("Nomi %v random response chance triggered. RandomResponseChance in channel %v set to %v.", companion.CompanionId, m.ChannelID, float64(companion.RoomObjects[m.ChannelID].RandomResponseChance))
        }
    }

    return respondToThis, verboseReason
}

