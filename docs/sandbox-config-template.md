# Sandbox Config Template

Let's take the example of golang template.

```yaml
golang:
  name: golang
  description: Go application (with module initalized).
  image: golang:latest
  vscodeConfig:
    applicationFolder: /playground-golang
  environment:
    - "KEY=VALUE"
  initCommand:
    - bash
    - "-c"
    - |
      mkdir -p /playground-golang

      go install -v golang.org/x/tools/gopls@latest

      cd /playground-golang
      go mod init playground-golang

      touch main.go
      cat > main.go << EOF
      package main

      import "fmt"

      func main() {
        fmt.Println("Hello Go Sandbox")
      }
      EOF

      # Sleep forever
      tail -f /dev/null
  ports:
    - containerPort: 8888
      hostPort: 8888
  messages:
    postStart: |
      Container might take a few seconds to be ready.
```

## Properties

- **name**

  - Name of the template.
  - Specified when running a template, e.g. `dev-sandbox run <name>`.
  - Shows up in `list-templates` command.

- **description**

  - Short text describing the template.
  - Shows up in `list-templates` command.

- **image**

  - Docker image to use when running the container.
  - Format: `iamge:tag`

- **vscodeConfig.applicationFolder**

  - Folder to open when opening the container in VSCode.

- **environment**

  - Set environment variables.
  - Format: `KEY=VALUE`.

- **ports**

  - Expose docker container port to host port.

- **messages**

  - Lifecycle messages that can be displayed throughout running the container.
  - Messages are go templates and following data is available for use:
    - `ContainerName`: Name of the docker container.

- **messages.postStart**
  - Message displayed after starting the container. Add any special information here. For example, postgres sandbox uses this to display the username, password and database.
