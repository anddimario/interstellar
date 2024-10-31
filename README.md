# Interstellar

Application deployer

## Features

- monitor and get release from github
- run executable
- blue green and canary deploy
- cli

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

```mermaid
stateDiagram-v2
  wait: Wait canary window
  check: Check if healthy
  add: Add the new version
  not_healthy: Remove new version
  remove_old: Remove old version
  state is_healthy <<choice>>
  check --> is_healthy
  is_healthy --> add : is healthy
  is_healthy --> not_healthy
  add --> wait
  wait --> remove_old
  not_healthy --> [*]
  remove_old --> [*]
```

### Blue green deploy

```mermaid
stateDiagram-v2
  wait: Wait positive healthchecks
  check: Check if healthy
  replace: Replace old with new
  not_healthy: Remove new version
  state is_healthy <<choice>>
  wait --> check
  check --> is_healthy
  is_healthy --> replace : is healthy
  is_healthy --> not_healthy
  replace --> [*]
  not_healthy --> [*]
```

## LICENSE

[License](LICENSE)
