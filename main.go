
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "os"
    "os/signal"
    "reflect"
    "regexp"
    "strconv"
    "strings"
    "syscall"
    "sync"
    "time"

    "github.com/bwmarrin/discordgo"
    NomiKin "github.com/d3tourrr/NomiKinGo"
)

var version = "v0.7.1"
var companions = make(map[*discordgo.Session]*Companion)
var roomPrimaries = make(map[string]string)

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

func printStructFields(c *Companion) {
    val := reflect.ValueOf(c).Elem()
    typ := reflect.TypeOf(c).Elem()

    fmt.Printf("Companion: %v\n", c.CompanionId)
    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        fieldName := typ.Field(i).Name
        fieldValue := field.Interface()

        fmt.Printf("  %s: %v\n", fieldName, fieldValue)
    }
}

func (c *Companion) Setup(envFile string) {
    f, err := os.Open(envFile)
    if err != nil {
        log.Fatalf("Error loading %s: %v\n", envFile, err)
        return
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
            if value == "\"ROOMS\"" {
                c.ChatStyle = "ROOMS"
            } else {
                c.ChatStyle = "NORMAL"
            }
        case "NOMI_ROOMS":
            c.Rooms = strings.Trim(value, "'")
        }
    }

    if _, exists := envVars["RESPOND_TO_PING"]; !exists {
        c.RespondPing = true
    }

    if _, exists := envVars["BOT_MESSAGE_REPLY_MAX"]; !exists {
        c.BotReplyMax = 10
    }

    if _, exists := envVars["CHAT_STYLE"]; !exists {
        c.ChatStyle = "NORMAL"
    }

    c.NomiKin = NomiKin.NomiKin{
        ApiKey: c.CompanionToken,
        CompanionId: c.CompanionId,
    }

    c.Tracker = NewBotMessageTracker()

    log.Printf("Finished setup of companion %v from file %v\n", c.CompanionId, envFile)
    // printStructFields(c)
}

func (c *Companion) RunDiscordBot() error {
    dg, err := discordgo.New("Bot " + c.DiscordBotToken)
    if err != nil {
        log.Fatalf("Error creating Discord session: %v", err)
    }

    c.DiscordSession = dg

    // For keyword triggering and Nomi room support
    dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

    dg.AddHandler(c.handleMessageCreate)

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

    // For Nomi rooms
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
                log.Printf("Creating/adding Nomi %v to room: %v\n", c.CompanionId, room.Name)
                // log.Printf("Creating/adding Nomi %v to room: %v\n  Note: %v\n  Backchanneling: %v\n  RandomResponseChance: %v\n", companion.CompanionId, room.Name, room.Note, room.Backchanneling, room.RandomResponseChance)
                if room.RandomResponseChance > 100 || room.RandomResponseChance < 0 {
                    return fmt.Errorf("RandomResponseChance must be between 0 and 100. Your value for Room %v is %v", room.Name, room.RandomResponseChance)
                }

                r, err := c.NomiKin.CreateNomiRoom(&room.Name, &room.Note, &room.Backchanneling, []string{c.CompanionId})
                if err != nil {
                    log.Printf("Error Nomi %v creating/adding to room %v\n Error: %v", c.CompanionId, room.Name, err)
                }

                c.RoomObjects[r.Name] = Room{Name: r.Name, Note: room.Note, Backchanneling: room.Backchanneling, Uuid: r.Uuid, Nomis: r.Nomis, RandomResponseChance: room.RandomResponseChance}

                if _, exists := roomPrimaries[r.Name]; !exists {
                    // We are primary
                    roomPrimaries[r.Name] = c.CompanionId
                }
            }
        }
    }

    updateStatus(dg)
    statusTicker := time.NewTicker(10 * time.Minute)
    defer statusTicker.Stop()
    go func() {
        for {
            select {
            case <-statusTicker.C:
                updateStatus(dg)
            }
        }
    }()

    log.Printf("Assigning companion %s to bot %s", c.CompanionId, dg.State.User.ID)
    companions[dg] = c

    go queue.ProcessMessages()

    select {}
}

func (companion *Companion) AmIPrimary(m *discordgo.MessageCreate) bool {
    sendPrimary := true
    if roomPrimaries[m.ChannelID] != companion.CompanionId {
        // We're not the primary
        log.Printf("%v is not primary for %v - %v is\n", companion.CompanionId, m.ChannelID, roomPrimaries[m.ChannelID])
        sendPrimary = false
    } else {
        log.Printf("%v is primary for %v\n", companion.CompanionId, m.ChannelID)
    }
    return sendPrimary
}

func (c *Companion) GetRoomMembers(roomId string) []string {
    roomInfo, err := c.NomiKin.RoomExists(&roomId)
    if err != nil {
        log.Printf("Error checking if room exists: %v\n", err)
        return []string{}
    }

    // Ensure roomInfo and roomInfo.Nomis are not nil
    if roomInfo == nil || roomInfo.Nomis == nil {
        log.Printf("Room info or room members are nil for room %s\n", roomId)
        return []string{}
    }

    // Collect members if the above checks pass
    var retMembers []string
    for _, m := range roomInfo.Nomis {
        if m.Uuid != "" {
            retMembers = append(retMembers, m.Uuid)
        }
    }

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
            return false
        }
        if c.CheckRoomStatus(roomId) == "Default" {
            return true
        }
        time.Sleep(time.Second * 1)
        waited++
    }
}

