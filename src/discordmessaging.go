package main

import (
    "log"
    "regexp"
    "math/rand"
    "strings"
    "time"

    "github.com/bwmarrin/discordgo"
)

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

func SendMessageToCompanion(m *discordgo.MessageCreate, companion *Companion) error {
    respondToThis, verboseReason := companion.ResponseNeeded(m)

    if respondToThis {
        VerboseLog("%v - Response required: %v", companion.CompanionId, verboseReason)
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
                // The `status` field on a Room doesn't update quickly enough when more than one Nomi is responding
                // So I'm doing this contrived nonsense to randomize a delay per Nomi, so they hopefully don't overlap
                rand.Seed(time.Now().UnixNano())
                maxWait := len(companion.GetRoomMembers(m.ChannelID))
                randWait := rand.Intn(maxWait) + rand.Intn(4)
                VerboseLog("%v - Sleeping %v seconds to avoid Nomi Room collision", companion.CompanionId, randWait)
                time.Sleep(time.Second * time.Duration(randWait))

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
            log.Printf("Error exchanging messages with companion: %v\n", err)
            stopTyping <- true
            return nil
        }

        stopTyping <- true

        // Send as a reply to the message that triggered the response, helps keep things orderly
        // But only if this is in a server - if it's a DM, send it as a straight message
        if m.GuildID == "" {
            _, sendErr := companion.DiscordSession.ChannelMessageSend(m.ChannelID, companionReply)
            if sendErr != nil {
                log.Printf("Error sending message: %v\n", err)
            }
        } else {
            _, sendErr := companion.DiscordSession.ChannelMessageSendReply(m.ChannelID, companionReply, m.Reference())
            if sendErr != nil {
                log.Printf("Error sending message: %v\n", err)
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

func (companion *Companion) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == companion.DiscordSession.State.User.ID {
        // We don't have to send our companion their own messages
        return
    }

    if companion.CompanionType == "NOMI" && companion.ChatStyle == "ROOMS" {
        // If we're in Rooms mode, drop messages for which we don't have a room setup for
        if companion.RoomObjects[m.ChannelID].Uuid == "" {
            return
        }
        
        // If the message is from a Nomi in the same room, we don't have to process it, because they already did
        if m.Author.Bot {
            sharedNomiRoom := false
            roomMems := companion.GetRoomMembers(m.ChannelID)
            for b, c := range Companions {
                if b.State.User.ID == m.Author.ID && Contains(roomMems, c.CompanionId) {
                    sharedNomiRoom = true
                    VerboseLog("%v message from Companion %v is in the same room %v - not forwarding to the room", companion.CompanionId, c.CompanionId, m.ChannelID)
                    break
                }
            }

            if sharedNomiRoom {
                return
            }
        }
    }

    message := QueuedMessage{
        Message: m,
        Companion: companion,
    }

    companion.Queue.Enqueue(message)
}

