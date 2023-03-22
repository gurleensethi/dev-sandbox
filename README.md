# dev-sandbox

Quickly run predefined sandbox environments inside docker containers.

# Install

```bash
go install github.com/gurleensethi/dev-sandbox@latest
```

# Usage

```text
NAME:
   dev-sandbox - run predefined sandbox templates in docker containers

USAGE:
   dev-sandbox [global options] command [command options] [arguments...]

COMMANDS:
   doctor               check for all requirements on your system
   list-templates, lst  list all the available templates
   list, ls             list all the dev sandboxes
   run, r               run a sandbox
   purge                delete all running sandboxes
   delete, rm           delete a sandbox
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

Run the `dev-sandbox doctor` command to check for system requirements.

<img alt="Welcome to VHS" src="https://raw.githubusercontent.com/gurleensethi/dev-sandbox/main/out.gif" />

# Sandbox Templates

| Name | Description |
| ---- | ----------- |
| golang | Go application (with module initalized).
| golang-web-app | A simple Go web app.
| nodejs | Nodejs environment.
| postgres | Postgres database.
| react | React application generated using vite.
| vanilla-javascript | A vanilla html/css/javascript environment.

# Contribution

PS: I built this tiny tool for myself, to quickly spin up programming sandboxes and open then in VSCode.

Contribtion is welcome, specially contributing more sandbox templates.
