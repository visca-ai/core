# users suspend

Update a user's status to 'suspended'. A suspended user cannot log into the platform

## Usage

```console
visca users suspend [flags] <username|user_id>
```

## Description

```console
 $ visca users suspend example_user
```

## Options

### -c, --column

|         |                                                    |
| ------- | -------------------------------------------------- |
| Type    | <code>[username\|email\|created at\|status]</code> |
| Default | <code>username,email,created at,status</code>      |

Specify a column to filter in the table.
