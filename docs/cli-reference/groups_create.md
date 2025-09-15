# groups create

Create a user group

## Usage

```console
visca groups create [flags] <name>
```

## Options

### -u, --avatar-url

|             |                                |
| ----------- | ------------------------------ |
| Type        | <code>string</code>            |
| Environment | <code>$VISCA_AVATAR_URL</code> |

Set an avatar for a group.

### --display-name

|             |                                  |
| ----------- | -------------------------------- |
| Type        | <code>string</code>              |
| Environment | <code>$VISCA_DISPLAY_NAME</code> |

Optional human friendly name for the group.

### -O, --org

|             |                                  |
| ----------- | -------------------------------- |
| Type        | <code>string</code>              |
| Environment | <code>$VISCA_ORGANIZATION</code> |

Select which organization (uuid or name) to use.
