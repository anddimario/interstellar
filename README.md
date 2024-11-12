# Interstellar

Application deployer

## Features

- monitor and get release from github
- run executable
- blue green and canary deploy
- rollback
- recovery from crash
- cli

## Install and requirements

- Deps: `gh` (github cli)

## CLI

```bash
interstellar -h
```

### Dev

```bash
go run main.go -h
```

## Diagrams

### New release Check

```mermaid
stateDiagram-v2
  check_ignore: Check if ignore
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
  new_release --> check_ignore
  check_ignore --> download
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
  post_deploy: Post deploy actions
  state is_healthy <<choice>>
  check --> is_healthy
  is_healthy --> add : is healthy
  is_healthy --> not_healthy
  add --> wait
  wait --> remove_old
  not_healthy --> [*]
  remove_old --> post_deploy
  post_deploy --> [*]
```

### Blue green deploy

```mermaid
stateDiagram-v2
  wait: Wait positive healthchecks
  check: Check if healthy
  replace: Replace old with new
  not_healthy: Remove new version
  post_deploy: Post deploy actions
  state is_healthy <<choice>>
  wait --> check
  check --> is_healthy
  is_healthy --> replace : is healthy
  is_healthy --> not_healthy
  not_healthy --> [*]
  replace --> post_deploy
  post_deploy --> [*]
```

### Rollback

```mermaid
stateDiagram-v2
  check: Check if deploy in progress, or version exists
  state if_check_ok <<choice>>
  get_release: Get Release
  decompress: Decompress Release
  start: Start the rollback version
  remove: Remove the actual version
  update_config: Update config
  check --> if_check_ok
  if_check_ok --> get_release : Check ok
  if_check_ok --> [*]
  get_release --> decompress
  decompress --> start
  start --> remove
  remove --> update_config
  update_config --> [*]
```

### Recovery from crash

```mermaid
stateDiagram-v2
  start: Startup
  check_deploy: Check if deploy in progress
  state if_check_ok <<choice>>
  kill: Kill other version processes
  start --> check_deploy
  check_deploy --> if_check_ok
  if_check_ok --> kill : Zombie deploy
  kill --> [*]
  if_check_ok --> [*]
```

**NOTE** The healthcheck will remove the backend from the backends list

## LICENSE

[License](LICENSE)
