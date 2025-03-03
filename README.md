# Pomodoro CLI

A beautiful terminal-based Pomodoro timer and task tracker built with Go and [Charm libraries](https://github.com/charmbracelet).

![Pomodoro CLI Screenshot](placeholder-for-screenshot.png)

## Features

- 🍅 Pomodoro timer with focus, short break, and long break modes
- 📋 Task management with completion tracking
- 🏆 Pomodoro count tracking per task
- ⌨️ Keyboard-driven interface
- 🎨 Beautiful terminal UI inspired by modern design patterns

## Installation

### Prerequisites

- Go 1.16 or higher

### Building from source

```bash
go install github.com/jackrudenko/pomodorocli/cmd/pomodorocli@latest
```

Or clone the repository:

```bash
git clone https://github.com/jackrudenko/pomodorocli.git
cd pomodorocli
go build -o pomodorocli ./cmd/pomodorocli
```

## Usage

Run the application:

```bash
./pomodorocli
```

### Keyboard Controls

#### Main View

- `q` / `Ctrl+C` - Quit the application
- `s` - Start/Stop the timer
- `p` - Pause/Resume the timer
- `n` - Add a new task
- `h` - Toggle show/hide completed tasks
- `j` / `down` - Move down in the task list
- `k` / `up` - Move up in the task list
- `Enter` - Select the current task and start the timer
- `Space` - Toggle the completion status of the selected task

#### Add Task View

- `Tab` - Switch between input fields
- `Enter` - Add the task
- `Esc` - Cancel and return to the main view

## Pomodoro Technique

The Pomodoro Technique is a time management method developed by Francesco Cirillo in the late 1980s. It uses a timer to break work into intervals, traditionally 25 minutes in length, separated by short breaks. 

1. Choose a task to work on
2. Start the timer (typically 25 minutes)
3. Work on the task until the timer rings
4. Take a short break (5 minutes)
5. After 4 pomodoros, take a longer break (15-30 minutes)

This application helps you follow this technique with a beautiful interface right in your terminal.

## License

MIT

## Credits

- UI design inspired by [Figma design](https://www.figma.com/design/5AywSgXKEmmz51cTdKkZVW/Pomodoro-Tracker)
- Built with [Charm libraries](https://github.com/charmbracelet)
  - [Bubble Tea](https://github.com/charmbracelet/bubbletea)
  - [Lipgloss](https://github.com/charmbracelet/lipgloss)
  - [Bubbles](https://github.com/charmbracelet/bubbles)
