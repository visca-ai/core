# provisioner keys create

Create a new provisioner key

## Usage

```console
visca provisioner keys create [flags] <name>
```

## Options

### -t, --tag

|             |                                       |
| ----------- | ------------------------------------- |
| Type        | <code>string-array</code>             |
| Environment | <code>$VISCA_PROVISIONERD_TAGS</code> |

Tags to filter provisioner jobs by.

### -O, --org

|             |                                  |
| ----------- | -------------------------------- |
| Type        | <code>string</code>              |
| Environment | <code>$VISCA_ORGANIZATION</code> |

Select which organization (uuid or name) to use.
