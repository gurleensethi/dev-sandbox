# dev-sandbox

Quickly run predefined sandbox environments inside docker containers.

# Install

```bash
go install github.com/gurleensethi/dev-sandbox@latest
```

# Usage

```bash
NAME:
   dev-sandbox - A new cli application

USAGE:
   dev-sandbox [global options] command [command options] [arguments...]

COMMANDS:
   doctor                        check for all requirements on your system
   list-templates, ls-templates  
   list, ls                      list all the dev sandboxes
   run, r                        run a sandbox
   purge                         delete all running sandboxes
   delete, rm                    delete a sandbox
   help, h                       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

<img alt="Welcome to VHS" src="https://raw.githubusercontent.com/gurleensethi/dev-sandbox/main/out.gif" />

# Templates

| Name | Description |
| ---- | ----------- |
| react | React application generated using vite.
| golang | Go application (with module initalized).
| go-web-app |  |
| nodejs | Nodejs environment.
| postgres | Postgres database.
