# Interstellar

## Features

- monitor and get release from github
- run executable

## Install and requirements

- Deps: `gh` (github cli)

## Diagrams

### New release Check

```mermaid
stateDiagram-v2
    check: Check release
    download: Download new release
    choose_port: Choose a port
    add_backend: Add new process as backend
    create_vm: Run the process
    canary: Canary
    blue_green: Blue green
    state new_release <<choice>>
    state deploy_type <<fork>>
    state done_deploy <<join>>
    check --> new_release
    new_release --> check : No new release
    new_release --> download
    download --> choose_port
    choose_port --> add_backend 
    add_backend --> create_vm
    create_vm --> deploy_type
    deploy_type --> canary
    deploy_type --> blue_green
    canary --> done_deploy
    blue_green --> done_deploy
    done_deploy --> [*]
```

### Canary deploy

TODO

### Blue green deploy

TODO