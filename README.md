# Radio TUI

A modern, responsive Terminal User Interface (TUI) for listening to internet radio stations. Built with Go and the [Charm](https://charm.sh/) libraries.

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![mpv](https://img.shields.io/badge/mpv-%23000000.svg?style=for-the-badge&logo=mpv&logoColor=white)

## Features

- **Global Discovery**: Search thousands of radio stations via the community-driven [Radio Browser API](https://www.radio-browser.info/).
- **Favorites Management**: Save your most-loved stations for quick access.
- **Real-time Metadata**: See the currently playing song title and artist directly in the status bar (via JSON-IPC).
- **Responsive Dashboard**: A `lazygit`-inspired grid layout that automatically collapses to a single column on smaller terminal windows.
- **Robust Playback**: Powered by `mpv` for high-quality, format-agnostic streaming.
- **Modern UI**: Styled with `lipgloss` featuring rounded borders and context-aware help menus.

## Getting Started

### Prerequisites

You must have `mpv` installed on your system to handle the audio streams.

```bash
# Ubuntu/Debian
sudo apt install mpv

# macOS (Homebrew)
brew install mpv

# Fedora
sudo dnf install mpv
```

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/radio-tui.git
   cd radio-tui
   ```

2. Run the application:
   ```bash
   go run .
   ```

3. (Optional) Build the binary:
   ```bash
   go build -o radio-tui
   ./radio-tui
   ```

## Keyboard Shortcuts

| Key | Action |
| :--- | :--- |
| `Tab` / `Shift+Tab` | Cycle focus between Search bar, Results, and Favorites |
| `Enter` (Search) | Initiate station search |
| `Enter` / `Space` (List) | Play selected station / Toggle Pause |
| `f` | Add/Remove station from Favorites |
| `Esc` | Jump focus back to Search bar |
| `q` / `Ctrl+C` | Quit and stop stream |

## Built With

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: The TUI framework.
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)**: Terminal CSS.
- **[Bubbles](https://github.com/charmbracelet/bubbles)**: Common TUI components (List, TextInput).
- **[mpv](https://mpv.io/)**: The Swiss-army knife audio engine.
