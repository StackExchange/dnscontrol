# Notifications

DNSControl includes built-in support for notifications, enabling you to post messages in team chats or send emails whenever DNS changes occur, with the functionality implemented in the [notifications package](https://github.com/StackExchange/dnscontrol/tree/main/pkg/notifications).

## Configuration

Notifications are set up in your credentials JSON file. They will use the `notifications` key to look for keys or configuration needed for various notification services.

{% code title="creds.json" %}
```json
{
  "r53": {},
  "gcloud": {},
  "notifications": {
    "slack_url": "https://api.slack.com/apps/0XXX0X0XX0/incoming-webhooks",
    "teams_url": "https://outlook.office.com/webhook/00000000-0000-0000-0000-000000000000@00000000-0000-0000-0000-000000000000/IncomingWebhook/00000000000000000000000000000000/00000000-0000-0000-0000-000000000000",
    "shoutrrr_url": "discover://token@id"
  }
}
```
{% endcode %}

## Usage

If you want to send a notification, add the `--notify` flag to the `dnscontrol preview` or `dnscontrol push` commands.

Below is an example where we add [the A record](language-reference/domain-modifiers/A.md) `foo` and display the notification output.

{% code title="dnsconfig.js" %}
```diff
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
+ A("foo", "1.2.3.4"),
);
```
{% endcode %}

### Preview example

In case of `dnscontrol preview`:

```shell
dnscontrol preview --notify
```

**The notification output**

```shell
**Preview: example.com[my_provider] -** CREATE foo.example.com A (1.2.3.4 ttl=86400)
```

### Push example

In case of `dnscontrol push`:

```shell
dnscontrol push --notify
```

**The notification output**

```shell
Successfully ran correction for **example.com[my_provider]** - CREATE foo.example.com A 1.2.3.4 ttl=86400
```

## Notification services

### Slack/Mattermost

If you want to use the Slack integration, you need to create a webhook in Slack.
Please see the [Slack documentation](https://api.slack.com/messaging/webhooks) or the [Mattermost documentation](https://developers.mattermost.com/integrate/webhooks/incoming/)

Configure `slack_url` to this webhook. Mattermost works as well, as they share the same api,

### Microsoft Teams

If you want to use the Teams integration, you need to create a webhook in Teams.
Please see the [Teams documentation](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook#add-an-incoming-webhook-to-a-teams-channel)

Configure `teams_url` to this webhook.

### Telegram

If you want to use the [Telegram](https://telegram.org/) integration, you need to create a Telegram bot and obtain a Bot Token, as well as a Chat ID. Get a Bot Token by contacting [@BotFather](https://telegram.me/botfather), and a Chat ID by contacting [@myidbot](https://telegram.me/myidbot).

Configure `telegram_bot_token` and `telegram_chat_id` to these values.

### Bonfire

This is Stack Overflow's built in chat system. This is probably not useful for most people.

Configure `bonfire_url` to be the full url including room and api key.

### Shoutrrr (email, Discord, Pushover, etc.)

DNSControl supports various notification methods via Shoutrrr, including email (SMTP), Discord, Pushover, and many others. For detailed setup instructions, click on the desired service:

* [Bark](https://containrrr.dev/shoutrrr/latest/services/bark/)
* [Discord](https://containrrr.dev/shoutrrr/latest/services/discord/)
* [Email](https://containrrr.dev/shoutrrr/latest/services/email/)
* [Google Chat](https://containrrr.dev/shoutrrr/latest/services/googlechat/)
* [Gotify](https://containrrr.dev/shoutrrr/latest/services/gotify/)
* [IFTTT](https://containrrr.dev/shoutrrr/latest/services/ifttt/)
* [Join](https://containrrr.dev/shoutrrr/latest/services/join/)
* [Matrix](https://containrrr.dev/shoutrrr/latest/services/matrix/)
* [Mattermost](https://containrrr.dev/shoutrrr/latest/services/mattermost/)
* [Ntfy](https://containrrr.dev/shoutrrr/latest/services/ntfy/)
* [OpsGenie](https://containrrr.dev/shoutrrr/latest/services/opsgenie/)
* [Pushbullet](https://containrrr.dev/shoutrrr/latest/services/pushbullet/)
* [Pushover](https://containrrr.dev/shoutrrr/latest/services/pushover/)
* [Rocketchat](https://containrrr.dev/shoutrrr/latest/services/rocketchat/)
* [Slack](https://containrrr.dev/shoutrrr/latest/services/slack/)
* [Teams](https://containrrr.dev/shoutrrr/latest/services/teams/)
* [Telegram](https://containrrr.dev/shoutrrr/latest/services/telegram/)
* [Zulip Chat](https://containrrr.dev/shoutrrr/latest/services/zulip/)

For a full overview of supported methods and configuration details, refer to the [Shoutrrr documentation](https://containrrr.dev/shoutrrr/latest/services/overview/).

Configure `shoutrrr_url` with the Shoutrrr URL to be notified.

{% code title="creds.json" %}
```json
{
  "notifications": {
    "shoutrrr_url": "discover://token@id"
  }
}
```
{% endcode %}
