# templates versions promote

Promote a template version to active.

## Usage

```console
visca templates versions promote [flags] --template=<template_name> --template-version=<template_version_name>
```

## Description

```console
Promote an existing template version to be the active version for the specified template.
```

## Options

### -t, --template

|             |                                   |
| ----------- | --------------------------------- |
| Type        | <code>string</code>               |
| Environment | <code>$VISCA_TEMPLATE_NAME</code> |

Specify the template name.

### --template-version

|             |                                           |
| ----------- | ----------------------------------------- |
| Type        | <code>string</code>                       |
| Environment | <code>$VISCA_TEMPLATE_VERSION_NAME</code> |

Specify the template version name to promote.

### -O, --org

|             |                                  |
| ----------- | -------------------------------- |
| Type        | <code>string</code>              |
| Environment | <code>$VISCA_ORGANIZATION</code> |

Select which organization (uuid or name) to use.
