# notifications

Manage visca notifications

Aliases:

- notification

## Usage

```console
visca notifications
```

## Description

```console
Administrators can use these commands to change notification settings.
  - Pause visca notifications. Administrators can temporarily stop notifiers from
dispatching messages in case of the target outage (for example: unavailable SMTP
server or Webhook not responding).:

     $ visca notifications pause

  - Resume visca notifications:

     $ visca notifications resume
```

## Subcommands

| Name                                             | Purpose              |
| ------------------------------------------------ | -------------------- |
| [<code>pause</code>](./notifications_pause.md)   | Pause notifications  |
| [<code>resume</code>](./notifications_resume.md) | Resume notifications |
