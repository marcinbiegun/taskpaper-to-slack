# taskpaper-to-slack

This program is for syncing tagged taskpaper nodes to messages on
Slack.

Work in progress, but it works.

## Example

You have this taskpaper file:

```
Some header:
  - buy milk

Today: @slack(messageid)
  - read emails @done
  - comment on pull requets
```

The Slack message with id `messageid` will be replaced with:

```
:calendar: *Workday*
:done: read emails
:todo: comment on pull requests
```

## Usage

Run with:

```
SLACK_TOKEN=foo SLACK_CHANNEL_ID=bar SLACK_SUBDOMAIN=baz go run main.go tasks.taskpaper
```