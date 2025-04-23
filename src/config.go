package main

import (
    "encoding/json"
    "io/ioutil"
    "os"
)

// CommandConfig représente une commande bash.
type CommandConfig struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Command     string `json:"command"`
}

// MenuConfig représente un menu (avec ses sous-menus).
type MenuConfig struct {
    Title    string         `json:"title"`
    Commands []CommandConfig `json:"commands"`
    Submenus []MenuConfig    `json:"submenus"`
}

// AppConfig enveloppe le menu racine et les fichiers à tail.
type AppConfig struct {
    Menu      MenuConfig  `json:"menu"`
    TailFiles []string    `json:"tail_files"`
}

// LoadConfig charge config.json s’il existe, sinon l’écrit avec la structure par défaut.
func LoadConfig(path string) (*AppConfig, error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        // création de la config par défaut
        defaultCfg := &AppConfig{
            Menu: MenuConfig{
                Title: "Battle Shell",
                Commands: []CommandConfig{
                    {"List Files", "Liste les fichiers", "ls -l"},
                    {"Sys Info", "Infos système", "uname -a"},
                    {"Net Config", "Configuration réseau", "ip a"},
                },
                Submenus: []MenuConfig{
                    {
                        Title: "Network Tools",
                        Commands: []CommandConfig{
                            {"Ping Google", "ping google.com", "ping -c3 google.com"},
                            {"Check Port 80", "nc -zv localhost 80", "nc -zv localhost 80"},
                        },
                    },
                    {
                        Title: "File Ops",
                        Commands: []CommandConfig{
                            {"Temp File", "crée un fichier temporaire", "touch /tmp/temp_file"},
                            {"Last Log", "affiche les 10 dernières lignes du syslog", "tail -n10 /var/log/syslog"},
                        },
                    },
                },
            },
            TailFiles: []string{},
        }
        data, err := json.MarshalIndent(defaultCfg, "", "  ")
        if err != nil {
            return nil, err
        }
        if err := ioutil.WriteFile(path, data, 0644); err != nil {
            return nil, err
        }
        return defaultCfg, nil
    }

    // si config.json existe, on la lit et on retourne exactement ce qui y est (menus + sous-menus + tail_files)
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg AppConfig
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
