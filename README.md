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
      * ⚠️ While you're on the `Bot` page, you must enable `Message Content Intent` or your companion will not be able to connect to Discord. (This is a new change to support responding to messages with certain keywords.)
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
1. Build and run the Docker container
   * Run either `start-windows-companion.ps1` on Windows (or in PowerShell) or `start-linux-companion.sh` on Linux (or in Bash, including Git Bash)
   * Or run the following commands (Note: the above scripts start the container in a detached state, meaning you don't see the log output. The below commands start the container in an attached state, which means you see the log output, but the container, and therefore the Companion/Discord integration dies when you close your console.)
     1. Build the Docker image: `docker build -t nomikin-discord .`
     1. Run the Docker container: `docker run nomikin-discord`
1. Interact with your companion in Discord!

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

# Updating

I'm adding new features to this integration with some frequency. To get the latest updates, run `git pull` in the directory you cloned in the above steps. Then, follow the setup steps to build and run the docker container again.

# Interacting in Discord with your companion

This integration is setup so that your companion will see messages where they are pinged (including replies to messages your companion's posts). Discord messages sent to companions are sent with a user configurable prefox to help your companion tell the difference between messages you send them in the Nomi app and messages that are sent to them from Discord. They look something like this.

> `*Discord Message from Bealy:* Hi @Vicky I'm one of the trolls that @.d3tour warned you about.`

In this message, a Discord user named `Bealy` sent a message to a companion named `Vicky` and also mentioned a Discord user named `.d3tour`.

Mentions of other users show that user's username Discord property, rather than their server-specific nickname. This was just the easiest thing to do and may change in the future (maybe with a feature flag you can set).

Companionss don't have context of what server or channel they are talking in, and don't see messages where they aren't mentioned in or being replied to.

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

