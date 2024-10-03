# Interstellar

## Features

- monitor and get release from github

## Install and requirements

- Deps: `gh` (github cli)

## Diagrams

### New release Check

```mermaid
stateDiagram-v2
    check: Check release
    download: Download new release
    create_vm: Create new vm
    canary: Canary
    blue_green: Blue green
    state new_release <<choice>>
    state deploy_type <<fork>>
    state done_deploy <<join>>
    check --> new_release
    new_release --> check : No new release
    new_release --> download
    download --> create_vm
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