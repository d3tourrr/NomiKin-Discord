package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "math/rand"
    "os"
    "os/signal"
    "net/http"
    "reflect"
    "regexp"
    "strings"
    "syscall"
    "time"

    "github.com/bwmarrin/discordgo"
)

func PrintStructFields(c *Companion) {
    val := reflect.ValueOf(c).Elem()
    typ := reflect.TypeOf(c).Elem()

    fmt.Printf("Companion: %v\n", c.CompanionId)
    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        fieldName := typ.Field(i).Name
        fieldValue := field.Interface()

        if fieldName == "DiscordBotToken" || fieldName == "CompanionToken" {
            // Keep secrets a secret
            fieldValue = "***redacted***"
        }

        fmt.Printf("  %s: %v\n", fieldName, fieldValue)
    }
}

func VerboseLog(s string, args ...interface{}) {
    if Verbose {
        parts := strings.Split(s, "%v")
        if len(parts)-1 != len(args) {
            log.Println("Verbose Logging: Number of format specifiers does not match number of arguments.")
        }

        var sb strings.Builder
        for i, part := range parts {
            sb.WriteString(part)
            if i < len(args) {
                sb.WriteString(fmt.Sprintf("%v", args[i]))
            }
        }

        log.Println(sb.String())
    }
}

func GetEnvFiles(dir string) ([]string, error) {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    var envFiles []string
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".env") {
            envFiles = append(envFiles, dir + "/" + file.Name())
        }
    }

    return envFiles, nil
}

func UpdateMessage(m *discordgo.MessageCreate, companion *Companion) string {
    updatedMessage := m.Content

    var err error
    if m.GuildID != "" {
        // Only if it's not a DM, otherwise this doesn't work - apparently this needs guild state info
        updatedMessage, err = m.ContentWithMoreMentionsReplaced(companion.DiscordSession)
        if err != nil {
            log.Printf("Error replacing Discord mentions with usernames: %v", err)
        }
    }

    // Do we need the normal or reply prefix?
    reply := false
    userPrefix := ""
    if m.MessageReference != nil && m.MessageReference.MessageID != "" {
        reply = true
    }

    if reply {
        if companion.ReplyPrefix == "" {
            userPrefix = companion.MessagePrefix
        } else {
            userPrefix = companion.ReplyPrefix
        }

        repliedMessage, err := companion.DiscordSession.ChannelMessage(m.ChannelID, m.MessageReference.MessageID)
        if err != nil {
            log.Printf("Error fetching replied message: %v\n", err)
        }

        reU := regexp.MustCompile(`\{\{USERNAME\}\}`)
        reR := regexp.MustCompile(`\{\{REPLY_TO\}\}`)
        userPrefix = reU.ReplaceAllString(userPrefix, m.Author.Username)
        userPrefix = reR.ReplaceAllString(userPrefix, repliedMessage.Author.Username)
    } else {
        userPrefix = companion.MessagePrefix
        re := regexp.MustCompile(`\{\{USERNAME\}\}`)
        userPrefix = re.ReplaceAllString(userPrefix, m.Author.Username)
    }

    updatedMessage = userPrefix + " " + updatedMessage
    updatedMessage = strings.TrimSpace(updatedMessage)

    return updatedMessage
}

func Contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func SendMessageToCompanion(m *discordgo.MessageCreate, companion *Companion) error {
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
            return fmt.Errorf("Error retrieving bot member: %v", err)
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

    if respondToThis {
        VerboseLog("Response required: %v", verboseReason)
        loopBreak := false
        if m.Author.Bot {
            reply := companion.Tracker.TrackMessage(m.Author.ID, companion)
            if !reply {
                // We've passed our threshold for messages from this bot within the last hour
                loopBreak = true
            }
        }

        if loopBreak && !(companion.CompanionType == "NOMI" && companion.ChatStyle == "ROOMS") {
            // Breaking the loop and don't worry about sending a message to the Nomi Room
            VerboseLog("%v Loop Prevention active [type: %v | mode: %v], halting reply chain.", companion.CompanionId, companion.CompanionType, companion.ChatStyle)
            return nil
        }

        updatedMessage := UpdateMessage(m, companion)

        var err error
        companionReply := ""
        err = nil

        // set the typing indicator
        stopTyping := make(chan bool)
        if !loopBreak {
            go func() {
                for {
                    select {
                    case <-stopTyping:
                        return
                    default:
                        companion.DiscordSession.ChannelTyping(m.ChannelID)
                        time.Sleep(5 * time.Second) // Adjust the interval as needed
                    }
                }
            }()
        }

        switch companion.CompanionType {
        case "NOMI":
            if companion.ChatStyle == "ROOMS" {
                NomiRoomSend(companion, m)
                if !loopBreak {
                    canSend := companion.WaitForRoom(companion.RoomObjects[m.ChannelID].Uuid)
                    if canSend {
                        roomId := companion.RoomObjects[m.ChannelID].Uuid
                        companionReply, err = companion.NomiKin.RequestNomiRoomReply(&roomId, &companion.NomiKin.CompanionId)
                    } else {
                        log.Printf("Error: Waited as long as we could, but room %v was never ready for a message from %v\n", m.ChannelID, companion.CompanionId)
                    }
                } else {
                    VerboseLog("%v Loop Prevention active [type: %v | mode: %v], halting reply chain.", companion.CompanionId, companion.CompanionType, companion.ChatStyle)
                }
            } else {
                companionReply, err = companion.NomiKin.SendNomiMessage(&updatedMessage)
            }
        case "KINDROID":
            companionReply, err = companion.NomiKin.SendKindroidMessage(&updatedMessage)
        }
        if err != nil {
            fmt.Printf("Error exchanging messages with companion: %v", err)
            stopTyping <- true
            return nil
        }

        stopTyping <- true

        // Send as a reply to the message that triggered the response, helps keep things orderly
        // But only if this is in a server - if it's a DM, send it as a straight message
        if m.GuildID == "" {
            _, sendErr := companion.DiscordSession.ChannelMessageSend(m.ChannelID, companionReply)
            if sendErr != nil {
                fmt.Println("Error sending message: ", err)
            }
        } else {
            _, sendErr := companion.DiscordSession.ChannelMessageSendReply(m.ChannelID, companionReply, m.Reference())
            if sendErr != nil {
                fmt.Println("Error sending message: ", err)
            }
        }

        return nil
    }

    // Even if a Nomi won't respond, if they are in ROOMS mode, we need to send the message to the correct room
    if companion.CompanionType == "NOMI" && companion.ChatStyle == "ROOMS" {
        NomiRoomSend(companion, m)
    }

    return nil
}

func NomiRoomSend(companion *Companion, m *discordgo.MessageCreate) {
    updatedMessage := UpdateMessage(m, companion)
    sendPrimary := companion.AmIPrimary(m)
    roomId := companion.RoomObjects[m.ChannelID].Uuid
    if sendPrimary {
        canSend := companion.WaitForRoom(companion.RoomObjects[m.ChannelID].Uuid)
        if canSend {
            _, err := companion.NomiKin.SendNomiRoomMessage(&updatedMessage, &roomId)
            if err != nil {
                log.Printf("Error sending message to room: %v\n", err)
            }
        } else {
            log.Printf("Error: Waited as long as we could, but room %v was never ready for a message from %v\n", m.ChannelID, companion.CompanionId)
        }
    }
}

func WaitForShutdown(bots []*discordgo.Session) {
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    <-stop
    for _, bot := range bots {
        bot.Close()
    }
    log.Println("All bots stopped.")
}

func UpdateStatus(dg *discordgo.Session) {
    statusMessageLocation := "https://raw.githubusercontent.com/d3tourrr/NomiKin-Discord/refs/heads/main/StatusMessage.txt"
    statusResp, err := http.Get(statusMessageLocation)
    if err != nil {
        log.Printf("Error retrieving status message: %v", err)
    }
    defer statusResp.Body.Close()

    statusMessageContent, err := ioutil.ReadAll(statusResp.Body)
    if err != nil {
        log.Printf("Error reading status message: %v", err)
    }

    discordStatus := Version + " " + string(statusMessageContent)
    err = dg.UpdateStatusComplex(discordgo.UpdateStatusData{
        Status: "online",
        Activities: []*discordgo.Activity{
            {
                Name: discordStatus,
                Type: discordgo.ActivityTypeGame,
            },
        },
    })
    if err != nil {
        log.Printf("Error setting status: %v", err)
    }
}

