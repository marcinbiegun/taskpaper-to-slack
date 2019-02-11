# taskpaper-to-slack

A tool for automating a workflow I'm following in my
currect distributed team - we keep our daily todo lists public as
messages on a Slack channel.

It works with plain text to-do lists created by
[Taskpaper](https://www.taskpaper.com/).

# How it works

It scans a Taskpaper file for nodes marked with `@slack(channel/msg)` and
syncs them to Slack.

The message must already exist at slack.

Only task nodes are synced (text nodes remain private).

`@done` and `@doing` tags are are recognized.

Tags and what comes after tags are removed.

## Example

You have this taskpaper file:

```
Some header:
  - buy milk

Monday, 11 Feb: @slack(B05KSNDD4/p1549566229043400)
  - read emails @done spent 40m
  lunch break 1h
  - release new version @doing
  - comment on pull requets
```

The corresponding Slack message will be replaced with:

```
:calendar: *Monday, 11 Feb*
:done: read emails
:doing: release new version
:todo: comment on pull requests
```

## Usage

Run with:

```
SLACK_TOKEN=foo SLACK_SUBDOMAIN=baz ./taskpaper-to-slack tasks.taskpaper
```
