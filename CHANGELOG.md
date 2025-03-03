# Changelog

## 0.9.2
* Bugfix : Consume a newer version of `NomiKinGo` that fixes a bug where Kindroids in non-`ROOMS` mode could not respond to messages.

## 0.9.1
* Feature: "Rooms" mode for Kindroids
  * Kindroids can now be configured to respond to messages in a channel randomly, rather than just when pinged.
  * This is done by setting the `CHAT_STYLE` variable in your `.env` file to `ROOMS` for Kindroids.
  * Kindroids in `ROOMS` mode have no long term memory. They pull from their backstory, key memories, and the last `KIN_ROOM_CONTEXT_MESSAGES` number of messages sent in the channel to generate a response.
  * See [README.md](./README.md) for more information on how to set this up.

## 0.8.3

* Feature: Log messages now include the name of the bot that generated them, not just the ID

## 0.8.2

* Feature: Emojis contained in a companion message are now applied as reactions to the message they are responding to. This comes with several new configuration points in the `.env` file.
  * `EMOJI_TO_REACT` - `TRUE` or `FALSE` - whether or not to react with emojis. Default is `TRUE`.
  * `MAX_REACTIONS` - The maximum number of reactions to apply to a message. Default is `5`.
  * `EMOJI_ALLOW_LIST` - A list of permissible emojis (wrapped in quotes, no spaces, no commas, etc. Ex: `"🐱🦁🐯🐈"`) to use as reactions. If there are emojis in this list, any emojis not in this list will be diregarded.
  * `EMOJI_BAN_LIST` - A list of emojis (wrapped in quotes, no spaces, no commas, etc. Ex: `"🐶🐕🦮🐕‍🦺"`) that should never be used as reactions. If there are emojis in this list, any emoji NOT in this list can be used as a reaction.
  * If both `EMOJI_ALLOW_LIST` and `EMOJI_BAN_LIST` are provided, `EMOJI_ALLOW_LIST` will take precedence and `EMOJI_BAN_LIST` will be ignored.
* Bugfix: Companions set to respond to DMs would not. Now, they do.
* Bugfix: `/showconfig` command is now broken up into several messages to work around Discord length limits.

## 0.8.1

> [!WARNING]
> Breaking Changes!
> 
> * `.env` files have a new naming format, and need to be saved somewhere new.
> * The new naming format is `CompanionName.env`.
> * `.env` files must go in the `./bots/` directory. You can make subdirectories inside of `./bots/` if you want to keep your crew organized, but *every* `.env` file in the `./bots/` directory will get loaded.

* **MAJOR CHANGE** - Run all your companions from one instance of this integration, in one Docker container
  * Put your `.env` files in the `./bots/` folder and they will all be automatically setup.
  * `.env` files now have a different naming format. They are now `CompanionName.env` instead of `.env.CompanionName`
  * No more support for a default `.env` file. All `.env` files need to have a name, even if you're only running one companion.
  * Run Nomis and Kindroids from the same instance.
