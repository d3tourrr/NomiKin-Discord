package main

import (
    "fmt"
    "log"
    "os"
    "sync"
)

func main() {
    fmt.Println(Banner)
    fmt.Println("Version: " + Version)
    fmt.Println("Help, info, contact: github.com/d3tourrr/NomiKin-Discord")

    if os.Getenv("NOMIKINLOGGING") == "verbose" {
        Verbose = true
        fmt.Println("\n>>> Verbose logging enabled by 'NOMIKINLOGGING' environment variable set to 'verbose'")
    } else {
        fmt.Println("\n>>> Optional: Enable verbose logging by setting 'NOMIKINLOGGING' environment variable to 'verbose'.")
    }

    fmt.Println("\n\n" + `_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~"(`)
    fmt.Println()

    envFiles, err := GetEnvFiles("./bots")
    if err != nil {
        log.Fatalf("Failed to read env files: %v", err)
        return
    }

    if len(envFiles) == 0 {
        fmt.Println("Found no .env files. Make sure you have files named 'CompanionName.env' in your 'bots' folder. Make sure they don't still have a '.bak' extension like the example file does.\nExiting...")
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

            c.CompanionName = c.DiscordSession.State.User.Username

        }(companion)
    }

    fmt.Println("\n\nSetup Complete")
    fmt.Println()

    wg.Wait()
}

