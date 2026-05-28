package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
)

type AudioItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Duration float64 `json:"duration"`
	Format   string  `json:"format"`
	GroupID  string  `json:"group_id"`
}

type Group struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Items []*AudioItem `json:"-"` // loaded from config
}
type ChatWheel struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Slots  []*AudioItem `json:"slots"`
	Hotkey string       `json:"hotkey"` // hotkey to show the wheel
}

type Config struct {
	Groups     []*Group     `json:"groups"`
	AudioItems []*AudioItem `json:"audio_items"`
	ChatWheels *ChatWheel   `json:"chat_wheels"`
	Volume     float64      `json:"volume"` // stored as 0-100 for slider binding
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(homeDir, ".oneclick", "config.json")

	var cfg Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg = *createDefaultConfig()
		// Save the newly created config with preloaded sounds
		cfg.Save()
	} else {
		data, err := os.ReadFile(configPath)
		if err != nil {
			cfg = *createDefaultConfig()
			cfg.Save()
		} else if err := json.Unmarshal(data, &cfg); err != nil {
			cfg = *createDefaultConfig()
			cfg.Save()
		} else {
			// Load any new files from sounds folder even when config exists
			cfg.preloadSoundsFolder()
			// Ensure active chat wheel has 9 slots for 3x3 grid
			cfg.ensureSlots()
			// Save updated config
			cfg.Save()
		}
	}

	return &cfg, nil
}

// ensureSlots ensures the active chat wheel has 9 slots for 3x3 grid
func (c *Config) ensureSlots() {
	wheel := c.GetActiveChatWheel()
	if wheel != nil {
		for len(wheel.Slots) < 8 {
			wheel.Slots = append(wheel.Slots, nil)
		}
	}
}

func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(homeDir, ".oneclick")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func createDefaultConfig() *Config {
	cfg := &Config{
		Groups:     []*Group{},
		AudioItems: []*AudioItem{},
		ChatWheels: &ChatWheel{ID: "wheel1", Name: "轮盘", Hotkey: "p", Slots: make([]*AudioItem, 9)},
		Volume:     100, // default 100% volume stored as 0-100
	}

	// Preload audio from sounds folder in project root
	cfg.preloadSoundsFolder()

	return cfg
}

// preloadSoundsFolder loads audio files from sounds folder
// First-level subdirectories in sounds are groups, directory name = group name
// Files inside subdirectories are audio files in that group
func (c *Config) preloadSoundsFolder() {
	// Check if sounds folder exists
	soundsDir := "sounds"
	if _, err := os.Stat(soundsDir); os.IsNotExist(err) {
		return
	}

	entries, err := os.ReadDir(soundsDir)
	if err != nil {
		return
	}

	supportedExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true}

	// Clear existing groups and audio to sync with filesystem
	// Since we don't support dynamic editing, just reload from disk every time
	c.Groups = nil
	c.AudioItems = nil

	for _, entry := range entries {
		if entry.IsDir() {
			// Subdirectory is a new group
			groupName := entry.Name()
			group := &Group{
				ID:   generateID(),
				Name: groupName,
			}
			c.Groups = append(c.Groups, group)

			// Read all files in this group directory
			groupPath := filepath.Join(soundsDir, groupName)
			files, err := os.ReadDir(groupPath)
			if err != nil {
				continue
			}

			for _, f := range files {
				if f.IsDir() {
					continue
				}
				name := f.Name()
				ext := strings.ToLower(filepath.Ext(name))
				if !supportedExts[ext] {
					continue
				}

				path := filepath.Join(groupPath, name)
				item := &AudioItem{
					ID:       generateID(),
					Name:     strings.TrimSuffix(name, ext),
					Path:     path,
					Duration: 0,
					Format:   strings.TrimPrefix(ext, "."),
					GroupID:  group.ID,
				}
				c.AddAudio(item)
			}
		}
	}
}

func (c *Config) GetGroupByID(id string) *Group {
	for _, g := range c.Groups {
		if g.ID == id {
			return g
		}
	}
	return nil
}

func (c *Config) GetAudioByID(id string) *AudioItem {
	for _, a := range c.AudioItems {
		if a.ID == id {
			return a
		}
	}
	return nil
}

func (c *Config) GetActiveChatWheel() *ChatWheel {
	return c.ChatWheels
}

func (c *Config) GetAudioByGroup(groupID string) []*AudioItem {
	var items []*AudioItem
	for _, a := range c.AudioItems {
		if a.GroupID == groupID {
			items = append(items, a)
		}
	}
	return items
}

func (c *Config) AddAudio(item *AudioItem) {
	c.AudioItems = append(c.AudioItems, item)
}

func (c *Config) SetChatWheelSlot(pos int, audioItem *AudioItem) {
	w := c.GetActiveChatWheel()
	if w == nil {
		return
	}
	w.Slots[pos] = audioItem
}

var id int64

func generateID() string {
	// Simple ID generation
	newid := atomic.AddInt64(&id, 1)
	return strconv.FormatInt(newid, 10)
}
