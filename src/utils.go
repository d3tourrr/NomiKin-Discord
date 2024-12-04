package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "os/signal"
    "path/filepath"
    "net/http"
    "reflect"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"
)

func PrintStructFields(c *Companion) {
    val := reflect.ValueOf(c).Elem()
    typ := reflect.TypeOf(c).Elem()

    fmt.Printf("Companion: %v\n", c.CompanionId)
    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        fieldName := typ.Field(i).Name
        fieldValue := field.Interface()

        if fieldName == "DiscordBotToken" || fieldName == "CompanionToken" {
            // Keep secrets a secret
            fieldValue = "***redacted***"
        }

        fmt.Printf("  %s: %v\n", fieldName, fieldValue)
    }
}

func VerboseLog(s string, args ...interface{}) {
    if Verbose {
        parts := strings.Split(s, "%v")
        if len(parts)-1 != len(args) {
            log.Println("Verbose Logging: Number of format specifiers does not match number of arguments.")
        }

        var sb strings.Builder
        for i, part := range parts {
            sb.WriteString(part)
            if i < len(args) {
                sb.WriteString(fmt.Sprintf("%v", args[i]))
            }
        }

        log.Println(sb.String())
    }
}

func GetEnvFiles(dir string) ([]string, error) {
	var envFiles []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".env") {
			envFiles = append(envFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return envFiles, nil
}

func Contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}


func WaitForShutdown(bots []*discordgo.Session) {
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    <-stop
    for _, bot := range bots {
        bot.Close()
    }
    log.Println("All bots stopped.")
}

func UpdateStatus(dg *discordgo.Session) {
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

    discordStatus := Version + " " + string(statusMessageContent)
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