func getEnvFiles(dir string) ([]string, error) {
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

type BotMessageTracker struct {
    lock    sync.RWMutex
    counts  map[string][]time.Time
}

func NewBotMessageTracker() BotMessageTracker {
    return BotMessageTracker{
        counts: make(map[string][]time.Time),
    }
}

func (tracker *BotMessageTracker) CleanupOldMessages() {
    tracker.lock.Lock()
    countsCopy := make(map[string][]time.Time, len(tracker.counts))

    for botID, timestamps := range tracker.counts {
        countsCopy[botID] = append([]time.Time{}, timestamps...)
    }
    tracker.lock.Unlock()

    threshold := time.Now().Add(-60 * time.Minute)
    for botID, timestamps := range countsCopy {
        var validTimestamps []time.Time
        for _, timestamp := range timestamps {
            if timestamp.After(threshold) {
                validTimestamps = append(validTimestamps, timestamp)
            }
        }

        tracker.lock.Lock()
        tracker.counts[botID] = validTimestamps
        tracker.lock.Unlock()
    }
}

func (tracker *BotMessageTracker) TrackMessage(botID string, companion *Companion) bool {
    if companion.BotReplyMax == -1 {
        // Companion is set to reply forever. No point tracking.
        return true
    }

    tracker.lock.Lock()
    defer tracker.lock.Unlock()

    tracker.counts[botID] = append(tracker.counts[botID], time.Now())

    if tracker.GetMessageCount(botID) > companion.BotReplyMax {
        log.Printf("Received more than %v (BOT_MESSAGE_REPLY_MAX) messages from bot %v within the last hour. Not responding.", companion.BotReplyMax, botID)
        tracker.counts[botID] = []time.Time{}
        return false
    }

    return true
}

func (tracker *BotMessageTracker) GetMessageCount(botID string) int {
    timestamps, exists := tracker.counts[botID]
    if !exists {
        return 0
    }

    threshold := time.Now().Add(-60 * time.Minute)
    count := 0
    for _, timestamp := range timestamps {
        if timestamp.After(threshold) {
            count++
        }
    }

    return count
}

func updateMessage(m *discordgo.MessageCreate, companion *Companion) string {
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

// Message queueing
var queue MessageQueue

type QueuedMessage struct {
    Message   *discordgo.MessageCreate
    Companion *Companion
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

        err := sendMessageToCompanion(queuedMessage.Message, queuedMessage.Companion)
        if err != nil {
            log.Printf("Failed to send message to Companion API: %v", err)
            q.Enqueue(queuedMessage) // Requeue the message if failed
        }

        time.Sleep(5 * time.Second) // Try to keep from sending messages toooo quickly
    }
}

// Message formatting and handling
func sendMessageToCompanion(m *discordgo.MessageCreate, companion *Companion) error {
    respondToThis := false

    // Is the companion mentioned or is this a reply to their message?
    if companion.RespondPing {
        for _, user := range m.Mentions {
            if user.ID == companion.DiscordSession.State.User.ID {
                respondToThis = true
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
            log.Printf("Nomi %v random response chance triggered. RandomResponseChance in channel %v set to %v.\n", companion.CompanionId, m.ChannelID, float64(companion.RoomObjects[m.ChannelID].RandomResponseChance))
        }
    }

    if respondToThis {
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
            return nil
        }

        updatedMessage := updateMessage(m, companion)

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
                // Only forward the message to a room if this companion is the primary for the room - avoid duplication
                sendPrimary := companion.AmIPrimary(m)

                roomId := companion.RoomObjects[m.ChannelID].Uuid
                if sendPrimary {
                    canSend := companion.WaitForRoom(companion.RoomObjects[m.ChannelID].Uuid)
                    if canSend {
                        _, err = companion.NomiKin.SendNomiRoomMessage(&updatedMessage, &roomId)
                    } else {
                        log.Printf("Waited as long as we could, but room %v was never ready for a message from %v\n", m.ChannelID, companion.CompanionId)
                    }
                }
                if !loopBreak {
                    canSend := companion.WaitForRoom(companion.RoomObjects[m.ChannelID].Uuid)
                    if canSend {
                        companionReply, err = companion.NomiKin.RequestNomiRoomReply(&roomId, &companion.NomiKin.CompanionId)
                    } else {
                        log.Printf("Waited as long as we could, but room %v was never ready for a message from %v\n", m.ChannelID, companion.CompanionId)
                    }
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
    // TODO: Clean this up, duplicated code to sanitize and send a message... In fact, there's probably plenty of
    // duplicated and messy code to clean up...
    if companion.CompanionType == "NOMI" && companion.ChatStyle == "ROOMS" {
        updatedMessage := updateMessage(m, companion)
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
                log.Printf("Waited as long as we could, but room %v was never ready for a message from %v\n", m.ChannelID, companion.CompanionId)
            }
        }
    }

    return nil
}

func (companion *Companion) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
            for b, c := range companions {
                if b.State.User.ID == m.Author.ID && contains(roomMems, c.CompanionId) {
                    sharedNomiRoom = true
                    log.Printf("%v message from Companion %v is in the same room %v - not forwarding to the room\n", companion.CompanionId, c.CompanionId, m.ChannelID)
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

    queue.Enqueue(message)
}

func main() {
    envFiles, err := getEnvFiles("./bots")
    if err != nil {
        log.Fatalf("Failed to read env files: %v", err)
        return
    }

    log.Printf("Env files: %v\n", envFiles)

    var wg sync.WaitGroup
    for _, envFile := range envFiles {
        companion := &Companion{}
        companion.Setup(envFile)

        wg.Add(1)
        go func(c *Companion) {
            defer wg.Done()
            err := c.RunDiscordBot()
            if err != nil {
                log.Printf("Error running bot for companion %s: %v\n", c.CompanionId, err)
            }

        }(companion)
    }

    wg.Wait()
}

func waitForShutdown(bots []*discordgo.Session) {
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    <-stop
    for _, bot := range bots {
        bot.Close()
    }
    log.Println("All bots stopped.")
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
    }
}
