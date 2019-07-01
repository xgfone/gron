# gron

A binary program like `crontab`, but more powerful.

## Build & Run

```shell
$ go build
```

```shell
NAME:
   gron - A crontab service

USAGE:
   gron [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     worker   An gron runner worker like crontab, but more powerful
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Gron Runner Worker

```shell
NAME:
   gron worker - An gron runner worker like crontab, but more powerful

USAGE:
   gron worker [command options] [arguments...]

OPTIONS:
   --addr value                The address to listen to. (default: ":8002")
   --job-cancel-hooks value    The comma-separated url lists to be called when a job is canceled.
   --job-result-hooks value    The comma-separated url lists to be called when a job is called.
   --job-schedule-hooks value  The comma-separated url lists to be called when a job is scheduled.
   --log-file value            The log file path, and the log is output to os.Stdout by default.
   --log-level value           The log level, such as debug, info, etc. (default: "info")
   --shell value               The shell path. (default: "/bin/sh")
   --timeout value             The default global timeout. (default: 1m0s)
```
