```markdown
# Battle Shell 🚀

A terminal-based UI to organize and launch your favorite bash commands with menus, sub-menus and dynamic arguments!

---

## 🔥 Features

- 🎛️ **Configurable menus & sub-menus** via `config.json`  
- ⚙️ **Setup** screen to customize your shell path (`bashPath`)  
- 📝 **Placeholders** `{arg}` prompt you for values before running  
- 🧹 **Auto-clear** output panel before each run  
- ❌ **Exit** button (or `q`) to quit cleanly  
- Automatic generation of `config.json` with default menu

---

## 📦 Installation

1. Clone the repo:  
   ```bash
   git clone https://your-repo.git
   cd battle-shell/src
   ```
2. Get dependencies:  
   ```bash
   go get github.com/rivo/tview
   ```
3. Build and run:  
   ```bash
   go build -o battle-shell main.go config.go
   ./battle-shell
   ```

---

## ⚙️ Configuration

On first run, `config.json` is created:
```json
{
  "menu": {
    "title": "Battle Shell",
    "commands": [
      { "name": "List Files", "description": "List files", "command": "ls -l" },
      …
    ],
    "submenus": [
      {
        "title": "Network Tools",
        "commands": [ … ]
      },
      …
    ]
  }
}
```
- **Edit** `config.json` to add or remove commands and sub-menus.  
- **Restart** the app to load changes.

---

## 🚀 Usage

- **↑/↓** or **j/k** to navigate  
- **Enter** to select  
- **Tab** to focus the output panel  
- If a command contains `{arg}`, a form appears to fill in values  
- Otherwise it runs immediately  
- **❌ Exit** or press `q` to quit  
- **⚙ Setup** to change the shell path (e.g. `/bin/zsh`)

---

## 💡 Example

Add a command with an argument:
```json
{ 
  "name": "Grep Logs", 
  "description": "Search logs for pattern", 
  "command": "grep '{pattern}' *.log" 
}
```
Select it, enter `pattern`, and see the filtered output.

---
