# delete

Delete a workspace

Aliases:

- rm

## Usage

```console
visca delete [flags] <workspace>
```

## Options

### --orphan

|      |                   |
| ---- | ----------------- |
| Type | <code>bool</code> |

Delete a workspace without deleting its resources. This can delete a workspace in a broken state, but may also lead to unaccounted cloud resources.

### -y, --yes

|      |                   |
| ---- | ----------------- |
| Type | <code>bool</code> |

Bypass prompts.
