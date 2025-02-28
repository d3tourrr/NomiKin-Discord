package main

import (
    "bytes"
    "fmt"
    "io"
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

        if fieldName == "DiscordBotToken" || fieldName == "CompanionToken" || fieldName == "NomiKin" {
            // Keep secrets a secret
            fieldValue = "***redacted***"
        }

        fmt.Printf("  %s: %v\n", fieldName, fieldValue)
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
    var workingCompanion *Companion
    for b, c := range Companions {
        if b.State.User.ID == dg.State.User.ID {
            workingCompanion = c
            break
        }
    }

    statusMessageLocation := "https://raw.githubusercontent.com/d3tourrr/NomiKin-Discord/refs/heads/main/StatusMessage.txt"
    statusResp, err := http.Get(statusMessageLocation)
    if err != nil {
        workingCompanion.Log("Error retrieving status message: %v", err)
    }
    defer statusResp.Body.Close()

    statusMessageContent, err := io.ReadAll(statusResp.Body)
    if err != nil {
        workingCompanion.Log("Error reading status message: %v", err)
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
        workingCompanion.Log("Error setting status: %v", err)
    }

    workingCompanion.VerboseLog("Updated Status: %v", discordStatus)
}

func SuppressGetRoomLogs(fn interface{}, args ...interface{}) []interface{} {
    originalOutput := log.Writer()
    defer log.SetOutput(originalOutput)

    log.SetOutput(&bytes.Buffer{})

    fnValue := reflect.ValueOf(fn)
    if fnValue.Kind() != reflect.Func {
        panic("withSuppressedLogsReturningAny: argument must be a function")
    }

    // Prepare arguments
    reflectArgs := make([]reflect.Value, len(args))
    for i, arg := range args {
        reflectArgs[i] = reflect.ValueOf(arg)
    }

    results := fnValue.Call(reflectArgs)
    output := make([]interface{}, len(results))

    for i, result := range results {
        output[i] = result.Interface()
    }

    return output
}

func SuppressLogs(fn func()) {
    originalOutput := log.Writer()
    defer log.SetOutput(originalOutput)
    log.SetOutput(&bytes.Buffer{})
    fn()
}

