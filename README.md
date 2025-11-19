# Lumos
A simple command-line tool that lets you quickly adjust your screen settings without digging through Windows menus. Control night mode, screen brightness, and HDR with one command.

## What It Does
- 🌙 Night Mode - Easily enable or disable Windows' blue light filter for better sleep
- 💡 Gamma Control - Adjust screen brightness and contrast (0-100%)
- 🎬 HDR Toggle - Quickly turn HDR on or off for supported displays
- 🚀 Simple & Fast - One command does it all, no complicated settings to navigate

## Installation

### Download Pre-built Binary
1. Go to the [Releases page](https://github.com/jipaix/lumos/releases)
2. Download the latest `lumos.exe`
3. Run it from any directory or add it to your PATH

### Build from Source
```bash
git clone https://github.com/jipaix/lumos
cd lumos
go build -o lumos.exe ./cmd/lumos
````

## Usage

```bash
lumos [--hdr on|off|toggle] [--gamma <0-100>] [--night on|off|toggle]
```

### Examples

```bash
# Enable HDR and set gamma to 75%
lumos --hdr on --gamma 75

# Toggle night light and disable HDR
lumos --night toggle --hdr off

# Set gamma only
lumos --gamma 50

# Toggle both HDR and Lumos
lumos --hdr toggle --night toggle
```

## Options

| Option      | Values          | Description              |
| ----------- | --------------- | ------------------------ |
| `--hdr`     | on, off, toggle | Control HDR              |
| `--gamma`   | 0–100           | Set gamma level          |
| `--night`   | on, off, toggle | Control Lumos      |
| `--help`    | –               | Show help message        |
| `--version` | –               | Show version information |

## Requirements

* Windows 10 or 11
* Go 1.24+ (only for building from source)

## Contributing

1. Fork the repository
2. Create a branch (`git checkout -b feature/my-change`)
3. Commit (`git commit -m "Describe change"`)
4. Push (`git push origin feature/my-change`)
5. Open a Pull Request

## License

Licensed under the MIT License. See the `LICENSE` file.

## Acknowledgments

* HDR via Windows Display Config API
* Night lights via registry configuration
* Gamma adjustment via Windows GDI APIs
