# tokens

Manage personal access tokens

Aliases:

- token

## Usage

```console
visca tokens
```

## Description

```console
Tokens are used to authenticate automated clients to Visca.
  - Create a token for automation:

     $ visca tokens create

  - List your tokens:

     $ visca tokens ls

  - Remove a token by ID:

     $ visca tokens rm WuoWs4ZsMX
```

## Subcommands

| Name                                      | Purpose        |
| ----------------------------------------- | -------------- |
| [<code>create</code>](./tokens_create.md) | Create a token |
| [<code>list</code>](./tokens_list.md)     | List tokens    |
| [<code>remove</code>](./tokens_remove.md) | Delete a token |
