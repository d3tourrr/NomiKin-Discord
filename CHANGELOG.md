# Changelog

## v0.4
* **BREAKING** Bot needs the `Message Content Intent` permission
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
