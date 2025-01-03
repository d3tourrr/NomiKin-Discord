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
    CompanionName   string
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
    EmojisToReact   bool
    MaxReactions    int
    EmojiAllowList  []string
    EmojiBanList    []string
    BotReplyMax     int
    ChatStyle       string
    Rooms           string
    NomiKin         NomiKin.NomiKin
    Tracker         BotMessageTracker
    Queue           MessageQueue
    DiscordSession  *discordgo.Session
    RoomObjects     map[string]Room
    ShowConfigEnabled bool
}

func (c *Companion) Setup(envFile string) {
    f, err := os.Open(envFile)
    if err != nil {
        log.Fatalf("Error loading %s: %v\n", envFile, err)
        return
    }

    if Verbose {
        // Can't use VerboseLog because the companion isn't setup yet
        log.Printf("companion.Setup: %v\n", envFile)
    }

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
        value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
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
                log.Fatalf("Companion Type must be set to either `NOMI` or `KINDROID`. Your value: '%v'. Set COMPANION_TYPE correctly in %v", value, envFile)
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
            c.RespondDM, err = strconv.ParseBool(value)
            if err != nil {
                log.Fatalf("RESPOND_TO_DIRECT_MESSAGE must be set to either TRUE or FALSE. Set RESPOND_TO_DIRECT_MESSAGE correctly in %v", envFile)
            }
        case "RESPONSE_KEYWORDS":
            c.Keywords = value
        case "EMOJIS_TO_REACT":
            c.EmojisToReact, err = strconv.ParseBool(value)
            if err != nil {
                log.Fatalf("EMOJIS_TO_REACT must be set to either TRUE or FALSE. Set EMOJIS_TO_REACT correctly in %v", envFile)
            }
        case "EMOJI_ALLOW_LIST":
            c.EmojiAllowList = strings.Split(value, "")
        case "EMOJI_BAN_LIST":
            c.EmojiBanList = strings.Split(value, "")
        case "MAX_REACTIONS":
            c.MaxReactions, err = strconv.Atoi(value)
            if err != nil {
                log.Fatalf("Max Reactions was not set to a number. Fix MAX_REACTIONS in %v", envFile)
            }
        case "BOT_MESSAGE_REPLY_MAX":
            c.BotReplyMax, err = strconv.Atoi(value)
            if err != nil {
                log.Fatalf("Bot Message Reply Max was not set to a number. Fix BOT_MESSAGE_REPLY_MAX in %v", envFile)
            }
        case "SHOWCONFIG_ENABLED":
            c.ShowConfigEnabled, err = strconv.ParseBool(value)
            if err != nil {
                log.Fatalf("SHOWCONFIG_ENABLED must be set to either TRUE or FALSE. Set SHOWCONFIG_ENABLED correctly in %v", envFile)
            }
        case "CHAT_STYLE":
            if value == "ROOMS" {
                c.ChatStyle = "ROOMS"
            } else {
                c.ChatStyle = "NORMAL"
            }
        case "NOMI_ROOMS":
            c.Rooms = strings.Trim(value, "'")
        }
    }

    if _, exists := envVars["RESPOND_TO_PING"]; !exists {
        c.VerboseLog("RESPOND_TO_PING not present in config. Setting default value 'TRUE'.")
        c.RespondPing = true
    }

    if _, exists := envVars["EMOJIS_TO_REACT"]; !exists {
        c.VerboseLog("EMOJIS_TO_REACT not present in config. Setting default value 'TRUE'.")
        c.EmojisToReact = true
    }

    if _, exists := envVars["MAX_REACTIONS"]; !exists {
        c.VerboseLog("MAX_REACTIONS not present in config. Setting default value '5'.")
        c.MaxReactions = 5
    }

    if _, exists := envVars["BOT_MESSAGE_REPLY_MAX"]; !exists {
        c.VerboseLog("BOT_MESSAGE_REPLY_MAX not present in config. Setting default value '10'.")
        c.BotReplyMax = 10
    }

    if _, exists := envVars["CHAT_STYLE"]; !exists {
        c.VerboseLog("CHAT_STYLE not present in config. Setting default value 'NORMAL'.")
        c.ChatStyle = "NORMAL"
    }

    if _, exists := envVars["SHOWCONFIG_ENABLED"]; !exists {
        c.VerboseLog("SHOWCONFIG_ENABLED not present in config. Setting default value 'TRUE'.")
        c.ShowConfigEnabled = true
    }

    c.NomiKin = NomiKin.NomiKin{
        ApiKey: c.CompanionToken,
        CompanionId: c.CompanionId,
    }

    c.Tracker = NewBotMessageTracker()

    // For Nomi room
    if c.CompanionType == "NOMI" {
        SuppressLogs(func() {
            c.NomiKin.Init()
        })

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
                c.VerboseLog("Creating/adding Nomi to room: %v\n  Note: %v\n  Backchanneling: %v\n  RandomResponseChance: %v", room.Name, room.Note, room.Backchanneling, room.RandomResponseChance)
                if room.RandomResponseChance > 100 || room.RandomResponseChance < 0 {
                    c.Log("Error: RandomResponseChance must be between 0 and 100. Your value for Room %v is %v", room.Name, room.RandomResponseChance)
                }

                r, err := c.NomiKin.CreateNomiRoom(&room.Name, &room.Note, &room.Backchanneling, []string{c.CompanionId})
                if err != nil {
                    c.Log("Error checking if room exists: %v", err)
                }

                c.RoomObjects[r.Name] = Room{Name: r.Name, Note: room.Note, Backchanneling: room.Backchanneling, Uuid: r.Uuid, Nomis: r.Nomis, RandomResponseChance: room.RandomResponseChance}

                if _, exists := RoomPrimaries[r.Name]; !exists {
                    // We are primary
                    c.VerboseLog("Is primary for room %v", r.Name)
                    RoomPrimaries[r.Name] = c.CompanionId
                }
            }
        }
    }

    c.Log("Finished companion setup from file %v", envFile)
    if Verbose {
        PrintStructFields(c)
    }

    if len(c.CompanionId) + len(c.CompanionName) > LogWidth {
        LogWidth = len(c.CompanionId) + len(c.CompanionName)
    }
}

func (c *Companion) RunDiscordBot() error {
    c.VerboseLog("companion.RunDiscordBot")
    dg, err := discordgo.New("Bot " + c.DiscordBotToken)
    if err != nil {
        log.Fatalf("Error creating Discord session: %v", err)
    }

    c.DiscordSession = dg

    // For keyword triggering and Nomi room support
    dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

    dg.AddHandler(c.HandleMessageCreate)
    dg.AddHandler(c.HandleSlashCommands)

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

    c.RegisterSlashCommands(dg)

    c.CompanionName = dg.State.User.Username
    c.Log("Assigning companion [%v | %v] to bot %v", c.CompanionId, c.CompanionName, dg.State.User.ID)
    Companions[dg] = c

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

    go c.Queue.ProcessMessages()

    select {}
}

