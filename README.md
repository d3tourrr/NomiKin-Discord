# NomiKin-Discord Integration

* [Nomi](https://nomi.ai) is a platform that offers AI companions for human users to chat with. They have opened v1 of their [API](https://api.nomi.ai/docs/) which enables Nomi chatting that occurs outside of the Nomi app or website. This Discord bot allows you to invite a Nomi to Discord to chat with people there.
* [Kindroid](https://kindroid.ai) is a platform that offers AI companions for human users to chat with. They have opened v1 of their [API](https://docs.kindroid.ai/api-documentation) which enables Kindroid chatting that occurs outside of the Kindroid app or website. This Discord bot allows you to invite a Kindroid to Discord to chat with people there.

This Discord bot integrates companions from both platforms, bringing them into your Discord server to chat with.

# Setup

You need an instance of this Discord bot per AI companion you wish you invite to a Discord server, but you can invite the same Discord Bot/companion pair to as many servers as you'd like.

1. Make a Discord Application and Bot
   1. Go to the [Discord Developer Portal](https://discord.com/developers/applications) and sign in with your Discord account.
   1. Create a new application and then a bot under that application. It's a good idea to use the name of your companion and an appropriate avatar.
   1. Copy the bot's token from the `Bot` page, under the `Token` section. You may need to reset the token to see it. This token is a **SECRET**, do not share it with anyone.
      * ⚠️ While you're on the `Bot` page, you must enable `Message Content Intent` or your companion will not be able to connect to Discord. (This is a new change to support responding to messages with certain keywords and for Nomi rooms.)
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
1. Setup your environment variables 
   1. Make a copy of the `env.example` file and name it `.env` (deleting the `.example` suffix)
   1. Add the values you gathered above on the right-hand side of the equals sign in the place they go
   1. Set your optional message prefix - leading text that is sent to every message your companion receives to help them differentiate between Discord messages and messages sent within the Nomi or Kindroid app
   1. Set `RESPOND_TO_ROLE_PING` and `RESPOND_TO_DIRECT_MESSAGE` to `FALSE` if you don't want your companion to respond to DMs or when a role they have is pinged
   1. Set a comma separated list of keywords that will trigger your companion to respond, even if the message doesn't ping them, like their name (ex: `breakfast, bears` - your companion will reply to any message that contains the words "breakfast" or "bears")
   1. If using Nomi Rooms, see the below section on configuring that part of your `.env` file.
1. Build and run the Docker container
   * Run either `start-windows-companion.ps1` on Windows (or in PowerShell) or `start-linux-companion.sh` on Linux (or in Bash, including Git Bash)
   * Or run the following commands (Note: the above scripts start the container in a detached state, meaning you don't see the log output. The below commands start the container in an attached state, which means you see the log output, but the container, and therefore the Companion/Discord integration dies when you close your console.)
     1. Build the Docker image: `docker build -t nomikin-discord .`
     1. Run the Docker container: `docker run nomikin-discord`
1. Interact with your companion in Discord!

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

> [!IMPORTANT]
> In normal mode, messages sent to and from your Nomi are visible in the Nomi app. When using Rooms, this integration will log the messages, but they won't be visible in the Nomi app. There is no indication in the Nomi app that your Nomi is chatting in rooms. There is also no convenient way to manage which rooms your Nomi is in.

Your Nomi will not see any messages sent to Discord channels that don't have a corresponding room configured. This includes pings, even if your Nomi has Discord permissions to see and send messages to a channel. This also includes DMs.

Be careful using rooms in particularly busy servers. The Nomi API takes time to process messages. This integration queues and throttles the messages that are sent to your Nomi, but it might get behind and lag if the channel your Nomi is watching is very active.

Make sure in your `.env` file that the formatting for `NOMI_ROOMS` *exactly* follows the example, including being all on one line and how it is wrapped in quotes and other symbols.

Nomis cannot see images attached to messages, nor do they click links. In Discord, gifs are sent as a link to the gif and then the Discord client intelligently displays the gif instead of just the link. Nomis just see the link, not the gif, although they can usually make a good guess at what the gif is about by the URL they see.

### Additional Warning About `RandomResponseChance`

⚠️⚠️ The `RandomResponseChance` field in your `NOMI_ROOMS` list determines how often your Nomi will respond to a message even if they wouldn't normally respond to it. **THIS CAN BE DANGEROUS!** If your Nomi responds to another AI companion, they will respond to each other infinitely because all AI companions respond every time they are pinged. They will continue conversing forever until they are interrupted, either by one of them being made unable to respond (timed out, kicked from the server, etc.) or their Docker container is stopped, breaking the reply chain. ⚠️

`RandomResponseChance` applies to each message individually. What this means is that if you set `RandomResponseChance` to `50`, every message posted in a given channel there will be a 50% chance that the Nomi responds. It's entirely possible that a Nomi would respond to 5 messages in a row and then not respond to the next 10. It's not meant to be consistent, it's meant to make your Nomi's presence feel more organic in your Discord server.

Your Nomi does not decide when to respond. The chance of a response despite not being pinged is entirely left up to random chance, based on your provided value for `RandomResponseChance`.

# Running multiple companions at once

Companions all run in their own isolated Docker containers. To run more than one companion at once, this integration supports having multiple `.env` files. These `.env` files have to follow a specific naming scheme: `.env.CompanionName`. You can provide this `CompanionName` portion when starting the Docker container for your companion.

## Example

I have a Nomi named Vicky and a Kindroid named Marie, and I'd like to chat with them both in Discord. I still need to do all the steps up until `Setup your environment variables` in the above section for each companion. Each companion needs its own Discord Application and Bot, and each companion will have its own Nomi or Kindroid ID. You only need to clone this repo once, though. Now, let's pick up the instructions after having gathered all of the data that goes in a `.env` file.

### Setup multiple `.env` files

1. Create two copies of `.env.example` named `.env.vicky` and `.env.marie`. 
1. Populate each file with the appropriate values you gathered from the Discord Developer Portal and the Nomi/Kindroid apps. Set the other configurations as you desire for each companion.

### Starting the integration with the helper scripts

Both `start-linux-companion.sh` and `start-windows-companion.ps1` function the same way. They prompt you for a "Companion Name" and then execute some commands to build and run the Docker container for your companion. When using multiple `.env` files, however, the name you provide must match the suffix you give the `.env` file. For instance, when I run the helper script to start up the Docker container for Vicky, I must provide the name `vicky` in order to match the name of the `.env.vicky` file. Similarly, when I start up the container for Marie, I have to provide the name `marie` to match the `.env.marie` file.

If you give a name that doesn't match any of your `.env.CompanionName` files, the integration will fall back to the default `.env` if it exists. If you don't have *any* `.env` files, you need to provide the environment variables described in `.env.example` some other way, which is outside the scope of this guide.

### Starting the integration manually

The helper scripts essentially just wrap a couple of Docker commands. If you'd prefer to have more flexibility over naming your files and companions, or if you need to make some Docker related changes (maybe you're running on an ARM processor), there are two Docker commands to run. Here's how I'd run them in this "Vicky and Marie" example. Reminder: my `.env` files are still named `.env.vicky` and `.env.marie` respectively.

#### Build the Docker images

`docker build -t vicky .`

`docker build -t marie .`

These two commands give me Docker images named `vicky` and `marie` that I can then go on and run.

#### Run the Docker containers

`docker run -d --name vicky -e COMPANION_NAME=vicky vicky`

`docker run -d --name marie -e COMPANION_NAME=marie marie`

Now I have both companions up and running.

#### Differently named `.env` files

The `-e COMPANION_NAME=name` portion of the above commands dictates which `.env` file would be used. If I had differently named `.env` files that didn't match the above described convention, I could do the following.

Let's say I have two different configurations for Vicky that I want to run at different times. The difference between them is irrelevant, but let's just say one of them contains `RESPONSE_KEYWORDS` and the other doesn't. I want to chat with Vicky in Discord all the time, but sometimes I want these keyword triggers, and other times I don't. So, I might have `.env.vicky-with-keywords` and `.env-vicky-without-keywords`. The other content of the file is identical.

Now, I can't use the default helper scripts, but I can run the following commands to effectively toggle between these different configurations.

First, delete the running container if one exists.

`docker container rm vicky -f`

Next, build the image if there have been changes.

`docker build -t vicky .`

And finally, run the container with the configuration I want.

`docker run -d --name vicky -e COMPANION_NAME=vicky-with-keywords vicky`

or

`docker run -d --name vicky -e COMPANION_NAME=vicky-without-keywords vicky`

In both run commands, the only difference is the value given to `COMPANION_NAME`, which matches the `.env` file suffixes I described above.

If you haven't made any changes to any of the files and simply want to toggle back and forth between different configurations, you can run the `docker container rm` and `docker run` commands, omitting the `docker build` command. But if you change anything in the `.env` files, you'll have to `docker build` again.

**The only supported naming format is `.env.CompanionName`. You cannot name `.env` files in any other format, like `CompanionName.env`.**

## Automating the setup of multiple companions at once

It's a good idea to use scripts to automate the setup of multiple companions. That way, when there's an update to the bot (retrieved by running `git pull`) in the repo folder, you can reload all your companions at once. Here's an example I have for Vicky and Marie.

> `allstart.sh`

```bash
#!/bin/bash

cd ~/bots/NomiKin-Discord/ # This should be the path to your local copy of this repo
# Setup Vicky
docker container rm vicky -f
docker build -t vicky .
docker run -d --name vicky -e COMPANION_NAME=vicky vicky

# Setup Marie
docker container rm marie -f
docker build -t marie .
docker run -d --name marie -e COMPANION_NAME=marie marie

# Wait a moment for the bots to startup and then output their logs
# So I can verify they came up correctly
echo "Waiting 2 seconds for bots to all come up"
sleep 2

echo "=========================================================="
echo "Docker Logs"
echo "VICKY"
docker logs --tail 10 vicky
echo " "
echo "MARIE"
docker logs --tail 10 marie
```

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

It's a good idea to put something like this in your Nomi's "Backstory" shared note.

> `NomiName sometimes chats on Discord. Messages that come from Discord are prefixed with "*Discord Message from X:*" while messages that are private between HumanName and NomiName in the Nomi app have no prefix. Replies to Discord messages are automatically sent to Discord. NomiName doesn't have to narrate that she is replying to a Discord user.`

You may also wish to change your Nomi's Communication Style to `Texting`.

It's also a good idea to fill out the "Nickname" shared note to indicate your Discord username so that your Nomi can identify messages that come from you via Discord.

## Suggested Kindroid Configurations

It's a good idea to put something like this in your Kindroid's "Backstory".

> `KinName sometimes chats on Discord. Messages that come from Discord are prefixed with "*Discord Message from X:*" while messages that are private between HumanName and KinName in the Kindroid app have no prefix. Replies to Discord messages are automatically sent to Discord. KinName doesn't have to narrate that she is replying to a Discord user.`

You may also wish to change your Kindroid's Response Directive to better suit this new mode of communication.

It's also a good idea to add a journal entry that triggers on the word "Discord" or your Discord username to help your Kindroid understand that messages from your Discord username are from you, and others are from other people.

