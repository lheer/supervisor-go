# supervisor-go

supervisor-go is an application that allows to control Unix processes. Think of it as a very simplified [supervisord](https://github.com/Supervisor/supervisor).
It's configured via a simple TOML file. View the [examples](examples) folder for configuration examples.

## Features

- Start and monitor Unix processes
- Restart on failure, max restart limits
- Execution order: Start a process after another is running or has exited

## Usage

```console
$ supervisor-go -c <path_to_config_file.toml>
```
