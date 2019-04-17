---
layout: default
title: Notifications
---
# Notifications

DNSControl has build in support for notifications when changes are made. This allows you to post messages in team chat, or send emails when dns changes are made.

Notifications are written in the [notifications package](https://github.com/StackExchange/dnscontrol/tree/master/pkg/notifications), and is a really simple interface to implement if you want to add
new types or destinations.

## Configuration

Notifications are set up in your credentials json file. They will use the `notifications` key to look for keys or configuration needed for various notification types.

```
  "r53": {
      ...
    },
  "gcloud": {
        ...
  } ,
  "notifications":{
      "bonfire_url": "https://chat.meta.stackexchange.com/feeds/rooms/123?key=xyz"
  }
```

You also must run `dnscontrol preview` or `dnscontrol push` with the `-notify` flag to enable notification sending at all.

## Notification types

### Bonfire

This is stack overflow's built in chat system. This is probably not useful for most people.

Configure `bonfire_url` to be the full url including room and api key.

## Future work

Yes, this seems pretty limited right now in what it can do. We didn't want to add a bunch of notification types if nobody was going to use them. The good news is, it should 
be really simple to add more. We gladly welcome any PRs with new notification destinations. Some easy possibilities:

- Email
- Slack
- Generic Webhooks

Please update this documentation if you add anything.