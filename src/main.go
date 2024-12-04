package main

import (
    "log"
    "os"
    "strings"
    "sync"
)

func main() {
    if strings.ToLower(os.Getenv("NOMIKINLOGGING")) == "verbose" {
        Verbose = true
        log.Printf("Verbose logging enabled by 'NOMIKINLOGGING' environment variable set to 'verbose'\n")
    } else {
        log.Printf(">>> Optional: Enable verbose logging for troubleshooting by setting 'NOMIKINLOGGING' environment variable to 'verbose'.\n")
    }

    envFiles, err := GetEnvFiles("./bots")
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

