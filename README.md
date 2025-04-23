# Battle Shell 🚀

A terminal-based TUI to organize and launch your favorite bash commands with:

- configurable menus & sub-menus  
- dynamic argument prompts  
- built-in shell input  
- file-tail viewer  
- background process watchers  

---

## 🔥 Features

- 🎛️ **Menus & Sub-menus** defined in `config.json`  
- ⚙️ **Setup** screen to change your shell path (`bashPath`)  
- 📝 **Named placeholders** `{arg}` prompt you for values before execution  
- 💻 **Free shell input** panel to run any bash command  
- 🧹 **Auto-clear** output before each run  
- 📜 **Tail view** to follow log files with live refresh  
- ⏱️ **Process watchers** that run commands periodically in background, color-coded by status  
- 🔢 **Auto-assigned shortcuts** `1…9,0` (AZERTY-aware) for menu items  
- ❌ **Exit** button (or `q`) to quit cleanly  
- Automatic generation of a default `config.json` on first run  

---

## 📦 Installation

1. Clone the repo:  
   ```bash
   git clone https://github.com/your-org/battle-shell.git
   cd battle-shell
   ```
2. Build:  
   ```bash
   go get ./...
   go build -o battleshell.bin main.go
   ```
3. Run:  
   ```bash
   ./battleshell.bin
   ```

---

## ⚙️ Configuration

On first run, `config.json` is created with defaults:

```json
{
  "menu": {
    "title": "Battle Shell",
    "commands": [
      { "name": "List Files", "description": "List files", "command": "ls -l" },
      { "name": "Sys Info",  "description": "System info", "command": "uname -a" }
    ],
    "submenus": [
      {
        "title": "Network Tools",
        "commands": [
          { "name": "Ping Google", "description": "ping google.com", "command": "ping -c3 google.com" }
        ]
      }
    ]
  },
  "tail_files": [],
  "processes": []
}
```

- **Edit** `config.json` to add/remove `commands`, `submenus`, `tail_files`, or `processes`.  
- **Restart** the app to apply changes.

---

## 🚀 Usage

- **↑/↓** or **j/k** to navigate menus  
- **1…9,0** to trigger items via auto-shortcuts (works on AZERTY top row)  
- **Enter** to select  
- **Tab** to cycle focus (menu ↔ output ↔ shell input)  
- **Type** in the shell input panel to run arbitrary bash commands  
- **Switch tabs** with **F1** (Main), **F2** (Tail), **F3** (Procs)  
- In **Tail** view, select a file to follow its last lines, auto-refreshing every 2s  
- In **Procs** view, watchers run in background; names turn green/red on success/failure; select to see last output  
- **❌ Exit** or press **q** to quit  

---

## 💡 Example

Add a command with an argument:

```json
{
  "menu": {
    "commands": [
      {
        "name": "Search Logs",
        "description": "grep pattern in logs",
        "command": "grep '{pattern}' /var/log/*.log"
      }
    ]
  }
}
```

- Select **Search Logs**, enter `pattern`, and view results in the output panel.
