# Changelog

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
