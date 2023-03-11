# Sandbox Config Template

Let's take the example of golang template.

```yaml
golang:
  name: golang
  description: Go application (with module initalized).
  image: golang:latest
  vscodeConfig:
    applicationFolder: /playground-golang
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

      sleep 100000
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
