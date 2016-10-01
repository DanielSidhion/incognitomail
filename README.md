IncognitoMail
====

IncognitoMail is a proof of concept implementation for the
[Incognito Emails](https://sidhion.com/blog/post/incognito-email-idea/)
idea.

This is the back-end part of the implementation.
It contains a server that should run
in the same machine as your Message Transfer Agent (MTA).
The front-end part is a
[browser add-on](https://github.com/DanielSidhion/incognitomail-addon)
that communicates with the server in this project.

## How it works

IncognitoMail has a very simple workflow.
Whenever you need to create a new account, a secret token will be generated.
Treat this token as that account's password!
Whoever has access to this token will be able to manage your account's handles.
To create a new account, you have to specify the email address
to send all messages received by the future Incognito Emails.
Whenever you need to create a new Incognito Email (a new _handle_),
pass the secret token for your account.
IncognitoMail will automatically generate a new address
and update your MTA to redirect all messages received
by this new address to the address registered with your account.
It's that simple,
and most of the work will be done by the browser add-on!

## Project Status

This is a new project, stability is currently not guaranteed.
Since it's in active development, parts may change (a lot).
IncognitoMail currently only supports
[Postfix](http://www.postfix.org/)
as the MTA.
A task list with current and future features is available below.
If you wish to use IncognitoMail, but need a certain feature added,
please open an issue.

## Current and future features and fixes

Checked boxes are features currently being developed.
All other items are planned features and may be changed or removed.

- [x] Add support for more commands from the command line utility
- [x] Improve usage output
- [x] Create tests for each module to ensure stability and robustness
- [ ] Remove old accounts and handles that were never used
- [ ] Check if a rate limiting feature for web sockets makes sense
- [x] Improve error messages returned to the user from the commandline tool
- [ ] Add permission checking before changing stuff in the MTA
- [x] Improve logging
- [ ] Support for multiple domains for the same MTA
- [x] Improve random string generation for secret tokens and handles

## Installing

To use IncognitoMail, you will first need
a working Postfix setup configured with either a
[canonical map](http://www.postfix.org/ADDRESS_REWRITING_README.html#canonical)
or a
[virtual alias map](http://www.postfix.org/ADDRESS_REWRITING_README.html#virtual).

- A canonical map is best suited if you are delivering messages to the local server
- A virtual alias map should be used if you plan to forward messages to another server, e.g. Gmail

[This guide](http://arstechnica.com/information-technology/2014/02/how-to-run-your-own-e-mail-server-with-your-own-domain-part-1/)
from ArsTechnica is great to follow and get your first Postfix setup running.

To get IncognitoMail, install Go and run:

```sh
$ go get github.com/danielsidhion/incognitomail/
$ go get github.com/danielsidhion/incognitomail/cmd/incognitomail
```

This will download the server code
and the `incognitomail` command line utility.

## Configuration

IncognitoMail can't function out of the box without configuration.
You need to at least provide information about your MTA setup.
The configuration file should be written in the
[INI format](https://en.wikipedia.org/wiki/INI_file),
which is very simple to use.
A sample file with all possible configuration options
is provided below for reference.

    [General] ; Make sure to include the section name before that section's keys!
    MailSystem = "postfix" ; For future reference only. Currently not used (server assumes postfix, will fail with any other value)
    UnixSockPath = "/tmp/incognito.sock" ; Address of the unix socket used for communication between the daemon process and the cli utility
    LockFilePath = "/var/lock/incognito.lock" ; Path to the file used to prevent two server processes from running at the same time
    ListenPath = "/incognitomail" ; Path where the HTTP server will listen for websocket connections
    ListenAddress = ":9090" ; Address for the HTTP server to listen. Always include the port number with the ":" prefix. An empty address (as in this case) will listen on all interfaces
    TLSCertFile = "server.pem" ; If using HTTPS, path to the server certificate. If signed by a CA, this file needs to be the concatenation of the server's certificate, any intermediates and the CA's certificate
    TLSKeyFile = "server.key" ; If using HTTPS, path to the private key file corresponding to the server certificate

    [Persistence]
    Type = "boltdb" ; For future reference only. Currently not used (server assumes boltdb, will fail with any other value)
    DatabasePath = "incognito.db" ; Path to the file where all the information about accounts and handles is stored

    [PostfixConfig]
    Domain = "@sidhion.com" ; The same domain configured in Postfix
    MapFilePath = "/tmp/postfix/canonical" ; Path to the map file used in Postfix. Can be either the canonical or the virtual alias map

## Usage

```
incognitomail [-c|--config <path>] [command [arguments]]

  -c string
    	path to a configuration file (shorthand)
  -config string
    	path to a configuration file
```

If the command is empty,
incognitomail will start a server process
and begin listening for connections.
Currently, available commands are:

- `new account <address>`: creates a new account, and registers `address` as the main email address to send all messages. Will output the generated secret token for that account
- `new handle <secret>`: creates a new handle for the account with the specified secret token
- `list <secret>`: lists all handles registered for a given account
- `stop`: stop the current server process

**Important**: please make sure that you run the server instance
with the same user/privileges as your Postfix setup!
Generally, this will mean running the server instance
as the `postfix` user,
but you should check the file and folder permissions
of your Postfix configuration to be sure.

## Daemonization

It is possible to run IncognitoMail as a daemon with the help of a service manager.
The only currently tested setup is with
[Upstart](http://upstart.ubuntu.com/),
but it shouldn't be much different from other managers.
You can use the sample configuration provided below for Upstart
(remember to modify it, specially the paths, to your setup):

    description "IncognitoMail service"
    author      "Daniel Sidhion"

    start on filesystem or runlevel [2345]
    stop on runlevel [!2345]

    script
      echo $$ > /var/run/incognitomail.pid
      exec incognitomail -c /etc/incognitomail.conf
    end script

    pre-stop script
      rm /var/run/incognitomail.pid
    end script