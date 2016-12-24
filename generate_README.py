#!/usr/bin/env python

import argparse
import os
from textwrap import dedent
import subprocess
import tempfile


DESTINATION_FILE = os.path.join(os.path.dirname(__file__), 'README.md')


def generate_content():
    lines = []

    lines.append("# `gaze` - capture and log process execution")
    lines.append(dedent("""
    Gaze is a command line tool that can be used to monitor and report the execution of a command. It becomes really
    powerful when used in `cron` entries and other commands that are run regularly such as scheduled backups and
    updates. There is no point having a backup procedure that silently fails.
    """))

    lines.append(dedent("""
    There are 3 types of behaviours that can be invoked once an execution has completed:
    - `logfile` : Simple logging of either structured json or human readable text to a given file path
    - `web` : Submit a POST or PUT request with a json payload to whatever url you want
    - `command` : Run the given command with a json payload piped to stdin
    """))

    lines.append(dedent("""
    Errors triggered while running behaviours do not affect the stdout/stderr output of the
    command being executed and so are only visible when the `-debug` flag is provided. This is to allow commands to be
    transparently logged without affecting other process flow while the captured command is part of a piped chain of
    commands. For Example:
    """))
    add_command_example(lines, "cat /etc/hosts | ./gaze grep localhost | rev")

    lines.append("The JSON payload looks something like the following:")
    lines.append("")
    add_command_example(lines, "./gaze -json sleep 1 | python -m json.tool")

    lines.append("### CLI")
    lines.append("")
    add_command_example(lines, "./gaze -help")

    lines.append("### Configuration")
    lines.append("")
    lines.append(dedent("""
    Behaviours and tags are configured via a config file. The config file is either read from 
    `$HOME/.config/gaze.toml` or from whatever file path the user provides on the `-config` flag. We use a `toml` 
    format for now since it allows quite expressive configuration without the strictness or annoyance of JSON."""))
    lines.append("")
    lines.append("For Example:")
    lines.append("")
    add_command_example(lines, "./gaze -example-config")

    lines.append("Specifying the config and watching the debug log:")

    cf = generate_example_config_file()
    add_command_example(lines, "./gaze -config {} -debug date".format(cf))

    text = "\n".join(lines)
    if not text.endswith("\n"):
        text += "\n"
    return text


def add_command_example(lines, command):
    lines.append("```")
    lines.append("$ {}".format(command))

    try:
        output = subprocess.check_output(command, stderr=subprocess.STDOUT, shell=True)
    except subprocess.CalledProcessError as e:
        output = e.output

    lines.append(output.strip())
    lines.append("```")
    lines.append("")


def generate_example_config_file():
    example_config = subprocess.check_output("./gaze -example-config", shell=True)
    tempdir = tempfile.mkdtemp("gaze")
    configfile = os.path.join(tempdir, "gaze.toml")
    with open(configfile, 'w') as f:
        f.write(example_config)
    return configfile


def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument("-o", "--overwrite", action='store_true', default=False, help="write the generated content to the destination file")
    return p.parse_args()


def main():
    args = parse_args()
    content = generate_content()
    if args.overwrite:
        with open(DESTINATION_FILE, 'w') as f:
            f.write(content)
    else:
        print content


if __name__ == '__main__':
    main()
