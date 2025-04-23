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
					{"List Files", "List directory", "ls -l"},
					{"Sys Info", "System info", "uname -a"},
					{"Net Config", "Network config", "ip a"},
				},
				Submenus: []MenuConfig{
					{Title: "Network Tools", Commands: []CommandConfig{{"Ping Google", "ping google.com", "ping -c3 google.com"}}},
				},
			},
			TailFiles: []string{},
			Processes: []ProcessConfig{},
		}
		data, _ := json.MarshalIndent(defaultCfg, "", "  ")
		ioutil.WriteFile(path, data, 0644)
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

