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
     * Or see [Nomi API Doc: Listing your Nomis](https://api.nomi.ai/docs/#listing-your-nomis)
   * **FOR KINDROID**
     * Open the side bar while chatting with a Kindroid and click General, then scroll to the bottom and expand API & advanced integration
     * Copy your API key
     * Get the Kindroid ID from the same place you copied your API key - note, you have to be chatting with the specific Kindroid who you wish to bring to Discord
1. Setup your environment variables 
   1. Make a copy of the `env.example` file and name it `.env` (deleting the `.example` suffix)
   1. Add the values you gathered above on the right-hand side of the equals sign in the place they go
1. Build and run the Docker container
   * Run either `start-windows-companion.ps1` on Windows (or in PowerShell) or `start-linux-companion.sh` on Linux (or in Bash, including Git Bash)
   * Or run the following commands (Note: the above scripts start the container in a detached state, meaning you don't see the log output. The below commands start the container in an attached state, which means you see the log output, but the container, and therefore the Companion/Discord integration dies when you close your console.)
     1. Build the Docker image: `docker build -t nomikin-discord .`
     1. Run the Docker container: `docker run nomikin-discord`
1. Interact with your companion in Discord!

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

