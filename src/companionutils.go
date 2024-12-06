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
    "github.com/d3tourrr/NomiKinGo"
)

func (companion *Companion) AmIPrimary(m *discordgo.MessageCreate) bool {
    sendPrimary := true
    if RoomPrimaries[m.ChannelID] != companion.CompanionId {
        // We're not the primary
        companion.VerboseLog("Not primary for %v - %v is", m.ChannelID, RoomPrimaries[m.ChannelID])
        sendPrimary = false
    } else {
        companion.VerboseLog("Is primary for %v", m.ChannelID)
    }
    return sendPrimary
}

func (c *Companion) GetRoomMembers(roomId string) []string {
    var roomInfo *NomiKin.Room
    callOut := SuppressGetRoomLogs(c.NomiKin.RoomExists, &roomId)

    if callOut[0] != nil {
        roomInfo = callOut[0].(*NomiKin.Room)
    }
    if callOut[1] != nil {
        c.Log("Error checking if room exists: %v", callOut[1].(error))
        return []string{}
    }

    if roomInfo == nil || roomInfo.Nomis == nil {
        c.Log("Room info or room members are nil for room %v", roomId)
        return []string{}
    }

    var retMembers []string
    for _, m := range roomInfo.Nomis {
        if m.Uuid != "" {
            retMembers = append(retMembers, m.Uuid)
        }
    }

    c.VerboseLog("Room members for %v are: %v", roomId, retMembers)
    return retMembers
}

func (c *Companion) CheckRoomStatus(roomId string) string {
    endpoint := "https://api.nomi.ai/v1/rooms/" + roomId
    repl, err := c.NomiKin.ApiCall(endpoint, "GET", nil)
    if err != nil {
        c.Log("Error calling Nomi %v API: %v", err)
    }

    var resp map[string]interface{}
    err = json.Unmarshal(repl, &resp)
    if err != nil {
        c.Log("Failed to decode JSON: %v", err)
    }

    status, ok := resp["status"].(string)
    if !ok {
        c.Log("Status field is missing for Room: %v", roomId)
    }

    return status
}

func (c *Companion) WaitForRoom(roomId string) bool {
    waitFor := 45
    waited := 0
    for {
        if waited > waitFor {
            c.VerboseLog("companion.WaitForRoom took longer than 45 seconds - Room: %v", roomId)
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
            companion.Log("Error retrieving bot member: %v", err)
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
            companion.VerboseLog("Random response chance triggered. RandomResponseChance in channel %v set to %v.", m.ChannelID, float64(companion.RoomObjects[m.ChannelID].RandomResponseChance))
        }
    }

    return respondToThis, verboseReason
}

func (c *Companion) VerboseLog(s string, args ...interface{}) {
    if Verbose {
        if c == nil {
            log.Printf("Attempted to call VerboseLog on a nil Companion instance. String: %v\n", s)
            return
        }

        if len(args) > 0 {
            c.Log(s, args...)
        } else {
            c.Log(s)
        }
    }
}

func (c *Companion) Log(s string, args ...interface{}) {
    if c == nil {
        log.Printf("Attempted to call Log on a nil Companion instance. String: %v\n", s)
        return
    }

    parts := strings.Split(s, "%v")
    if len(parts)-1 != len(args) {
        log.Printf("[%v] Logging Error: Number of format specifiers does not match number of arguments. String: %v\n", c.CompanionId, s)
    }

    var sb strings.Builder
    for i, part := range parts {
        sb.WriteString(part)
        if i < len(args) {
            sb.WriteString(fmt.Sprintf("%v", args[i]))
        }
    }

    cBit := "[" + c.CompanionId + "]"
    cPad := LogWidth - len(cBit) + 3
    if cPad < 0 {
        cPad = 0
    }
    log.Printf("%v%v %v", strings.Repeat(" ", cPad), cBit, sb.String())
}

