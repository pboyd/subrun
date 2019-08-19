`subrun` is a small utility that runs shell commands after receiving Google
Cloud Pub/Sub messages.

## Installation

To build from source, install Go 1.12 or greater and run:

```sh
go get github.com/pboyd/subrun
```

## Usage

Run it like:

```sh
subrun -config subrun.yaml
```

Where `subrun.yaml` is the path to a config file. The config file is YAML, here
is a minimal example:

```yaml
subscriptions:
- name: git-listener
  dir: /path/to/git/clone
  trigger:
    pubsub:
      project: fake-project
      topic: git-change-notice
      credentialsFile: /path/to/credentials.json
  tasks:
  - cmd: git pull
```

This will wait for a message on a `git-change-notice` topic then execute `git
pull` in the clone directory.

If all the commands return 0, the pubsub message will be `ack`ed. If any
command returns non-zero, the message will be `nack`ed, and no further commands
will be run.

The body of the message will be passed to `STDIN` on the executed commands.
