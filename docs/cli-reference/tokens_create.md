# tokens create

Create a token

## Usage

```console
visca tokens create [flags]
```

## Options

### --lifetime

|             |                                    |
| ----------- | ---------------------------------- |
| Type        | <code>string</code>                |
| Environment | <code>$VISCA_TOKEN_LIFETIME</code> |

Specify a duration for the lifetime of the token.

### -n, --name

|             |                                |
| ----------- | ------------------------------ |
| Type        | <code>string</code>            |
| Environment | <code>$VISCA_TOKEN_NAME</code> |

Specify a human-readable name.

### -u, --user

|             |                                |
| ----------- | ------------------------------ |
| Type        | <code>string</code>            |
| Environment | <code>$VISCA_TOKEN_USER</code> |

Specify the user to create the token for (Only works if logged in user is admin).
