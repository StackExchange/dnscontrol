# Notifications

DNSControl has build in support for notifications when changes are made. This allows you to post messages in team chat, or send emails when dns changes are made.

Notifications are written in the [notifications package](https://github.com/StackExchange/dnscontrol/tree/main/pkg/notifications), and is a really simple interface to implement if you want to add
new types or destinations.

## Configuration

Notifications are set up in your credentials JSON file. They will use the `notifications` key to look for keys or configuration needed for various notification types.

{% code title="creds.json" %}
```json
  "r53": {
      ...
    },
  "gcloud": {
        ...
  } ,
  "notifications": {
      "slack_url": "https://api.slack.com/apps/0XXX0X0XX0/incoming-webhooks",
      "teams_url": "https://outlook.office.com/webhook/00000000-0000-0000-0000-000000000000@00000000-0000-0000-0000-000000000000/IncomingWebhook/00000000000000000000000000000000/00000000-0000-0000-0000-000000000000"
  }
```
{% endcode %}

## Usage

If you want to send a notification, add the `--notify` flag to the `dnscontrol preview` or `dnscontrol push` commands.

Below is an example where we add [the A record](language-reference/domain/A.md) `foo` and display the notification output.

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

## Notification types

### Slack/Mattermost

If you want to use the Slack integration, you need to create a webhook in Slack.
Please see the [Slack documentation](https://api.slack.com/messaging/webhooks) or the [Mattermost documentation](https://docs.mattermost.com/developer/webhooks-incoming.html)

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

## Future work

Yes, this seems pretty limited right now in what it can do. We didn't want to add a bunch of notification types if nobody was going to use them. The good news is, it should
be really simple to add more. We gladly welcome any PRs with new notification destinations. Some easy possibilities:

- Email
- Generic Webhooks

Please update this documentation if you add anything.
