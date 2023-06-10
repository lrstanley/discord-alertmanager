## :gear: Usage

#### Application Options
| Environment vars | Flags | Type | Description |
| --- | --- | --- | --- |
| - | `-v, --version` | bool | prints version information and exits |
| - | `--version-json` | bool | prints version information in JSON format and exits |
| `DEBUG` | `-D, --debug` | bool | enables debug mode |

#### Discord Options
| Environment vars | Flags | Type | Description |
| --- | --- | --- | --- |
| `DISCORD_TOKEN` | `--discord.token` | string | Discord bot token [**required**] |

#### Alertmanager Options
| Environment vars | Flags | Type | Description |
| --- | --- | --- | --- |
| `ALERTMANAGER_URL` | `--alertmanager.url` | string | Alertmanager URL [**required**] |
| `ALERTMANAGER_USERNAME` | `--alertmanager.username` | string | Alertmanager username (if configured) |
| `ALERTMANAGER_PASSWORD` | `--alertmanager.password` | string | Alertmanager password (if configured) |

#### Logging Options
| Environment vars | Flags | Type | Description |
| --- | --- | --- | --- |
| `LOG_QUIET` | `--log.quiet` | bool | disable logging to stdout (also: see levels) |
| `LOG_LEVEL` | `--log.level` | string | logging level [**default: info**] [**choices: debug, info, warn, error, fatal**] |
| `LOG_JSON` | `--log.json` | bool | output logs in JSON format |
| `LOG_PRETTY` | `--log.pretty` | bool | output logs in a pretty colored format (cannot be easily parsed) |
| `LOG_PATH` | `--log.path` | string | path to log file (disables stdout logging) |
