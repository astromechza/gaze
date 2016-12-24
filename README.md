# `gaze` - capture and log process execution
Gaze is a command line tool that can be used to monitor and report the execution of a command. It becomes really
powerful when used in `cron` entries and other commands that are run regularly such as scheduled backups and
updates. There is no point having a backup procedure that silently fails.

There are 3 types of behaviours that can be invoked once an execution has completed:
- `web` : Submit a POST or PUT request with a json payload to whatever url you want
- `command` : Run the given command with a json payload piped to stdin
- `logfile` : Simple logging of either structured json or human readable text to a given file path

The `web` and `command` behaviours are the most valuable as they allow you to take action upon failures or to 
generally monitor the health of the command being run. Use `web` to submit the payload to your own dashboard or 
use `command` to launch a script that submits data to a `graphite` monitoring instance.

Errors triggered while running behaviours do not affect the stdout/stderr output of the
command being executed and so are only visible when the `-debug` flag is provided. This is to allow commands to be
transparently logged without affecting other process flow while the captured command is part of a piped chain of
commands. For Example:

```
$ cat /etc/hosts | ./gaze grep localhost | rev
ecafretni kcabpool eht erugifnoc ot desu si tsohlacol #
tsohlacol	1.0.0.721
 tsohlacol             1::
```

The JSON payload looks something like the following:

```
$ ./gaze -json sleep 1 | python -m json.tool
{
    "captured_output": "",
    "command": [
        "sleep",
        "1"
    ],
    "elapsed_seconds": 1.0092849,
    "end_time": "2016-12-24T14:27:29.741858883-03:00",
    "exit_code": 0,
    "exit_description": "Execution finished with no error",
    "hostname": "Bens-MacBook-Pro.local",
    "name": "sleep.1",
    "start_time": "2016-12-24T14:27:28.732574103-03:00",
    "tags": []
}
```

### CLI

```
$ ./gaze -version
Version: 0.3 (commit fc3e45e @ 2016-12-24)

 .    '                   .  "   '
            .  .  .                 '      '
    "'       .   .
                                     '     '
  .    '      _______________
          ==c(___(o(______(_()
                  \=\
                   )=\
                  //|\\
                 //|| \\
                // ||  \\
               //  ||   \\
              //         \\

Project: https://github.com/AstromechZA/gaze
```

```
$ ./gaze -help
TODO

  -config string
    	path to a gaze config file (default = $HOME/.config/gaze.json)
  -debug
    	mutes normal stdout and stderr and just outputs debug messages
  -example-config
    	output an example config and exit
  -extra-tags string
    	comma-seperated extra tags to add to the structure
  -json
    	mutes normal stdout and stderr and just outputs the json report on stdout
  -name string
    	override the auto generated name for the task
  -version
    	Print the version string
```

### Configuration

Behaviours and tags are configured via a config file. The config file is either read from 
`$HOME/.config/gaze.toml` or from whatever file path the user provides on the `-config` flag. We use a `toml` 
format for now since it allows quite expressive configuration without the strictness or annoyance of JSON.

For Example:

```
$ ./gaze -example-config
tags = ["tagA", "tagB"]

[behaviours]
  [behaviours.cmd]
    type = "command"
    when = "successes"
    include_output = true
    [behaviours.cmd.settings]
      args = ["-m", "json.tool"]
      command = "python"
  [behaviours.logging]
    type = "logfile"
    when = "failures"
    include_output = false
    [behaviours.logging.settings]
      directory = "/var/log"
      filename = "gaze.log"
      format = "human"
  [behaviours.request]
    type = "web"
    when = "always"
    include_output = true
    [behaviours.request.settings]
      method = "POST"
      url = "http://127.0.0.1:8080"
      [behaviours.request.settings.headers]
        API-TOKEN = "MY_TOKEN"
```

Specifying the config and watching the debug log:
```
$ ./gaze -config /var/folders/sl/fvkg182n1_x0hn2k7pfkprcm0000gn/T/tmpgQgJxfgaze/gaze.toml -debug date
2016-12-24T14:27:30.808 gaze INFO - Logging initialised.
2016-12-24T14:27:30.808 gaze INFO - Loading config from /var/folders/sl/fvkg182n1_x0hn2k7pfkprcm0000gn/T/tmpgQgJxfgaze/gaze.toml
2016-12-24T14:27:30.809 gaze INFO - Loaded config: {
  "Behaviours": {
    "cmd": {
      "Type": "command",
      "When": "successes",
      "IncludeOutput": true,
      "Settings": {
        "args": [
          "-m",
          "json.tool"
        ],
        "command": "python"
      }
    },
    "logging": {
      "Type": "logfile",
      "When": "failures",
      "IncludeOutput": false,
      "Settings": {
        "directory": "/var/log",
        "filename": "gaze.log",
        "format": "human"
      }
    },
    "request": {
      "Type": "web",
      "When": "always",
      "IncludeOutput": true,
      "Settings": {
        "headers": {
          "API-TOKEN": "MY_TOKEN"
        },
        "method": "POST",
        "url": "http://127.0.0.1:8080"
      }
    }
  },
  "Tags": [
    "tagA",
    "tagB"
  ]
} (err: <nil>)
2016-12-24T14:27:30.809 gaze INFO - Attempting to use 'date' as commandName
Sat Dec 24 14:27:30 UYT 2016
2016-12-24T14:27:30.813 gaze INFO - Command exited with code 0
2016-12-24T14:27:30.813 gaze INFO - Running behaviour of type command..
2016-12-24T14:27:30.841 gaze INFO - Behaviour completed.
2016-12-24T14:27:30.841 gaze INFO - Running behaviour of type logfile..
2016-12-24T14:27:30.841 gaze INFO - Skipping because it only runs on failures
2016-12-24T14:27:30.841 gaze INFO - Running behaviour of type web..
2016-12-24T14:27:30.841 gaze INFO - Making POST request to http://127.0.0.1:8080..
2016-12-24T14:27:30.843 gaze INFO - Behaviour completed.
```

The provided `example_python_receiver.py` script acts as an example web server accepting the payload from the 
`web` behaviour. It's output looks something like the following:

```
2016-12-24 14:27:30,657 : INFO : Starting example server at: ('', 8080)...
2016-12-24 14:27:30,842 : INFO : Incoming POST request on /
2016-12-24 14:27:30,842 : INFO : Header 'content-length' -> '342'
2016-12-24 14:27:30,842 : INFO : Header 'accept-encoding' -> 'gzip'
2016-12-24 14:27:30,842 : INFO : Header 'api-token' -> 'MY_TOKEN'
2016-12-24 14:27:30,842 : INFO : Header 'user-agent' -> 'Go-http-client/1.1'
2016-12-24 14:27:30,842 : INFO : Header 'host' -> '127.0.0.1:8080'
2016-12-24 14:27:30,842 : INFO : Header 'content-type' -> 'application/json'
2016-12-24 14:27:30,842 : INFO : Content: {
  "captured_output": "Sat Dec 24 14:27:30 UYT 2016\n", 
  "hostname": "Bens-MacBook-Pro.local", 
  "name": "date", 
  "tags": [
    "tagA", 
    "tagB"
  ], 
  "start_time": "2016-12-24T14:27:30.809459399-03:00", 
  "exit_description": "Execution finished with no error", 
  "exit_code": 0, 
  "elapsed_seconds": 0.003985623, 
  "command": [
    "date"
  ], 
  "end_time": "2016-12-24T14:27:30.813445022-03:00"
}
127.0.0.1 - - [24/Dec/2016 14:27:30] "POST / HTTP/1.1" 204 -
```