* Nomi Rooms improvements when two Nomis are in the same Room (requires them to be on the same user's Nomi account).
  * They will now make sure that messages from Discord users are only forwarded to the Room once.
  * They will not forward messages from other Nomis in the same Room to the Room.
  * These two changes mean that when you're running multiple Nomis in the same Room, there's no duplication of messages sent to the Room.
  * A small random delay is given to each Nomi before they send a message to a Room - this is to prevent two or more Nomis from talking over each other (manifested as a `RoomNotReady` error). This isn't perfect. Each Nomi will pause for a random amount of time based on how many other Nomis are in the Room with them.
* Verbose logging mode
  * Default logging is "turned down" a little to make for cleaner logs. If you want more verbose logging (maybe you're troubleshooting an issue), set the `NOMIKINLOGGING` environment variable to `verbose` when starting your Docker container
  * Ex: `docker run --name nomikin-discord -e NOMIKINLOGGING=verbose nomikin-discord`
* `/showconfig` command
  * Each companion now has a `/showconfig` command that will display the configuration of that bot in your server.
  * Disable the `/showconfig` command in your `.env` file using the `SHOWCONFIG_ENABLED` setting.
* Cool new ASCII art when starting up

```
   _  __           _ __ ___
  / |/ /__  __ _  (_) //_(_)__
 /    / _ \/  ' \/ / ,< / / _ \
/_/|_/\___/_/_/_/_/_/|_/_/_//_/     __
    ____/ _ \(_)__ _______  _______/ /
   /___/ // / (_-</ __/ _ \/ __/ _  /
   __ /____/_/___/\__/\___/_/  \_,_/
  / /  __ __  ___/ /_  // /____  __ ______
 / _ \/ // / / _  //_ </ __/ _ \/ // / __/
/_.__/\_, /  \_,_/____/\__/\___/\_,_/_/
     /___/

Help, info, contact: github.com/d3tourrr/NomiKin-Discord


_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~_.~"(_.~"(_.~"(_.~"(_.~"(
```

## 0.7.1

* New feature: Bot loop prevention
  * Previously, if one companion pinged the other, they would be caught in a loop forever or until a user manually interrupted them.
  * Now, you can configure `RESPOND_TO_PING` in your `.env` file to prevent a companion from responding if they are directly pinged. They will still respond if some other reason to respond is triggered, like if one of their keywords is used.
  * **ADDITIONALLY** companions now keep track of how many replies they've exchanged with another bot in the last hour. If that number exceeds the amount set in your `.env` file's `BOT_MESSAGE_REPLY_MAX` setting, they will not respond. Set this to `-1` to have your companion never stop replying to other companions when they are pinged. Set it to `0` to have them never reply to another companion.
* Bugfix: v0.6.3 had a bug where DMs to companions didn't include the content of the message sent. This is fixed.

## v0.6.3

* New feature: `REPLY_PREFIX`
  * In your `.env` file, use the `REPLY_PREFIX` variable to give messages that are sent as replies to other messages a different prefix than normal.

## v0.6.2

* Bugfix: When using a Nomi in `ROOMS` mode, messages sent in channels that your Nomi had Discord permissions to see, but did not have a corresponding Room configured for, would throw an error in your logs. Now, messages for Nomis in Rooms mode that happen in channels that don't have a corresponding Room are dropped before they are otherwise processed.

## v0.6.1

* New feature! Nomi has released new API functionality that works like a group chat: [Rooms!](https://api.nomi.ai/docs/). This version of the NomiKin-Discord integration allows you to specify channels on Discord as Nomi Rooms.
  * Your Nomi still sends a message when they are pinged (or one of their keywords is used), the experience to get a Nomi to say something remains unchanged.
    * Except you may provide a value to the `RandomResponseChance` property in your `.env` file per channel/room that determines how likely it is that your Nomi will respond to a message in a given room despite a response not being explicitly triggered.
  * The difference now is that your Nomi will be aware of all the messages sent in the specified channel when they do eventually respond. They'll see all messages in the channels you specify whether pinged or not, but only reply when pinged or randomly per `RandomResponseChance`.
  * See the [readme](./README.md) file in this repo for information on setting this up.
  * You don't have to use Rooms. You can leave your existing Nomi configurations as they are, or use a `CHAT_STYLE` other than `ROOMS` to have the default experience where Nomis only see the messages they're pinged in.
* **Kindroid functionality is unchanged.**

## v0.5.2

* Bugfix: Companion now stops typing when an error occurs.

## v0.5.1

* Consume newer version of NomiKinGo that prevents the sending of oversized messages to companions.
* Auto-update bot status

## v0.5

* Support for multiple `.env` files, see new README.md instructions
* Replies in DMs are no longer a "reply" to the message that triggered it, rather just sent as a normal message to better emulate how users chat in DMs

## v0.4

* **⚠️BREAKING⚠️** Bot needs the `Message Content Intent` permission
  * In the Discord Developer Portal select your bot and navigate to the `Bot` page. Scroll down and toggle on `Message Content Intent`.
* `RESPOND_TO_ROLE_PING` environment variable toggles whether your companion will respond when a role it has is pinged
* `RESPOND_TO_DIRECT_MESSAGE` environment variable toggles whether your companion will respond to DMs
* `RESPONSE_KEYWORDS` environment variable specifies a list of words that your companion will respond to whether pinged or not

## v0.3

* Moved to consolidated Nomi and Kindroid integration
* Moved core messaging function to [github.com/d3tourrr/NomiKinGo](https://github.com/d3tourrr/NomiKinGo) (other integrations coming eventually!)
* Bot now shows `<name> is typing...` indicator while responding to a message

## v0.0.2

* Bot now shows online in Discord, along with the version of the integration and a link to this repo
* Added version number
* **Nomi** - Updated error message that gets displayed in Discord when a "NoReply" error is sent by the Nomi API. This happens when the Nomi takes too long to respond to a message. The message still shows in the Nomi app, but isn't returned via API.
* Added convenience scripts for managing Docker containers

## init

* Initial integration functionality
