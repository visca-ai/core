# templates pull

Download the active, latest, or specified version of a template to a path.

## Usage

```console
visca templates pull [flags] <name> [destination]
```

## Options

### --tar

|      |                   |
| ---- | ----------------- |
| Type | <code>bool</code> |

Output the template as a tar archive to stdout.

### --zip

|      |                   |
| ---- | ----------------- |
| Type | <code>bool</code> |

Output the template as a zip archive to stdout.

### --version

|      |                     |
| ---- | ------------------- |
| Type | <code>string</code> |

The name of the template version to pull. Use 'active' to pull the active version, 'latest' to pull the latest version, or the name of the template version to pull.

### -y, --yes

|      |                   |
| ---- | ----------------- |
| Type | <code>bool</code> |

Bypass prompts.

### -O, --org

|             |                                  |
| ----------- | -------------------------------- |
| Type        | <code>string</code>              |
| Environment | <code>$VISCA_ORGANIZATION</code> |

Select which organization (uuid or name) to use.
