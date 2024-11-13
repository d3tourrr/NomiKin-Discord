
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "math/rand"
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

var version = "v0.6.1"

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

type Room struct {
    Name    string
    Note    string
    Uuid    string
    Backchanneling bool
    Nomis   []NomiKin.Nomi
    RandomResponseChance int
}

var Rooms map[string]Room

// Message queueing
var queue MessageQueue

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

        err := sendMessageToCompanion(queuedMessage.Session, queuedMessage.Message)
        if err != nil {
            log.Printf("Failed to send message to Companion API: %v", err)
            q.Enqueue(queuedMessage) // Requeue the message if failed
        }

        time.Sleep(5 * time.Second) // Try to keep from sending messages toooo quickly
    }
}

// Message formatting and handling
func sendMessageToCompanion(s *discordgo.Session, m *discordgo.MessageCreate) error {
    respondToThis := false

    // Is the companion mentioned or is this a reply to their message?
    for _, user := range m.Mentions {
        if user.ID == s.State.User.ID {
            respondToThis = true
            break
        }
    }

    // Is the companion mentioned by role
    // Doesn't work in DMs, no need to check if the bot is also mentioned specifically
    if strings.ToUpper(os.Getenv("RESPOND_TO_ROLE_PING")) == "TRUE" && !respondToThis && m.GuildID != "" {
        // Check this every time in case the bot is added to or removed from roles, not in DMs
        botID := s.State.User.ID

        botMember, err := s.GuildMember(m.GuildID, botID)
        if err != nil {
            return fmt.Errorf("Error retrieving bot member: %v", err)
        }

        for _, roleID := range botMember.Roles {
            roleMention := fmt.Sprintf("<@&%s>", roleID)
            if strings.Contains(m.Content, roleMention) {
                respondToThis = true
                break
            }
        }
    }

    // Does this message contain one of the reponse keywords?
    if os.Getenv("RESPONSE_KEYWORDS") != "" && !respondToThis {
        responseKeywords := os.Getenv("RESPONSE_KEYWORDS")
        words := strings.Split(responseKeywords, ",")
        charCleaner := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
        messageWords := strings.Fields(strings.ToLower(charCleaner.ReplaceAllString(m.Message.Content, "")))

        for _, word := range words {
            trimmedWord := strings.ToLower(strings.TrimSpace(word))
            for _, messageWord := range messageWords {
                if trimmedWord == messageWord {
                    respondToThis = true
                    break
                }
            }
            if respondToThis {
                break
            }
        }
    }

    // Is this a DM?
    respondToDM := strings.ToUpper(os.Getenv("RESPOND_TO_DIRECT_MESSAGE"))
    if m.GuildID == "" {
        // If this is a DM, respond if RESPOND_TO_DIRECT_MESSAGE is true, ignore if it's false,
        // and leave `respondToThis` at it's normal value otherwise - still respond if pinged/keyword
        switch respondToDM {
            case "TRUE":
            respondToThis = true
            case "FALSE":
            respondToThis = false
        }
    }

    // Is this a Nomi Room? Random chance to respond
    if os.Getenv("COMPANION_TYPE") == "NOMI" && os.Getenv("CHAT_STYLE") == "ROOMS" && !respondToThis && m.GuildID != "" {
        rand.Seed(time.Now().UnixNano())
        randomValue := rand.Float64() * 100
        if randomValue < float64(Rooms[m.ChannelID].RandomResponseChance) {
            respondToThis = true
            fmt.Printf("Nomi %v random response chance triggered. RandomResponseChance in channel %v set to %v.\n", os.Getenv("COMPANION_ID"), m.ChannelID, float64(Rooms[m.ChannelID].RandomResponseChance))
        }
    }

    if respondToThis {
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
            fmt.Printf("Improper companion type. Set COMPANION_TYPE environment variable to either 'NOMI' or 'KINDROID'. Your value: %v\n", companionType)
            return nil
        }

        // Replacing mentions makes it so the companion sees the usernames instead of <@userID> syntax
        updatedMessage := m.Content
        var err error
        if m.GuildID != "" {
            // But only if it's not a DM, otherwise this doesn't work - apparently this needs guild state info
            updatedMessage, err = m.ContentWithMoreMentionsReplaced(s)
            if err != nil {
                log.Printf("Error replacing Discord mentions with usernames: %v", err)
            }
        }

        // Add the message prefix if one exists, substitute sender username
        userPrefix := os.Getenv("MESSAGE_PREFIX")
        if userPrefix != "" {
            re := regexp.MustCompile(`\{\{USERNAME\}\}`)
            userPrefix = re.ReplaceAllString(userPrefix, m.Author.Username)
            updatedMessage = userPrefix + " " + updatedMessage
        }

        // Construct the NomiKin obj and send the message
        nomikin := NomiKin.NomiKin {
            ApiKey: companionToken,
            CompanionId: companionId,
        }

        companionReply := ""
        err = nil

        // set the typing indicator
        stopTyping := make(chan bool)
        go func() {
            for {
                select {
                case <-stopTyping:
                    return
                default:
                    s.ChannelTyping(m.ChannelID)
                    time.Sleep(5 * time.Second) // Adjust the interval as needed
                }
            }
        }()

        switch companionType {
        case "NOMI":
            if os.Getenv("CHAT_STYLE") == "ROOMS" {
                roomId := Rooms[m.ChannelID].Uuid
                _, err = nomikin.SendNomiRoomMessage(&updatedMessage, &roomId)
                time.Sleep(3 * time.Second) // Avoid Nomi not ready for message error
                companionReply, err = nomikin.RequestNomiRoomReply(&roomId, &nomikin.CompanionId)
            } else {
                companionReply, err = nomikin.SendNomiMessage(&updatedMessage)
            }
        case "KINDROID":
            companionReply, err = nomikin.SendKindroidMessage(&updatedMessage)
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
            _, sendErr := s.ChannelMessageSend(m.ChannelID, companionReply)
            if sendErr != nil {
                fmt.Println("Error sending message: ", err)
            }
        } else {
            _, sendErr := s.ChannelMessageSendReply(m.ChannelID, companionReply, m.Reference())
            if sendErr != nil {
                fmt.Println("Error sending message: ", err)
            }
        }

        return nil
    }

    // Even if a Nomi won't respond, if they are in ROOMS mode, we need to send the message to the correct room
    // TODO: Clean this up, duplicated code to sanitize and send a message... In fact, there's probably plenty of
    // duplicated and messy code to clean up...
    if os.Getenv("COMPANION_TYPE") == "NOMI" && os.Getenv("CHAT_STYLE") == "ROOMS" {
        nomikin := NomiKin.NomiKin {
            ApiKey: os.Getenv("COMPANION_TOKEN"),
            CompanionId: os.Getenv("COMPANION_ID"),
        }
        updatedMessage := m.Content
        var err error
        if m.GuildID != "" {
            updatedMessage, err = m.ContentWithMoreMentionsReplaced(s)
            if err != nil {
                log.Printf("Error replacing Discord mentions with usernames: %v", err)
            }
        }

        userPrefix := os.Getenv("MESSAGE_PREFIX")
        if userPrefix != "" {
            re := regexp.MustCompile(`\{\{USERNAME\}\}`)
            userPrefix = re.ReplaceAllString(userPrefix, m.Author.Username)
            updatedMessage = userPrefix + " " + updatedMessage
        }

        roomId := Rooms[m.ChannelID].Uuid
        _, err = nomikin.SendNomiRoomMessage(&updatedMessage, &roomId)
    }

    return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {
        return
    }

    message := QueuedMessage{
        Message: m,
        Session: s,
    }

    queue.Enqueue(message)
}

