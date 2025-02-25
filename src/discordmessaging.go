package main

import (
    "encoding/base64"
    "regexp"
    "math/rand"
    "net/url"
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
            companion.Log("Error replacing Discord mentions with usernames: %v", err)
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
            companion.Log("Error fetching replied message: %v", err)
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

func SendMessageToCompanion(m *discordgo.MessageCreate, companion *Companion, botRespondNoForward bool) error {
    loopBreak := false
    respondToThis, verboseReason := companion.ResponseNeeded(m)

    if respondToThis {
        companion.VerboseLog("Response required: %v", verboseReason)
        if m.Author.Bot {
            loopBreak = !companion.Tracker.TrackMessage(m.Author.ID, companion)
        }

        if loopBreak && !(companion.ChatStyle == "ROOMS") {
            // Breaking the loop and don't worry about sending a message
            companion.VerboseLog("Loop Prevention active [type: %v | mode: %v], halting reply chain.", companion.CompanionType, companion.ChatStyle)
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
                if !botRespondNoForward {
                    NomiRoomSend(companion, m)
                }

                if !loopBreak {
                    canSend := companion.WaitForRoom(companion.NomiRoomObjects[m.ChannelID].Uuid)
                    if canSend {
                        roomId := companion.NomiRoomObjects[m.ChannelID].Uuid
                        companionReply, err = companion.NomiKin.RequestNomiRoomReply(&roomId, &companion.NomiKin.CompanionId)
                    } else {
                        companion.Log("Error: Waited as long as we could, but room %v was never ready for a message", m.ChannelID)
                    }
                } else {
                    companion.VerboseLog("Loop Prevention active [type: %v | mode: %v], halting reply chain.", companion.CompanionType, companion.ChatStyle)
                    return nil
                }
            } else {
                companionReply, err = companion.NomiKin.SendNomiMessage(&updatedMessage)
            }
        case "KINDROID":
            if companion.ChatStyle == "ROOMS" {
                req := ""
                encoded := url.QueryEscape(m.Author.Username)
                base64Str := base64.StdEncoding.EncodeToString([]byte(encoded))
                alphanumeric := strings.Map(func(r rune) rune {
                    if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
                        return r
                    }
                    return -1
                }, base64Str)
                if len(alphanumeric) > 32 {
                    req = alphanumeric[:32]
                } else {
                    req = alphanumeric
                }

                // TODO: Come back and make this something you can set
                nsfwFilter := bool(false)

                conversation, err := companion.GetConversation(m)
                
                companion.VerboseLog("Sending: ShareID: %v | Req: %v | Conversation Count: %v | Last Msg: ", companion.KinShareId, &req, len(*conversation), (*conversation)[0].Text)
                companionReply, err = companion.NomiKin.SendKindroidDiscordBot(&companion.KinShareId, &nsfwFilter, &req, *conversation)
                if err != nil {
                    companion.Log("Error sending message to Kindroid: %v", err)
                }
            } else {
                companionReply, err = companion.NomiKin.SendKindroidMessage(&updatedMessage)
            }
        }
        if err != nil {
            companion.Log("Error exchanging messages with companion: %v", err)
            stopTyping <- true
            return nil
        }

        stopTyping <- true

        // Send as a reply to the message that triggered the response, helps keep things orderly
        // But only if this is in a server - if it's a DM, send it as a straight message
        if m.GuildID == "" {
            _, sendErr := companion.DiscordSession.ChannelMessageSend(m.ChannelID, companionReply)
            if sendErr != nil {
                companion.Log("Error sending message: %v", err)
            }
        } else {
            _, sendErr := companion.DiscordSession.ChannelMessageSendReply(m.ChannelID, companionReply, m.Reference())
            if sendErr != nil {
                companion.Log("Error sending message: %v", err)
            }
        }

        // Parse out emojis to use as reactions to the original message
        if companion.EmojisToReact {
            companion.VerboseLog("EmojisToReact enabled. Adding reactions to message from %v", m.Author.ID)
            eligibleEmojis := companion.GetEligibleEmojis(companionReply)
            randEmojis := make([]string, companion.MaxReactions)

            if len(eligibleEmojis) > 0 {
                companion.VerboseLog("Eligible Emojis: %v", eligibleEmojis)
                if companion.MaxReactions >= len(eligibleEmojis) {
                    randEmojis = eligibleEmojis
                } else {
                    for i := len(eligibleEmojis) - 1; i > 0; i-- {
                        j := rand.Intn(i + 1)
                        eligibleEmojis[i], eligibleEmojis[j] = eligibleEmojis[j], eligibleEmojis[i]
                    }

                    randEmojis = eligibleEmojis[:companion.MaxReactions]
                    companion.VerboseLog("RandEmojis: %v (%v/%v)", randEmojis, len(randEmojis), len(eligibleEmojis))
                }

                companion.Log("Adding reactions to message from %v: %v", m.Author.ID, randEmojis)
                for _, emoji := range randEmojis {
                    err := companion.DiscordSession.MessageReactionAdd(m.ChannelID, m.ID, emoji)
                    if err != nil {
                        companion.Log("Error adding reaction to message: %v", err)
                    }
                }
            } else {
                companion.VerboseLog("No eligible emojis found")
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
    roomId := companion.NomiRoomObjects[m.ChannelID].Uuid
    if sendPrimary {
        // The `status` field on a Room doesn't update quickly enough when more than one Nomi is responding
        // So I'm doing this contrived nonsense to randomize a delay per Nomi, so they hopefully don't overlap
        maxWait := len(companion.GetRoomMembers(m.ChannelID))
        randWait := rand.Intn(maxWait) + rand.Intn(4)
        companion.VerboseLog("Sleeping %v seconds to avoid Nomi Room collision", randWait)
        time.Sleep(time.Second * time.Duration(randWait))

        canSend := companion.WaitForRoom(companion.NomiRoomObjects[m.ChannelID].Uuid)
        if canSend {
            _, err := companion.NomiKin.SendNomiRoomMessage(&updatedMessage, &roomId)
            if err != nil {
                companion.Log("Error sending message to room: %v", err)
            }
        } else {
            companion.Log("Error: Waited as long as we could, but room %v was never ready for a message", m.ChannelID)
        }
    }
}

func (companion *Companion) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // For when we need to respond to another bot without forwarding the message to the room
    botRespondNoForward := false

    if m.Author.ID == companion.DiscordSession.State.User.ID {
        // We don't have to send our companion their own messages
        return
    }

    // Ignore messages with embeds
    if len(m.Embeds) > 0 {
        companion.VerboseLog("Dropping message from %v because it has embeds", m.Author.ID)
        return
    }

    if companion.CompanionType == "NOMI" && companion.ChatStyle == "ROOMS" {
        // If we're in Rooms mode, drop messages for which we don't have a room setup for
        if companion.NomiRoomObjects[m.ChannelID].Uuid == "" {
            return
        }
        
        // If the message is from a Nomi in the same room, we don't have to process it, because they already did
        if m.Author.Bot {
            roomMems := companion.GetRoomMembers(m.ChannelID)
            for b, c := range Companions {
                if b.State.User.ID == m.Author.ID && Contains(roomMems, c.CompanionId) {
                    botRespondNoForward = true
                    companion.VerboseLog("Message from Companion %v is in the same room %v - not forwarding to the room", c.CompanionId, m.ChannelID)
                    break
                }
            }
        }
    }

    message := QueuedMessage{
        Message: m,
        Companion: companion,
        BotRespondNoForward: botRespondNoForward,
    }

    companion.Queue.Enqueue(message)
}

