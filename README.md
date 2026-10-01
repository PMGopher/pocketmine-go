<p align="center">
	<a href="https://github.com/PMGopher/pocketmine-go">
		<picture>
			<source srcset=".github/readme/logo-dark.svg" media="(prefers-color-scheme: dark)">
			<img src=".github/readme/logo-light.svg" alt="PocketMine-go" width="560" loading="eager" />
		</picture>
	</a><br>
	<b>A highly customisable, open source server software for Minecraft: Bedrock Edition, written in Go</b>
</p>

<p align="center">
	<a href="go.mod"><img alt="Go 1.26+" src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white"></a>
	<img alt="Minecraft: Bedrock Edition 1.26.50" src="https://img.shields.io/badge/Bedrock-1.26.50-62B47A">
	<a href="LICENSE"><img alt="License: LGPL-3.0" src="https://img.shields.io/badge/license-LGPL--3.0-blue"></a>
	<a href="https://github.com/PMGopher/pocketmine-go/stargazers"><img alt="GitHub stars" src="https://img.shields.io/github/stars/PMGopher/pocketmine-go?style=flat&logo=github"></a>
</p>

## What is this?
PocketMine-go is a server for Minecraft: Bedrock Edition, built in Go. It brings the gameplay,
plugin API and server tooling of **PocketMine-MP** to a fast, single-binary Go server.

If you want a Bedrock server with **custom functionality**, you're in the right place.

- 🧩 **Plugin API in plain Go**: events, commands, permissions, configs and a task scheduler, with
  the compiler checking your types before your players ever find a bug
- 🗺️ **Multi-world support**: load several worlds at once, and drop a world from vanilla Bedrock
  straight into `worlds/`
- 🏎️ **Built for performance**: chunk generation, population and lighting run on worker goroutines,
  so the main thread keeps ticking at 20 TPS while players explore
- 📦 **One binary, no runtime**: build once and run anywhere Go runs: Windows, Linux or macOS,
  on x86 or ARM