func main() {
    // Support for multiple .env files named .env.companionName
    namedEnvFile := false
    companionName := os.Getenv("COMPANION_NAME")
    if companionName != "" {
        namedEnvFile = true
        err := godotenv.Load(".env." + companionName)
        if err != nil {
            namedEnvFile = false
            log.Printf("Error loading .env.%v file: %v", companionName, err)
        } else {
            log.Printf("Loaded env file: .env.%v", companionName)
        }
    }
    if !namedEnvFile {
        // Fall back to .env file if there's no .env.companionName
        err := godotenv.Load()
        if err != nil {
            log.Printf("Error loading .env file: %v", err)
        } else {
            log.Printf("Loaded env file: .env")
        }
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

    // For keyword triggering and Nomi room support
    dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

    dg.AddHandler(messageCreate)

    err = dg.Open()
    if err != nil {
        log.Fatalf("Error opening Discord connection: %v", err)
    }

    // For Nomi rooms
    if os.Getenv("COMPANION_TYPE") == "NOMI" {
        companionToken := os.Getenv("COMPANION_TOKEN")
        if companionToken == "" {
            log.Fatal("No companion token provided. Set COMPANION_TOKEN environment variable.")
        }

        companionId := os.Getenv("COMPANION_ID")
        if companionId == "" {
            log.Fatal("No companion AI ID provided. Set COMPANION_ID environment variable.")
        }

        nomi := NomiKin.NomiKin {
            ApiKey: companionToken,
            CompanionId: companionId,
        }

        nomi.Init()

        if os.Getenv("CHAT_STYLE") == "ROOMS" {
            roomsString := os.Getenv("NOMI_ROOMS")
            if roomsString == "" {
                log.Fatal("Nomi is in ROOMS mode but no rooms were provided.")
            }

            var rooms []Room
            if err := json.Unmarshal([]byte(roomsString), &rooms); err != nil {
                log.Fatalf("Error parsing NOMI_ROOMS: %v", err)
            }

            Rooms = make(map[string]Room)

            for _, room := range rooms {
                log.Printf("Creating/adding Nomi %v to room\n Name: %v\n Note: %v\n Backchanneling: %v\n", nomi.CompanionId, room.Name, room.Note, room.Backchanneling)
                if room.RandomResponseChance > 100 || room.RandomResponseChance < 0 {
                    log.Fatalf("RandomResponseChance must be between 0 and 100. Your value for Room %v is %v", room.Name, room.RandomResponseChance)
                    return
                }

                r, err := nomi.CreateNomiRoom(&room.Name, &room.Note, &room.Backchanneling, []string{nomi.CompanionId})
                if err != nil {
                    log.Printf("Error Nomi %v creating/adding to room %v\n Error: %v", nomi.CompanionId, room.Name, err)
                }

                Rooms[r.Name] = Room{Name: r.Name, Note: room.Note, Backchanneling: room.Backchanneling, Uuid: r.Uuid, Nomis: r.Nomis, RandomResponseChance: room.RandomResponseChance}
            }
        }
    }

    updateStatus(dg)
    statusTicker := time.NewTicker(10 * time.Minute)
    defer statusTicker.Stop()
    go func() {
        for {
            select {
            case <- statusTicker.C:
                updateStatus(dg)
            }
        }
    }()


    // Kick off message processing
    go queue.ProcessMessages()

    fmt.Println("Bot is now running. Press CTRL+C to exit.")
    select {}
}

func updateStatus(dg *discordgo.Session) {
    // Set bot online/custom status
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
        log.Printf("Status update successful: %v", discordStatus)
    }
}
