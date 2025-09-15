# schedule override-stop

Override the stop time of a currently running workspace instance.

## Usage

```console
visca schedule override-stop <workspace-name> <duration from now>
```

## Description

```console

  * The new stop time is calculated from *now*.
  * The new stop time must be at least 30 minutes in the future.
  * The workspace template may restrict the maximum workspace runtime.

 $ visca schedule override-stop my-workspace 90m
```