- 🎮 **Up to date**: talks to Bedrock **1.26.50** clients through
  [gophertunnel](https://github.com/sandertv/gophertunnel), with Xbox Live authentication on by
  default
- 🧱 **Familiar behaviour**: blocks, items, entities, crafting, enchanting and world generation
  work the way PocketMine-MP server owners know

## :x: PocketMine-go is NOT a vanilla Minecraft server software.
**It is poorly suited to hosting vanilla survival servers.**
Like PocketMine-MP, it doesn't have many features from the vanilla game, such as vanilla world
generation, redstone circuits and mob AI.

If you just want to play **vanilla survival multiplayer**, consider using the
[official Minecraft: Bedrock server software](https://minecraft.net/download/server/bedrock)
instead.

If that's not an option for you, you can add the features you're missing yourself with
[plugins](#-writing-plugins).

## 🚀 Getting started
You need [Go 1.26 or newer](https://go.dev/dl/).

```bash
git clone https://github.com/PMGopher/pocketmine-go.git
cd pocketmine-go
go build -o pocketmine-go ./cmd/pocketmine-go     # on Windows: -o pocketmine-go.exe
./pocketmine-go
```

The first start asks a few questions (language, server name, port, ...) and creates your
configuration. Then open Minecraft, go to **Servers → Add Server** and enter your machine's IP
and port `19132`.

Type `help` in the console to see the commands, and `stop` (or press Ctrl+C) to shut down. Worlds
and players are saved on the way out.

### Building for another system
Go cross-compiles out of the box. For example, to build a Windows server from Linux or macOS:

```bash
GOOS=windows GOARCH=amd64 go build -o pocketmine-go.exe ./cmd/pocketmine-go
GOOS=linux   GOARCH=arm64 go build -o pocketmine-go     ./cmd/pocketmine-go   # e.g. a Raspberry Pi
```

### Command-line options

| Option | What it does |
|---|---|
| `--data=<path>` | Folder for configs, worlds and players (default: the current folder) |
| `--plugins=<path>` | Plugins folder (default: `plugins` in the current folder) |
| `--no-wizard` | Skip the first-start questions and use the defaults |
| `--no-log-file` | Don't write `server.log` |
| `--enable-ansi` / `--disable-ansi` | Force coloured console output on or off |
| `--version` | Print the version and exit |
| `--<key>=<value>` | Override any `server.properties` or `pocketmine.toml` setting, e.g. `--server-port=19133`, `--xbox-auth=false`, `--debug.level=2` |

### The data folder

```text
<data>/
├── server.properties        port, MOTD, game mode, difficulty, view distance, xbox-auth, ...
├── pocketmine.toml          advanced settings: worlds, chunk sending, aliases, timings, ...
├── ops.txt  white-list.txt  banned-players.txt  banned-ips.txt
├── worlds/<name>/           level.dat + db/, the same layout as vanilla Bedrock
├── players/<name>.dat       inventories, positions, XP, ...
├── plugin_data/<plugin>/    each plugin's own files (config.toml, ...)
├── plugin_list.toml         allow or block plugins by name
├── resource_packs/          resource packs offered to players
├── crashdumps/              written if the server ever crashes
└── server.log
```

All configuration files are [TOML](https://toml.io). Upgrading from a version that used YAML?
`pocketmine.yml`, `plugin_list.yml`, `resource_packs.yml` and plugins' `config.yml` are converted
automatically on the first start, settings included; the old files are kept as `.yml.bak`.

Worlds made by vanilla Bedrock load as they are; older world formats (Anvil, McRegion, PMAnvil) are
converted automatically, with a backup in `backups/worlds/`.

## 🧩 Writing plugins
Plugins are Go packages compiled into the server. There is no plugin file to drop in a folder:
you import the package and rebuild, and the compiler checks your plugin against the server's API.

A complete, commented example lives in its own repository:
**[PMGopher/example](https://github.com/PMGopher/example)**. It has a welcome title on join, a
`/example` command, a config file, a repeating task and a test. The fastest way to start is to copy
it. Converting a plugin from PocketMine-MP? Its
[AGENTS.md](https://github.com/PMGopher/example/blob/main/AGENTS.md) is a step-by-step conversion
guide.

### 1. Describe the plugin: `plugin.toml`

```toml
name = "HelloPlugin"
version = "1.0.0"
main = "hello.Main"
api = ["5.0.0"]

[commands.hello]
description = "Says hello"
permission = "helloplugin.command.hello"

[permissions."helloplugin.command.hello"]
default = true
```

### 2. Write it: `hello.go`

```go
package hello

import (
	"embed"

	"pocketmine-go/pocketmine/command"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/plugin"
	"pocketmine-go/pocketmine/server"
)

//go:embed plugin.toml
var files embed.FS

// Register the plugin when its package is imported.
func init() {
	plugin.RegisterGoPlugin(files, func() plugin.Plugin { return &Main{} })
}

type Main struct{ plugin.PluginBase }

func (m *Main) OnEnable() error {
	m.GetLogger().Info("Hello, world!")
	srv := m.GetServer().(*server.Server)
	return srv.GetPluginManager().RegisterEvents(&listener{m}, m)
}

// Commands from plugin.toml land here.
func (m *Main) OnCommand(sender command.Sender, cmd command.CommandLike, label string, args []string) bool {
	sender.SendMessage("Hello, " + sender.GetName() + "!")
	return true
}

// Every exported method that takes one event is an event handler.
type listener struct{ m *Main }

func (l *listener) OnJoin(e *playerevent.PlayerJoinEvent) {
	l.m.GetLogger().Info(e.GetPlayer().GetName() + " joined")
}
```

To ship default files (like a `config.toml`), put them in a `resources/` folder next to
`plugin.toml`, embed it too (`//go:embed plugin.toml resources`) and call `m.SaveDefaultConfig()`.

Your plugin is a Go module of its own (`go mod init github.com/you/helloplugin`). Its `go.mod`
needs no `require` for the server; a `go.work` next to it points at your server clone while you
develop (see the example plugin).

### 3. Register it: `cmd/pocketmine-go/plugins.go`

In your clone of the server, fetch the plugin and add a blank import of it, then rebuild:

```bash
go get github.com/PMGopher/example@latest
```

```go
package main

import (
	_ "github.com/PMGopher/example"   // the example plugin
	_ "github.com/you/helloplugin"    // your own plugin
)
```

On the next start the console shows `Loading HelloPlugin v1.0.0` and `Enabling HelloPlugin v1.0.0`.
The server checks the `api` version, `depend`/`softdepend`/`loadbefore`, `load` order and
`plugin_list.toml` before enabling it. Your plugin gets its own folder in `plugin_data/`.

### What a plugin can use

| You want to... | Use |
|---|---|
| React to something happening | `RegisterEvents(listener, plugin)`, or `plugin.RegisterEvent[E]` for a single handler. Events live in `pocketmine/event/...` (player, block, entity, inventory, world, server) |
| Add a command | Declare it in `plugin.toml`, handle it in `OnCommand` |
| Run code later or repeatedly | `m.GetScheduler().ScheduleDelayedTask` / `ScheduleRepeatingTask` with `scheduler.NewClosureTask(func() { ... })` (20 ticks = 1 second) |
| Store settings | `m.SaveDefaultConfig()`, `m.GetConfig().Get(key, default)`, `m.ReloadConfig()` |
| Reach the server | `m.GetServer().(*server.Server)`: players, worlds, broadcasting, commands, bans, ... |
| Talk to a player | `*player.Player`: `SendMessage`, `SendTitle`, `SendTip`, `SendPopup`, `SendForm`, `Teleport`, `SetGamemode`, inventories, ... |

## 🏗️ Under the hood
- **Networking** by [gophertunnel](https://github.com/sandertv/gophertunnel): RakNet, login,
  encryption and every packet, including the sub-chunk and client blob cache paths the latest
  clients use.
- **One main thread, many goroutines**: the game ticks on one thread like PocketMine-MP, so plugin
  code never has to think about locks; world generation, population and lighting fan out to worker
  goroutines.
- **Worlds** in LevelDB, the format vanilla Bedrock uses, with block-state and item upgraders for
  worlds from older versions. Vanilla blocks that have no behaviour here yet are kept as they are,
  so they still show up and save unchanged.
- **Crash dumps** in `crashdumps/`, timings reports and a query protocol for server lists.

Want to see how much of PocketMine-MP is covered? See **[PROGRESS.md](PROGRESS.md)**.

## 🤝 Contributing
Contributions are welcome! Before you start:

- Read **[AGENTS.md](AGENTS.md)**: it covers the architecture, the code conventions and the known
  issues, for humans and AI agents alike.
- Run `go vet ./... && go test ./...` before opening a pull request.

Found a bug? [Open an issue](https://github.com/PMGopher/pocketmine-go/issues) with your server
log and, if there is one, the crash dump from `crashdumps/`.

## 🙏 Credits
- **[MEMOxiiii](https://github.com/MEMOxiiii)**: developer of PocketMine-go.
- **[gophertunnel](https://github.com/sandertv/gophertunnel)**: the Bedrock protocol implementation.

## Licensing information
PocketMine-go is licensed under LGPL-3.0. Please see the [LICENSE](LICENSE) file for details.

PocketMine-go is developed by [PMGopher](https://github.com/PMGopher) and
[MEMOxiiii](https://github.com/MEMOxiiii). It is not affiliated with Mojang or Microsoft. All brands
and trademarks belong to their respective owners. PocketMine-go is not a Mojang-approved software,
nor is it associated with Mojang.
