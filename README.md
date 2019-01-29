# taskpaper-to-slack

This program is for syncing tagged taskpaper nodes to messages on
Slack.

## Example

You have this taskpaper file:

```
Some header:
  - buy milk

Workday: @slack(messageid)
  - finish tasks
```

The Slack message with id `messageid` will be replaced with:

```
:calendar: *Workday*
:todo: finish tasks
```
