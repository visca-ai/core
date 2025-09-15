# open vscode

Open a workspace in VS Code Desktop

## Usage

```console
visca open vscode [flags] <workspace> [<directory in workspace>]
```

## Options

### --generate-token

|             |                                                |
| ----------- | ---------------------------------------------- |
| Type        | <code>bool</code>                              |
| Environment | <code>$VISCA_OPEN_VSCODE_GENERATE_TOKEN</code> |

Generate an auth token and include it in the vscode:// URI. This is for automagical configuration of VS Code Desktop and not needed if already configured. This flag does not need to be specified when running this command on a local machine unless automatic open fails.
