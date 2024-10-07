
package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "regexp"
    "strings"
    "sync"
    "time"

    "github.com/bwmarrin/discordgo"
    "github.com/joho/godotenv"
    NomiKin "github.com/d3tourrr/NomiKinGo"
)

var version = "v0.3"

type QueuedMessage struct {
    Message   *discordgo.MessageCreate
    Session   *discordgo.Session
}

type MessageQueue struct {
    messages []QueuedMessage
    mu       sync.Mutex
}

func (q *MessageQueue) Enqueue(message QueuedMessage) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.messages = append(q.messages, message)
}

func (q *MessageQueue) Dequeue() (QueuedMessage, bool) {
    q.mu.Lock()
    defer q.mu.Unlock()

    if len(q.messages) == 0 {
        return QueuedMessage{}, false
    }

    message := q.messages[0]
    q.messages = q.messages[1:]
    return message, true
}

func (q *MessageQueue) ProcessMessages() {
    for {
        queuedMessage, ok := q.Dequeue()
        if !ok {
            time.Sleep(1 * time.Second) // No messages in queue, sleep for a while
            continue
        }

        err := sendMessageToAPI(queuedMessage.Session, queuedMessage.Message)
        if err != nil {
            log.Printf("Failed to send message to Companion API: %v", err)
            q.Enqueue(queuedMessage) // Requeue the message if failed
        }

        time.Sleep(5 * time.Second) // Try to keep from sending messages toooo quickly
    }
}

func sendMessageToAPI(s *discordgo.Session, m *discordgo.MessageCreate) error {
    // Check if the message mentions the bot
    for _, user := range m.Mentions {
        if user.ID == s.State.User.ID {
            companionToken := os.Getenv("COMPANION_TOKEN")
            if companionToken == "" {
                fmt.Println("No companion token provided. Set COMPANION_TOKEN environment variable.")
                return nil
            }

            companionId := os.Getenv("COMPANION_ID")
            if companionId == "" {
                fmt.Println("No companion AI ID provided. Set COMPANION_ID environment variable.")
                return nil
            }

            companionType := os.Getenv("COMPANION_TYPE")
            if companionType == "" {
                fmt.Println("No companion AI type provided. Set COMPANION_TYPE environment variable.")
                return nil
            }

            companionType = strings.ToUpper(companionType)
            if companionType != "NOMI" && companionType != "KINDROID" {
                fmt.Printf("Improper companion type. Set COMPANION_TYPE environment variable to either 'NOMI' or 'KINDROID'. Your value: %v", companionType)
                return nil
            }

            // Replacing mentions makes it so the companion sees the usernames instead of <@userID> syntax
            updatedMessage, err := m.ContentWithMoreMentionsReplaced(s)
            if err != nil {
                log.Printf("Error replacing Discord mentions with usernames: %v", err)
            }

            userPrefix := os.Getenv("MESSAGE_PREFIX")

            if userPrefix != "" {
                re := regexp.MustCompile(`\{\{USERNAME\}\}`)
                userPrefix = re.ReplaceAllString(userPrefix, m.Author.Username)
                updatedMessage = userPrefix + " " + updatedMessage
            }

            nomikin := NomiKin.NomiKin {
                ApiKey: companionToken,
                CompanionId: companionId,
            }

            companionReply := ""
            err = nil

            switch companionType {
                case "NOMI":
                companionReply, err = nomikin.SendNomiMessage(&updatedMessage)
                case "KINDROID":
                companionReply, err = nomikin.SendKindroidMessage(&updatedMessage)
            }
            if err != nil {
                fmt.Printf("Error exchanging messages with companion: %v", err)
                return nil
            }

            // Send as a reply to the message that triggered the response, helps keep things orderly
            _, sendErr := s.ChannelMessageSendReply(m.ChannelID, companionReply, m.Reference())
            if sendErr != nil {
                fmt.Println("Error sending message: ", err)
            }
            return nil
        }
    }
    return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore messages from the bot itself
    if m.Author.ID == s.State.User.ID {
        return
    }

    message := QueuedMessage{
        Message: m,
        Session: s,
    }

    queue.Enqueue(message)
}

var queue MessageQueue

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Printf("Error loading .env file - hopefully environment variables are set another way: %v", err)
    }

    botToken := os.Getenv("DISCORD_BOT_TOKEN")
    if botToken == "" {
        fmt.Println("No bot token provided. Set DISCORD_BOT_TOKEN environment variable.")
        return
    }

    dg, err := discordgo.New("Bot " + botToken)
    if err != nil {
        log.Fatalf("Error creating Discord session: %v", err)
    }

    dg.AddHandler(messageCreate)

    err = dg.Open()
    if err != nil {
        log.Fatalf("Error opening Discord connection: %v", err)
    }

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
    discordStatus := version + " " + string(statusMessageContent)
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
    } else {
        log.Println("Status update successful")
    }

    go queue.ProcessMessages()

    fmt.Println("Bot is now running. Press CTRL+C to exit.")
    select {}
}
