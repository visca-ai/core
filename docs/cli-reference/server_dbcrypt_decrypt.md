# server dbcrypt decrypt

Decrypt a previously encrypted database.

## Usage

```console
visca server dbcrypt decrypt [flags]
```

## Options

### --postgres-url

|             |                                       |
| ----------- | ------------------------------------- |
| Type        | <code>string</code>                   |
| Environment | <code>$VISCA_PG_CONNECTION_URL</code> |

The connection URL for the Postgres database.

### --postgres-connection-auth

|             |                                        |
| ----------- | -------------------------------------- |
| Type        | <code>password\|awsiamrds</code>       |
| Environment | <code>$VISCA_PG_CONNECTION_AUTH</code> |
| Default     | <code>password</code>                  |

Type of auth to use when connecting to postgres.

### --keys

|             |                                                            |
| ----------- | ---------------------------------------------------------- |
| Type        | <code>string-array</code>                                  |
| Environment | <code>$VISCA_EXTERNAL_TOKEN_ENCRYPTION_DECRYPT_KEYS</code> |

Keys required to decrypt existing data. Must be a comma-separated list of base64-encoded keys.

### -y, --yes

|      |                   |
| ---- | ----------------- |
| Type | <code>bool</code> |

Bypass prompts.
