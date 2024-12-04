package main

import (
    "bufio"
    "encoding/json"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/bwmarrin/discordgo"
    NomiKin "github.com/d3tourrr/NomiKinGo"
)

type Companion struct {
    DiscordBotToken string
    CompanionToken  string
    CompanionId     string
    CompanionType   string
    MessagePrefix   string
    ReplyPrefix     string
    RespondPing     bool
    RespondRole     bool
    RespondDM       bool
    Keywords        string
    BotReplyMax     int
    ChatStyle       string
    Rooms           string
    NomiKin         NomiKin.NomiKin
    Tracker         BotMessageTracker
    Queue           MessageQueue
    DiscordSession  *discordgo.Session
    RoomObjects     map[string]Room
}

func (c *Companion) Setup(envFile string) {
    f, err := os.Open(envFile)
    if err != nil {
        log.Fatalf("Error loading %s: %v\n", envFile, err)
        return
    }

    VerboseLog("companion.Setup: %v", envFile)

    scanner := bufio.NewScanner(f)
    envVars := make(map[string]string)
    
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        if len(line) == 0 || line[0] == '#' {
            continue
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        envVars[key] = value
    }

    if err := scanner.Err(); err != nil {
        log.Fatalf("Error reading .env file: %v", err)
    }

    for key, value := range envVars {
        switch key {
        case "DISCORD_BOT_TOKEN":
            c.DiscordBotToken = value
        case "COMPANION_TOKEN":
            c.CompanionToken = value
        case "COMPANION_ID":
            c.CompanionId = value
        case "COMPANION_TYPE":
            if value != "NOMI" && value != "KINDROID" {
                log.Fatalf("Companion Type must be set to either `NOMI` or `KINDROID`. Set COMPANION_TYPE correctly in %v", envFile)
            } else {
                c.CompanionType = value
            }
        case "MESSAGE_PREFIX":
            c.MessagePrefix = value
        case "REPLY_PREFIX":
            c.ReplyPrefix = value
        case "RESPOND_TO_PING":
            c.RespondPing, err = strconv.ParseBool(value)
            if err != nil {
                log.Fatalf("RESPOND_TO_PING must be set to either TRUE or FALSE. Set RESPOND_TO_PING correctly in %v", envFile)
            }
        case "RESPOND_TO_ROLE_PING":
            c.RespondRole, err = strconv.ParseBool(value)
            if err != nil {
                log.Fatalf("RESPOND_TO_ROLE_PING must be set to either TRUE or FALSE. Set RESPOND_TO_ROLE_PING correctly in %v", envFile)
            }
        case "RESPOND_TO_DIRECT_MESSAGE":
            c.RespondRole, err = strconv.ParseBool(value)
            if err != nil {
                log.Fatalf("RESPOND_TO_DIRECT_MESSAGE must be set to either TRUE or FALSE. Set RESPOND_TO_DIRECT_MESSAGE correctly in %v", envFile)
            }
        case "RESPONSE_KEYWORDS":
            c.Keywords = value
        case "BOT_MESSAGE_REPLY_MAX":
            c.BotReplyMax, err = strconv.Atoi(value)
            if err != nil {
                log.Fatalf("Bot Message Reply Max was not set to a number. Fix BOT_MESSAGE_REPLY_MAX in %v", envFile)
            }
        case "CHAT_STYLE":
            if strings.Trim(value, "\"") == "ROOMS" {
                c.ChatStyle = "ROOMS"
            } else {
                c.ChatStyle = "NORMAL"
            }
        case "NOMI_ROOMS":
            c.Rooms = strings.Trim(value, "'")
        }
    }

    if _, exists := envVars["RESPOND_TO_PING"]; !exists {
        VerboseLog("RESPOND_TO_PING not present in config. Setting default value 'TRUE'.")
        c.RespondPing = true
    }

    if _, exists := envVars["BOT_MESSAGE_REPLY_MAX"]; !exists {
        VerboseLog("BOT_MESSAGE_REPLY_MAX not present in config. Setting default value '10'.")
        c.BotReplyMax = 10
    }

    if _, exists := envVars["CHAT_STYLE"]; !exists {
        VerboseLog("CHAT_STYLE not present in config. Setting default value 'NORMAL'.")
        c.ChatStyle = "NORMAL"
    }

    c.NomiKin = NomiKin.NomiKin{
        ApiKey: c.CompanionToken,
        CompanionId: c.CompanionId,
    }

    c.Tracker = NewBotMessageTracker()

    // For Nomi room
    if c.CompanionType == "NOMI" {
        c.NomiKin.Init()

        if c.ChatStyle == "ROOMS" {
            roomsString := c.Rooms
            if roomsString == "" {
                log.Fatalf("Companion %v is in ROOMS mode but no rooms were provided.", c.CompanionId)
            }

            var rooms []Room
            if err := json.Unmarshal([]byte(roomsString), &rooms); err != nil {
                log.Fatalf("Companion %v Error parsing NOMI_ROOMS: %v", c.CompanionId, err)
            }

            c.RoomObjects = map[string]Room{}
            for _, room := range rooms {
                VerboseLog("Creating/adding Nomi %v to room: %v\n  Note: %v\n  Backchanneling: %v\n  RandomResponseChance: %v", c.CompanionId, room.Name, room.Note, room.Backchanneling, room.RandomResponseChance)
                if room.RandomResponseChance > 100 || room.RandomResponseChance < 0 {
                    log.Printf("Error: RandomResponseChance must be between 0 and 100. Your value for Room %v is %v", room.Name, room.RandomResponseChance)
                }

                r, err := c.NomiKin.CreateNomiRoom(&room.Name, &room.Note, &room.Backchanneling, []string{c.CompanionId})
                if err != nil {
                    log.Printf("Error Nomi %v creating/adding to room %v\n Error: %v", c.CompanionId, room.Name, err)
                }

                c.RoomObjects[r.Name] = Room{Name: r.Name, Note: room.Note, Backchanneling: room.Backchanneling, Uuid: r.Uuid, Nomis: r.Nomis, RandomResponseChance: room.RandomResponseChance}

                if _, exists := RoomPrimaries[r.Name]; !exists {
                    // We are primary
                    VerboseLog("%v is primary for room %v", c.CompanionId, r.Name)
                    RoomPrimaries[r.Name] = c.CompanionId
                }
            }
        }
    }

    log.Printf("Finished setup of companion %v from file %v\n", c.CompanionId, envFile)
    if Verbose {
        PrintStructFields(c)
    }
}

func (c *Companion) RunDiscordBot() error {
    VerboseLog("companion.RunDiscordBot: %v", c.CompanionId)
    dg, err := discordgo.New("Bot " + c.DiscordBotToken)
    if err != nil {
        log.Fatalf("Error creating Discord session: %v", err)
    }

    c.DiscordSession = dg

    // For keyword triggering and Nomi room support
    dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

    dg.AddHandler(c.HandleMessageCreate)

    dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        go func() {
            for {
                time.Sleep(10 * time.Minute)
                c.Tracker.CleanupOldMessages()
            }
        }()
    })

    err = dg.Open()
    if err != nil {
        log.Fatalf("Error opening Discord connection: %v", err)
    }

    UpdateStatus(dg)
    statusTicker := time.NewTicker(10 * time.Minute)
    defer statusTicker.Stop()
    go func() {
        for {
            select {
            case <-statusTicker.C:
                UpdateStatus(dg)
            }
        }
    }()

    log.Printf("Assigning companion %s to bot %s", c.CompanionId, dg.State.User.ID)
    Companions[dg] = c

    go c.Queue.ProcessMessages()

    select {}
}

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

