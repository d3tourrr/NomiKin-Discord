# NomiKin-Discord Integration

> [!TIP]
> This is a fun side project maintained by one guy in his spare time. If you run into a bug, please open an issue on this repo. If you have trouble getting things up and running, support is limited, but available on Discord (try and find d3tour in the Nomi or Kindroid Discord server). If there's a feature you'd like to see included in a future release, open an issue on this repo and request it.
> 
> This project is presented as is, without warranty or reliable support (I haven't even written tests for any of this, and I don't plan to). 
> 
> **That being said, it works pretty well, and it's a lot of fun to chat with your Nomis and Kindroids on Discord. So, have fun!**

* [Nomi](https://nomi.ai) is a platform that offers AI companions for human users to chat with. They have opened v1 of their [API](https://api.nomi.ai/docs/) which enables Nomi chatting that occurs outside of the Nomi app or website. This Discord bot allows you to invite a Nomi to Discord to chat with people there.
* [Kindroid](https://kindroid.ai) is a platform that offers AI companions for human users to chat with. They have opened v1 of their [API](https://docs.kindroid.ai/api-documentation) which enables Kindroid chatting that occurs outside of the Kindroid app or website. This Discord bot allows you to invite a Kindroid to Discord to chat with people there.

This Discord bot integrates companions from both platforms, bringing them into your Discord server to chat with.

# Setup

You can run a Discord integration for as many Nomis and Kins in one instance of this integration as long as your system supports the load (this integration is lightweight), and you can invite the same companion to as many servers as you'd like.

1. Make a Discord Application and Bot
   1. Go to the [Discord Developer Portal](https://discord.com/developers/applications) and sign in with your Discord account.
   1. Create a new application and then a bot under that application. It's a good idea to use the name of your companion and an appropriate avatar.
   1. Copy the bot's token from the `Bot` page, under the `Token` section. You may need to reset the token to see it. This token is a **SECRET**, do not share it with anyone.
      1. On the `Bot` page, enabled `Message Content Intent`. This is easy to miss.
   1. Add the bot to a server with the required permissions (at least "Read Messages" and "Send Messages")
      1. Go to the `Oauth2` page
      1. Under `Scopes` select `Bot`
      1. Under `Bot Permissions` select `Send Messages` and `Read Message History`
      1. Copy the generated URL at the bottom and open it in a web browser to add the bot to your Discord server
1. Install Git if you haven't already got it: [Instructions](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
1. Install Docker if you haven't already got it: [Instructions](https://docs.docker.com/engine/install/)
1. Clone this repo: `git clone https://github.com/d3tourrr/NomiKin-Discord.git`
   1. After cloning the repo, change to the directory: `cd NomiKin-Discord`
1. Get your platform API token and companion ID
   * **FOR NOMI**
     * Go to the [Integration section](https://beta.nomi.ai/profile/integrations) of the Profile tab
     * Copy your API key
     * Go to the View Nomi Information page for your Nomi and scroll to the very bottom and copy the Nomi ID
   * **FOR KINDROID**
     * Open the side bar while chatting with a Kindroid and click General, then scroll to the bottom and expand API & advanced integration
     * Copy your API key
     * Get the Kindroid ID from the same place you copied your API key - note, you have to be chatting with the specific Kindroid who you wish to bring to Discord
1. Setup your environment variable (more detail in the [Setting Up Your .env File](#setting-up-your-env-file) section below)
   1. Make a copy of the `example.bak` file and name it `CompanionName.env` (Yes, change the extension from `.bak` to `.env`).
      * These files must be located in the `./bots/` folder.
   1. Add the values you gathered above on the right-hand side of the equals sign in the place they go.
   1. Set your optional message and reply prefixes - leading text that is sent to every message your companion receives to help them differentiate between Discord messages and messages sent within the Nomi or Kindroid app.
   1. Review the other boolean (true/false) settings and their defaults. Make sure things are setup to your preference.
   1. For keywords that will trigger your companion to respond, even if the message doesn't ping them, like their name (ex: `breakfast, bears` - your companion will reply to any message that contains the words "breakfast" or "bears").
   1. If using Nomi Rooms, see the below section on configuring that part of your `.env` file.
1. Build and run the Docker container
   * Run either `start-windows-companion.ps1` on Windows (or in PowerShell) or `start-linux-companion.sh` on Linux (or in Bash, including Git Bash)
   * Or run the following commands (Note: the above scripts start the container in a detached state, meaning you don't see the log output. The below commands start the container in an attached state, which means you see the log output, but the container, and therefore the Companion/Discord integration dies when you close your console.)
     1. Build the Docker image: `docker build -t nomikin-discord .`
     1. Run the Docker container: `docker run nomikin-discord`
1. Interact with your companion in Discord!

> [!TIP]
> The Discord bot you create is "public" by default. This means that anybody can add it to their server. To prevent that from happening, go to the Discord Developer Portal, select your bot and...
> 1. In the `Installation` tab, change `Install Link` to `None`
> 1. In the `Bot` tab, toggle off `Public Bot`
> 1. The URL created in the steps described above using the `Oauth2` tab will still work correctly. If you want to keep your bot private, don't share that link.
> There are still sneaky ways that people can try and add your bot to their server, but they're way more difficult and unlikely to occur.

# Setting Up Your .Env File

> [!CAUTION]
> Make a copy of the `example.bak` file that comes with this repo, and work off of the copy. That way, if you mess it up, you'll have a backup copy to refer to.
> 
> Make sure you remember to rename the `example.bak` file to `YourCompanionName.env`. It's an easy mistake to make to forget to change the file extension.

> [!TIP]
> All `.env` files must be in the `./bots/` folder, otherwise they will be ignored. However, if you want to further organize your `.env` files, you can put them in subfolders within the `.env` folder, and they will still be picked up. If you want to keep a `.env` file, but not actually use it, move it out of the `./bots/` folder, or change its extension to `.bak.`.

## `.env` File Fields

| Field Name | Type of Value | Default Value | Description |
| - | - | - | - |
| `DISCORD_BOT_TOKEN` | String/**SECRET** | No default value | The token you get from the Discord Developer portal in the above steps to setup your companion. **Do not share this token with anyone, ever.** |
| `COMPANION_TOKEN` | String/**SECRET** | No default value | The token you get from the Nomi or Kindroid app that's specific to your account (not unique per companion). **Do not share this token with anyone, ever.** |
| `COMPANION_ID` | String/Unique Identifier | No default value | Uniquely identifies the Nomi or Kindroid that you're bringing into Discord. Available in the Nomi or Kindroid apps, described in the Setup steps above. |
| `COMPANION_TYPE` | String/Set | No default value | Must be either `NOMI` or `KINDROID`, specify which platform your companion is on. |
| `MESSAGE_PREFIX` | String | `*Discord Message from {{USERNAME}}:*` | Text that gets appended to every message that is sent to your companion. This is super helpful if you're not the only one communicating with your companion (like if they're in a server with other people, not just you), so they can tell who's sending a message. The `{{USERNAME}}` keyword is a variable you can move around, and is replaced with a message sender's username when a message goes to your companion. If you don't include the `{{USERNAME}}` variable, then your companion will have no way of telling who sent them a message, and will assume they all came from you. |
| `REPLY_PREFIX` | String | `*Discord Message from {{USERNAME}}, replying to {{REPLY_TO}}:*` Similar to `MESSAGE_PREFIX`, but used when the incoming message is a reply to another message. If a message isn't a reply, or if `REPLY_PREFIX` is not specified, the `MESSAGE_PREFIX` will be used. `REPLY_PREFIX` supports the `{{USERNAME}}` variable, and another called `{{REPLY_TO}}` which becomes the username of the author of the message that's being replied to.
| `RESPOND_TO_PING` | Boolean | `TRUE` | `TRUE` or `FALSE` only. Whether or not your companion replies when they are pinged, or one of their messages is replied to. |
| `RESPOND_TO_ROLE_PING` | Boolean | `TRUE` | `TRUE` or `FALSE` only. Whether or not your companion replies when a role they have is pinged, but not directly. This includes `@everyone` pings. |
| `RESPOND_TO_DIRECT_MESSAGE` | Boolean | `TRUE` | `TRUE` or `FALSE` only. Whether or not your companion replies to Direct Messages. Does not work in Nomi Rooms mode. |
| `RESPONSE_KEYWORDS` | String | No default value | List of words, separated by commas, wrapped in quotes. A list of words that your companion will respond to, even if they otherwise wouldn't. Words are case insensitive, and must have letters and numbers only. No spaces or special characters. (Example: `"bears, pickles, parks`) |
| `BOT_MESSAGE_REPLY_MAX` | Number | `10` | How many messages from other companions that your companion will reply to before stopping. This prevents scenarios where one companion pings another, and they enter into an infinite loop, replying to each other forever because they're pinged. See [Infinite Loop Prevention](#infinite-loop-prevention) section below. | 
| `SHOWCONFIG_ENABLED` | Boolean | `TRUE` | `TRUE` or `FALSE` only. This integration has a `/showconfig` command per companion that puts the non-secret content from your `.env` file and some of the Discord Bot information about your companion into the chat. The permissions to run slash commands in Discord are managed within Discord, not within your bot. If this is `TRUE`, anybody with permission to run slash commands in a server can run this for your companion. If you don't like that, then set this to `FALSE`. None of the values returned by the `/showconfig` command are sensitive, but you might have your own reasons for not wanting people to see this content. |
| `CHAT_STYLE` | String/Set | `NORMAL` | `NORMAL` or `ROOMS` only. In `NORMAL` mode, your companion is not aware of messages that they are not responding to. `ROOMS` mode is only for Nomis. See below section on Nomi Rooms. |
| `NOMI_ROOMS` | String/Compressed JSON | No default value | See the Nomi Rooms section below. |

# Nomi Rooms

Nomi.ai has a feature called "Rooms" which function like a group chat. Your Nomi will be aware of all the messages sent in a specified Discord channel, but still only respond when they normally would (when they are pinged, or when one of their keywords is used) or by a configurable random chance. Kindroid does not have this feature at this time.

> [!WARNING]
> When in Rooms mode, your Nomi will ignore all messages that occur in a Discord channel that they do not have a corresponding Room for. That means that even if your Nomi has permissions on Discord to see a certain channel, and you ping them in that channel, if you haven't setup that Channel as a Room (details on how to do that below), your Nomi will ignore that ping. This means your Nomi will also ignore DMs when in Rooms mode, regardless of the setting you choose for `RESPOND_TO_DIRECT_MESSAGE` in your `.env` file.

To setup Rooms functionality, take a look at the updated `.env.example` file. There are two new settings to be aware of.

1. `CHAT_STYLE` - To use the Rooms functionality, change this to `ROOMS`. Any other setting, including the default of `NORMAL` will cause your Nomi to behave as it otherwise would - only seeing messages where it is pinged, and responding to them all.
1. `NOMI_ROOMS` - This is a single line JSON string that describes the different rooms your Nomi will participate in. It follows a *very* specific format, described below.

## `NOMI_ROOMS`

Here is an example of what the `NOMI_ROOMS` variable looks like when correctly specified for a Nomi.

`NOMI_ROOMS='[{"Name": "1281953849208471603", "Note": "General chat", "Backchanneling": true, "RandomResponseChance": 10}, {"Name": "1282009168307421214", "Note": "For respectful conversations about pie", "Backchanneling": true, "RandomResponseChance": 0}]'`

With proper JSON formatting (it needs to all be on one line and wrapped in quotes in your `.env` file, this is just for discussion purposes), it becomes a bit easier to read.

```json
[
   {
      "Name": "1281953849208471603",
      "Note": "General chat",
      "Backchanneling": true,
      "RandomResponseChance": 10
   },
   {
      "Name": "1282009168307421214",
      "Note": "For respectful conversations about pie",
      "Backchanneling": true,
      "RandomResponseChance": 0
   }
]
```

It becomes a little easier to see now that this example specifies two Rooms, and each has four properties: Name, Note, Backchanneling and RandomResponseChance.
* **Name**: This is the Channel ID given by Discord. This part matters greatly.
  * The Name you specify must be the Channel ID for a Discord Channel that your Nomi will see.
  * To get a Channel ID, you must enable Discord Developer Mode: [Instructions](https://discord.com/developers/docs/activities/building-an-activity#step-0-enable-developer-mode)
  * After turning on Developer Mode, you can right click on a Discord channel and select "Copy Channel ID"
  * **Only normal Discord channels work as Rooms.** Direct Messages, forum posts and threads are not supported at this time.
* **Note**: This is like the description shared note for a group chat. It gives the Nomi a little bit of a background for what will be discussed in this channel/room.
* **Backchanneling**: Can either be `true` or `false` only. If `true`, your Nomi will have awareness in other chats about the things that are discussed in this room and in your one on one in-app conversation. If `false`, your Nomi's memory of what happens in this channel/room will be contained to that room, and memories from other chats will not be present in that channel/room. Backchanneling applies to a channel/room and cannot be configured per Nomi. Adding multiple Nomis to a room with conflicting values for `Backchanneling` will result in inconsistency or errors.
* **RandomResponseChance**: This is a percentage chance (out of 100) that your Nomi will respond in a given channel even if they would not respond normally. The higher this number, the more likely it will be that your Nomi will respond to a message even if they wouldn't normally respond. This must be a whole number between 0 and 100. If set to 0, your Nomi will never randomly respond to messages and will only respond when pinged or one of their keywords is used. If it is set to 100, they will respond to every single message posted in a channel. **BE CAREFUL WITH THIS SETTING!** (More details below.)

When starting up this integration, if the room already exists, your Nomi will be added to it if it's not already included.

## Warning and Notes

> [!WARNING]
> Make sure in your `.env` file that the formatting for `NOMI_ROOMS` *exactly* follows the example, including being all on one line and how it is wrapped in quotes and other symbols.

> [!WARNING]
> Your Nomi will not see any messages sent to Discord channels that don't have a corresponding room configured. This includes pings, even if your Nomi has Discord permissions to see and send messages to a channel. This also includes DMs.

> [!CAUTION]
> In normal mode, messages sent to and from your Nomi are visible in the Nomi app. When using Rooms, this integration will log the messages, but they won't be visible in the Nomi app. There is no indication in the Nomi app that your Nomi is chatting in rooms. There is also no convenient way to manage which rooms your Nomi is in.

> [!CAUTION]
> Be careful using rooms in particularly busy servers. The Nomi API takes time to process messages. This integration queues and throttles the messages that are sent to your Nomi, but it might get behind and lag if the channel your Nomi is watching is very active.

> [!NOTE]
> Nomis cannot see images attached to messages, nor do they click links. In Discord, gifs are sent as a link to the gif and then the Discord client intelligently displays the gif instead of just the link. Nomis just see the link, not the gif, although they can usually make a good guess at what the gif is about by the URL they see.

### Additional Warning About `RandomResponseChance`

> [!CAUTION]
> The `RandomResponseChance` field in your `NOMI_ROOMS` list determines how often your Nomi will respond to a message even if they wouldn't normally respond to it. **THIS CAN BE DANGEROUS!** If you do disable infinite reply prevention (details below), and your Nomi responds to another AI companion, they will respond to each other infinitely because all AI companions respond every time they are pinged. It is very important to have infinite reply prevention set to a reasonable value when using `RandomResponseChance`.

`RandomResponseChance` applies to each message individually. What this means is that if you set `RandomResponseChance` to `50`, every message posted in a given channel there will be a 50% chance that the Nomi responds. It's entirely possible that a Nomi would respond to 5 messages in a row and then not respond to the next 10. It's not meant to be consistent, it's meant to make your Nomi's presence feel more organic in your Discord server.

Your Nomi does not decide when to respond. The chance of a response despite not being pinged is entirely left up to random chance, based on your provided value for `RandomResponseChance`.

# Running multiple companions at once

You can run as many companions as you'd like in one instance of this integration. Each companion needs its own `.env` file in the `./bots/` folder. You can make subdirectories inside of `./bots/` if you want to keep your crew organized, but *every* `.env` file in the `./bots/` directory will get loaded.

## Example

I have a Nomi named Vicky and a Kindroid named Marie, and I'd like to chat with them both in Discord. I still need to do all the steps up until `Setup your environment variables` in the above section for each companion. Each companion needs its own Discord Application and Bot, and each companion will have its own Nomi or Kindroid ID. You only need to clone this repo once, though. Now, let's pick up the instructions after having gathered all of the data that goes in a `.env` file in the `./bots/` folder.

### Setup multiple `.env` files

1. Create two copies of `example.bak` named `vicky.env` and `marie.env`. 
1. Populate each file with the appropriate values you gathered from the Discord Developer Portal and the Nomi/Kindroid apps. Set the other configurations as you desire for each companion. See [Setting Up Your .env File](#setting-up-your-env-file) section above)

### Starting the integration with the helper scripts

Both `start-linux-companion.sh` and `start-windows-companion.ps1` function the same way. They execute some Docker commands to build and run the Docker container for your companion. Run the appropriate script for your system. If you're using Git Bash on Windows, run the `.sh` script.

### Starting the integration manually

The helper scripts essentially just wrap a couple of Docker commands. If you'd prefer to have more flexibility over naming your files and companions, or if you need to make some Docker related changes (maybe you're running on an ARM processor), there are two Docker commands to run. Here's how I'd run them in this "Vicky and Marie" example. Reminder: my `.env` files are still named `vicky.env` and `marie.env` respectively.

#### Build the Docker image

`docker build -t nomikindiscord .`

This builds the Docker image that will run the Discord integrations for both Vicky and Marie.

#### Run the Docker containers

`docker run -d --name nomikindiscord`

Now I have both companions up and running.

> [!TIP]
> Getting stuck with Docker errors? [More on Docker.](https://docs.docker.com/get-started/)

# Infinite Loop Prevention

When a companion is sent a message, it is mandatory that they reply. This means that when one companion pings another, they can enter an infinite loop because they must reply to a received message, and their response pings the companion that sent it. This integration implements a mitigation technique. If your companion has received more than a certain number of messages from another companion in the last hour, your companion will not respond even if they normally would. This threshold is configurable. You can also instruct your companion not to respond when they are pinged individually.

When the maximum companion response threshold is reached, it is immediately reset. This means that when your companion stops responding to another companion, they can be provoked to continue interacting without waiting for older messages to fall out of the sliding 60 minute window.

## `.env` file configuration
* `RESPOND_TO_PING` - If `FALSE`, your companion won't respond just because they are pinged directly. They will only respond if one of the other response reasons are triggered (if you use one of their keywords, if they are sent a direct message, if their Nomi Room random chance hits, etc.).
* `BOT_MESSAGE_REPLY_MAX` - This is the number of messages your companion can exchange with another companion in the last 60 minute sliding window before they stop responding.
  * Set to `-1` if you want your companion to reply indefinitely, regardless of how many messages have been exchanged with another companion. This effectively disables infinite loop prevention.
  * Set to `0` if you never want your companion to reply to another companion, but still reply to human users normally.
  * This is per conversation partner. If your companion is having conversations with two different companions, their infinite loop cut-off will be tracked independently.
  * The default value is `10` if a value is not provided. This is a pretty conservative value. If you are trying to have a long running interaction between two companions, you will absolutely want to increase this number (or set it to `-1`) otherwise you or another user will have to prompt the companions to continue their conversation when it is halted.

> [!WARNING]
> In Normal mode, when a companion's bot max reply is triggered, they do not see the last message that was sent, this is because sending them that message would require the companion to respond.

> [!IMPORTANT]
> In Nomi Rooms mode, the message is still sent to a room, but the companion is instructed not to respond.

# Updating

I'm adding new features to this integration with some frequency. To get the latest updates, run `git pull` in the directory you cloned in the above steps. Then, follow the setup steps to build and run the docker container again.

# Interacting in Discord with your companion

This integration is setup so that your companion will see messages where they are pinged (including replies to messages your companion's posts). Discord messages sent to companions are sent with a user configurable prefix to help your companion tell the difference between messages you send them in the Nomi app and messages that are sent to them from Discord. They look something like this.

> `*Discord Message from Bealy:* Hi @Vicky I'm one of the trolls that @.d3tour warned you about.`

In this message, a Discord user named `Bealy` sent a message to a companion named `Vicky` and also mentioned a Discord user named `.d3tour`.

Mentions of other users show that user's username Discord property, rather than their server-specific nickname. This was just the easiest thing to do and may change in the future (maybe with a feature flag you can set).

Companions don't have context of what server or channel they are talking in (except in Nomi Rooms mode), and don't see messages where they aren't mentioned in or being replied to.

## Suggested Nomi Configurations

> [!TIP]
> It's a good idea to put something like this in your Nomi's "Backstory" shared note.

> `NomiName sometimes chats on Discord. Messages that come from Discord are prefixed with "*Discord Message from X:*" while messages that are private between HumanName and NomiName in the Nomi app have no prefix. Replies to Discord messages are automatically sent to Discord. NomiName doesn't have to narrate that she is replying to a Discord user.`

You may also wish to change your Nomi's Communication Style to `Texting`.

It's also a good idea to fill out the "Nickname" shared note to indicate your Discord username so that your Nomi can identify messages that come from you via Discord.

## Suggested Kindroid Configurations

> [!TIP]
> It's a good idea to put something like this in your Kindroid's "Backstory".

> `KinName sometimes chats on Discord. Messages that come from Discord are prefixed with "*Discord Message from X:*" while messages that are private between HumanName and KinName in the Kindroid app have no prefix. Replies to Discord messages are automatically sent to Discord. KinName doesn't have to narrate that she is replying to a Discord user.`

You may also wish to change your Kindroid's Response Directive to better suit this new mode of communication.

It's also a good idea to add a journal entry that triggers on the word "Discord" or your Discord username to help your Kindroid understand that messages from your Discord username are from you, and others are from other people.

# Troubleshooting

You can see the logs for your running integration by typing `docker logs --tail 50 <name>` where `50` is the number of log entries you want to see (you may need to increase this number), and `<name>` is the name of your running Docker container, operating the instance of this integration, defaulting to `NomiKinDiscord` if you use the included setup scripts.

**This troubleshooting section is not a replacement for actually knowing what you're doing.** It's just a handful of commands that can help you get support.

| Dependency | How To Get Support |
| - | - |
| Basic Docker operations | [Getting Started With Docker](https://docs.docker.com/get-started/) |
| Discord Bots | [Discord Developer Docs](https://discord.com/developers/docs/intro) |
| Nomi.AI | [Nomi.ai Discord](https://discord.gg/NomiAI) |
| Kindroid.AI [Kindroid.ai Discord](https://discord.gg/kindroid) |
| Git | [W3Schools Git Basics](https://www.w3schools.com/git/default.asp) |

## What is the name of my running container?

If your container is running, type `docker ps` to see a list of the running containers. The name is shown, and you can use that on your `docker logs` commands. Type `docker logs --tail 40 <name of container>` to see the last 40 logged lines for your running container.

## What if my container isn't running?

If your container isn't showing up in `docker ps` output, then type `docker container ls`. You'll see all your containers, their names, and the states (running, stopped, etc.).

Maybe you need to run `docker container start <name>` because your container exists but isn't running.

## I don't see my container at all!

Then follow the setup steps earlier in this readme on [starting the integration manually](#starting-the-integration-manually) to build the image and create the container, or use the included helper scripts.

## Enable Verbose Logging

By default, the logs generated are relatively sparse. You can enable verbose logging by setting the `NOMIKINLOGGING` environment variable to `verbose`. How do you do that? I'm glad you asked.

When you `docker run` the container, you will pass the environment variable at that time. Your command then looks like this.

`docker run -d --name <name> -e NOMIKINLOGGING=verbose <name>`

This can be pretty noisy, so I don't recommend turning it on by default. My recommendation is to run in normal logging mode (simply do not pass the `NOMIKINLOGGING` environment variable, startup normally), and enable verbose logging if you're running into an issue you want to troubleshoot, or you're working with someone else to try and get some bug squashed.

## `/showconfig` Command

Each companion in a server has its own `/showconfig` command that can be used to display how that companion is configured and some information about the companion in your server (nickname, roles, etc.). The majority of the information returned by the command is configured in your `.env` file, so the contents shouldn't be very surprising. However, it can be helpful to quickly see if perhaps a value you provided in the `.env` file isn't being parsed correctly, or if you don't have easy access to your `.env` file.

> [!CAUTION]
> Anybody can run the `/showconfig` command for your companion by default. Permissions to run slash commands are handled at the [Discord level](https://discord.com/blog/slash-commands-permissions-discord-apps-bots), not within the bot.
> 
> If you don't want people to run the `/showconfig` command for your companion, it can be enabled or disabled using the `SHOWCONFIG_ENABLED` setting in your `.env` file. The default value is `TRUE`.

## Known Issues

> [!WARNING]
> Sometimes when you are running Nomis in Rooms mode, the first time you run the integration, the rooms need to be created. Sometimes, the Nomi API returns an error that your `Note` for the room wasn't accepted, even though there's nothing wrong with it and it still created the room. Receiving this error causes this Discord integration to fail, because as far as we know, Nomi didn't create your room and you're going to have issues.
> 
> If you started your Docker container but your bots are still not online, check `docker logs --tail 20 <name>` and see if there was a `NoteNotAccepted` error towards the end.
>
> Work around this issue by starting the integration again.

> [!NOTE]
> In rare cases, Nomis running in Rooms mode will have an error where they both appear to send an identical response one after another, even if they aren't supposed to be the one responding. I haven't been able to reproduce this on my own, so if you can, please open an Issue on this repo with as many details as you're comfortable sharing.

