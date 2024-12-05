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
    fmt.Println("Help, info, contact: github.com/d3tourrr/NomiKin-Discord\n")

    if os.Getenv("NOMIKINLOGGING") == "verbose" {
        Verbose = true
        fmt.Println(">>> Verbose logging enabled by 'NOMIKINLOGGING' environment variable set to 'verbose'\n")
    } else {
        fmt.Println(">>> Optional: Enable verbose logging by setting 'NOMIKINLOGGING' environment variable to 'verbose'.\n")
    }

    fmt.Println(`_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~"(`)
    fmt.Println()

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

