package main

import (
    "sync"
    "time"
)

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
    companion.VerboseLog("Message from %v count: %v/%v", botID, len(tracker.counts[botID]), companion.BotReplyMax)

    if tracker.GetMessageCount(botID) > companion.BotReplyMax {
        companion.Log("Loop Prevention triggered. Got more than %v (BOT_MESSAGE_REPLY_MAX) messages from bot %v within the last hour.", companion.BotReplyMax, botID)
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

