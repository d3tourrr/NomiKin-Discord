package main

import (
    "fmt"
    "image"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/bwmarrin/discordgo"
    "github.com/nfnt/resize"
)

func (c *Companion) RegisterSlashCommands(s *discordgo.Session) {
    command := &discordgo.ApplicationCommand{
        Name: "showconfig",
        Description: c.DiscordSession.State.User.Username + ": List configuration details",
    }

    _, err := s.ApplicationCommandCreate(c.DiscordSession.State.User.ID, "", command)
    if err != nil {
        c.Log("Cannot create slash command for showconfig: %v", err)
    }
    c.VerboseLog("Registered 'showconfig' slash command")
}

func (c *Companion) HandleSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if i.Type == discordgo.InteractionApplicationCommand {
        switch i.ApplicationCommandData().Name {
        case "showconfig":
            c.Log("Command 'showconfig' triggered [command enabled: %v]", c.ShowConfigEnabled)
            var embed *discordgo.MessageEmbed
            var err error
            desc := "Bot Info: " + Version + " [NomiKin-Discord](https://github.com/d3tourrr/NomiKin-Discord) by <@498559262411456566>"
            var user *discordgo.User
            if i.Member != nil && i.Member.User != nil {
                user = i.Member.User
            } else {
                user = i.User
            }
            footer := &discordgo.MessageEmbedFooter{
                Text: fmt.Sprintf("Requested for %v (%v - %v)\n[at UTC: %v]", user.GlobalName, user.Username, user.ID, time.Now().UTC().Format("2006-01-02 @ 15:04:04")),
                IconURL: user.AvatarURL("80"),
            }

            if !c.ShowConfigEnabled {
                embed = &discordgo.MessageEmbed{
                    Title: "`/showconfig` command disabled for this companion",
                    Description: desc,
                    Color: 0xff0000,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name: "Adjust `.env` file's `SHOWCONFIG_ENABLED` setting if you are responsible for this companion and want this command turned on.",
                            Value: "",
                            Inline: false,
                        },
                    },
                    Footer: footer,
                } 

                err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                    Type: discordgo.InteractionResponseChannelMessageWithSource,
                    Data: &discordgo.InteractionResponseData{
                        Embeds: []*discordgo.MessageEmbed{embed},
                    },
                })
            } else {
                err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                    Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
                    Data: &discordgo.InteractionResponseData{
                        Content: "Fetching configuration details...",
                    },
                })

                avatarUrl := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", s.State.User.ID, s.State.User.Avatar)
                color, err := GetPrimaryColorFromImage(avatarUrl)
                if err != nil {
                    color = 0xffffff
                }

                member, _ := c.DiscordSession.State.Member(i.GuildID, c.DiscordSession.State.User.ID)
                botNick := member.Nick
                if botNick == "" {
                    botNick = "<none>"
                }
                roles, _ := c.DiscordSession.GuildRoles(i.GuildID)
                roleMap := make(map[string]string)
                for _, role := range roles {
                    roleMap[role.ID] = role.Name
                }
                var botRoleNames []string
                for _, roleID := range member.Roles {
                    if roleName, ok := roleMap[roleID]; ok {
                        botRoleNames = append(botRoleNames, roleName)
                    }
                }

                // Bot info
                embed = &discordgo.MessageEmbed{
                    Title: fmt.Sprintf("Discord Bot Details: %v", c.DiscordSession.State.User.Username),
                    Description: desc,
                    Color: color,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name: "Discord Bot Info",
                            Value: fmt.Sprintf("**Bot ID:** `%v`\n**Username:** `%v`\n**Server Nickname:** `%v`\n**Roles:** `%v`", c.DiscordSession.State.User.ID, c.DiscordSession.State.User.Username, botNick, strings.Join(botRoleNames, "`, `")),
                            Inline: false,
                        },
                    },
                    Footer: footer,
                }

                _, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
                    Embeds: []*discordgo.MessageEmbed{embed},
                })

                // Companion info
                embed = &discordgo.MessageEmbed{
                    Title: fmt.Sprintf("Companion Details: %v", c.DiscordSession.State.User.Username),
                    Description: desc,
                    Color: color,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name: "Companion Info",
                            Value: fmt.Sprintf("**Companion ID:** `%v`\n**Companion Type:** `%v`", c.CompanionId, c.CompanionType),
                            Inline: false,
                        },
                        {
                            Name: "Message Prefix",
                            Value: fmt.Sprintf("`%v`", c.MessagePrefix),
                            Inline: true,
                        },
                        {
                            Name: "Reply Prefix",
                            Value: fmt.Sprintf("`%v`", c.ReplyPrefix),
                            Inline: false,
                        },
                    },
                    Footer: footer,
                }

                if c.CompanionType == "KINDROID" && c.ChatStyle == "ROOMS" {
                    embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
                        Name: "Kin Room Info",
                        Value: fmt.Sprintf("Share ID: `%v`\nRandom Response Default: `%v`\nContext Messages Sent: `%v`\nNSFW Filter: `%v`", c.KinShareId, c.KinRandomResponseDefault, c.KinRoomContextMessages, c.KinNsfwFilter),
                        Inline: false,
                    })
                }

                _, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
                    Embeds: []*discordgo.MessageEmbed{embed},
                })

                // Response triggers
                embed = &discordgo.MessageEmbed{
                    Title: fmt.Sprintf("Response Triggers: %v", c.DiscordSession.State.User.Username),
                    Description: desc,
                    Color: color,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name: "Ping/Replied To",
                            Value: fmt.Sprintf("`%v`", strconv.FormatBool(c.RespondPing)),
                            Inline: true,
                        },
                        {
                            Name: "Role Ping",
                            Value: fmt.Sprintf("`%v`", strconv.FormatBool(c.RespondRole)),
                            Inline: true,
                        },
                        {
                            Name: "Direct Message",
                            Value: fmt.Sprintf("`%v`", strconv.FormatBool(c.RespondDM)),
                            Inline: true,
                        },
                        {
                            Name: "Response Keywords",
                            Value: c.Keywords,
                            Inline: false,
                        },
                        {
                            Name: "Bot Reply Loop Prevention Max",
                            Value: fmt.Sprintf("`%v`", fmt.Sprint(c.BotReplyMax)),
                            Inline: false,
                        },
                    },
                    Footer: footer,
                }

                _, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
                    Embeds: []*discordgo.MessageEmbed{embed},
                })

                // Emoji to reactions
                embed = &discordgo.MessageEmbed{
                    Title: fmt.Sprintf("Emojis to Reactions Details: %v", c.DiscordSession.State.User.Username),
                    Description: desc,
                    Color: color,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name: "Emoji to Reactions Enabled",
                            Value: fmt.Sprintf("`%v`", c.EmojisToReact),
                            Inline: true,
                        },
                        {
                            Name: "Max Reactions",
                            Value: fmt.Sprintf("`%v`", c.MaxReactions),
                            Inline: true,
                        },
                        {
                            Name: "Emoji Allow List",
                            Value: fmt.Sprintf("%v", c.EmojiAllowList),
                            Inline: false,
                        },
                        {
                            Name: "Emoji Ban List",
                            Value: fmt.Sprintf("%v", c.EmojiBanList),
                            Inline: false,
                        },
                    },
                    Footer: footer,
                }

                _, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
                    Embeds: []*discordgo.MessageEmbed{embed},
                })

                // Chat style
                embed = &discordgo.MessageEmbed{
                    Title: fmt.Sprintf("Configuration Details: %v", c.DiscordSession.State.User.Username),
                    Description: desc,
                    Color: color,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name: "Chat Style",
                            Value: fmt.Sprintf("`%v`", c.ChatStyle),
                            Inline: false,
                        },
                    },
                    Footer: footer,
                }

                // only add c.Rooms if it's under 500 characters
                if len(c.Rooms) < 500 {
                    embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
                        Name: "Rooms JSON",
                        Value: fmt.Sprintf("```json\n%v\n```", c.Rooms),
                        Inline: false,
                    })
                } else {
                    embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
                        Name: "Rooms JSON",
                        Value: fmt.Sprintf("Rooms JSON is `%v` long which is too big to be attached. Check your `.env` file.", c.Rooms),
                        Inline: false,
                    })
                }

                _, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
                    Embeds: []*discordgo.MessageEmbed{embed},
                })
            }

            if err != nil {
                c.Log("Failed to respond to 'showconfig' with embed: %v", err)
            }
        }
    }
}

func GetPrimaryColorFromImage(url string) (int, error) {
    resp, err := http.Get(url)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    img, _, err := image.Decode(resp.Body)
    if err != nil {
        return 0, err
    }

    img = resize.Resize(100, 100, img, resize.Lanczos3)

    var r, g, b, count uint64
    for y := 0; y < img.Bounds().Dy(); y++ {
        for x := 0; x < img.Bounds().Dx(); x++ {
            color := img.At(x, y)
            rr, gg, bb, _ := color.RGBA()
            r += uint64(rr)
            g += uint64(gg)
            b += uint64(bb)
            count++
        }
    }

    r /= count
    g /= count
    b /= count
    discordColor := (int(r/256) << 16) | (int(b/256) << 8) | int(b/256)
    return discordColor, nil
}
