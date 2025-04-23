package battleshell

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// SelectActionConfig décrit une action paramétrée sur la ligne sélectionnée.
type SelectActionConfig struct {
	Name        string `json:"name"`                  // libellé dans la modale
	Description string `json:"description,omitempty"` // info secondaire (facultatif)
	Template    string `json:"template"`              // ex. "docker logs -f {ID}"
}

// CommandConfig représente une commande bash dans le menu principal.
type CommandConfig struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Command       string               `json:"command"`
	Fields        []string             `json:"fields,omitempty"`         // noms de colonnes pour {ID},{Image},…
	SelectActions []SelectActionConfig `json:"select_actions,omitempty"` // zéro ou plusieurs
}

// MenuConfig représente un menu (avec ses sous-menus).
type MenuConfig struct {
	Title    string         `json:"title"`
	Commands []CommandConfig `json:"commands"`
	Submenus []MenuConfig    `json:"submenus"`
}

// ProcessConfig pour suivi de processus préprogrammés.
type ProcessConfig struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	Interval int    `json:"interval_seconds"`
}

// AppConfig enveloppe tout.
type AppConfig struct {
	Menu      MenuConfig      `json:"menu"`
	TailFiles []string        `json:"tail_files"`
	Processes []ProcessConfig `json:"processes"`
}

// LoadConfig charge ou crée config.json.
func LoadConfig(path string) (*AppConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultCfg := &AppConfig{
			Menu: MenuConfig{
				Title: "Battle Shell",
				Commands: []CommandConfig{
					{
						Name:        "List Files",
						Description: "List directory",
						Command:     "ls -l",
					},
					{
						Name:        "Sys Info",
						Description: "System info",
						Command:     "uname -a",
					},
					{
						Name:        "Net Config",
						Description: "Network config",
						Command:     "ip a",
					},
				},
				Submenus: []MenuConfig{
					{
						Title: "Network Tools",
						Commands: []CommandConfig{
							{
								Name:        "Ping Google",
								Description: "ping google.com",
								Command:     "ping -c3 google.com",
							},
						},
					},
				},
			},
			TailFiles: []string{},
			Processes: []ProcessConfig{},
		}
		data, _ := json.MarshalIndent(defaultCfg, "", "  ")
		_ = ioutil.WriteFile(path, data, 0644)
		return defaultCfg, nil
	}

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
