# `gaze` - capture and log process execution 

Gaze is a command line tool that can be used to monitor and report the execution 
of a command. It becomes really powerful when used in `cron` entries and
other commands that are run regularly such as scheduled backups and updates. 
There is no point having a backup procedure that silently fails... 

> Warning: most of this has not been implemented yet

### TODO:

- Execute given arguments
- Time execution
- Capture Exit Code 
- Read config from file
- `-json` outputs json report on `stdout`. does not read config
- `-debug` hijacks normal stdout/stderr to print debug messages.
    conflicts with -json

### Config File Driven

Since adding tons of args to gaze would be cumbersome, most
of the configuration is provided by a config file in the current user
home directory (or overridden via the `-config` flag).

This config file defines where reports are logged to and how the reports
are forwarded to other applications or for further processing.

```
[behaviours]

[[<name of entry>]]
type = <behaviour type>
<settings>

[[<name of entry]]
type = <behaviour type>
<settings>
``` 

### Available behaviours

#### Log to file 

Logs the structured report to a rotating log file. This would be the most 
common behaviour and should generally be there by default. 

```
type = logfile
when = <string (always|failure|success)>
directory = <string default = /var/log/>
filename = <string default = gaze>
format = <string (human|machine)>
rotatepolicy = <string (none|daily|hourly|monthly|1MB|10MB|100MB|NMB)>
stderrpolicy = <string (include|tail|ignore)>
stdoutpolicy = <string (include|tail|ignore)>
```

#### Send over http

Sends the json report as a POST or PUT request to a given url.

```
type = web
when = <string (always|failure|success)>
method = <string (POST|PUT)
url = <string>
headers = <key-values>
http_proxy = <string>
https_proxy = <string>
stderrpolicy = <string (include|tail|ignore)>
stdoutpolicy = <string (include|tail|ignore)>
```

#### Run command

Runs the given command with the report supplied via stdin.

```
type = command
when = <string (always|failure|success)>
command = "myscript.sh"
stderrpolicy = <string (include|tail|ignore)>
stdoutpolicy = <string (include|tail|ignore)>
```

The `command` will be run as a shell command, so can include pipings and 
redirection.

This behaviour is really the most powerful and can be used to achieve some quite
complex things if necessary:

- send result to monitoring infrastructure (eg: Graphite)
- send an email or IM message

## Some example scenarios

### Scheduled backups

You have a crontab entry that backs up your database, but you need to make sure
you are monitoring both the successes and failures of this nightly backup.

Gaze config:

```ini
[behaviours]

[[log_all]]
type = log
when = always
stderrpolicy = ignore 
stdoutpolicy = ignore 
directory = /var/log/backups 
filename = gazelog.log 
format = human

[[email_me_failures]]
type = command 
when = failure
stderrpolicy = include 
stdoutpolicy = ignore 
command = "mail -s 'Backup failure' admin@example.com"
```

Crontab:

```
@daily  gaze -c /backup.toml rsync -a blah blah blah
```

Now you have a system in place to automatically capture a log file of all
the backup attempts including execution times, elapsed time, exit codes etc. and
when failures occur, an email will be sent to you that includes the stderr of 
the execution.

Easy!