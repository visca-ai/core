# support bundle

Generate a support bundle to troubleshoot issues connecting to a workspace.

## Usage

```console
visca support bundle [flags] <workspace> [<agent>]
```

## Description

```console
This command generates a file containing detailed troubleshooting information about the visca deployment and workspace connections. You must specify a single workspace (and optionally an agent name).
```

## Options

### -y, --yes

|      |                   |
| ---- | ----------------- |
| Type | <code>bool</code> |

Bypass prompts.

### -O, --output-file

|             |                                                |
| ----------- | ---------------------------------------------- |
| Type        | <code>string</code>                            |
| Environment | <code>$VISCA_SUPPORT_BUNDLE_OUTPUT_FILE</code> |

File path for writing the generated support bundle. Defaults to visca-support-$(date +%s).zip.

### --url-override

|             |                                                 |
| ----------- | ----------------------------------------------- |
| Type        | <code>string</code>                             |
| Environment | <code>$VISCA_SUPPORT_BUNDLE_URL_OVERRIDE</code> |

Override the URL to your visca deployment. This may be useful, for example, if you need to troubleshoot a specific visca replica.
