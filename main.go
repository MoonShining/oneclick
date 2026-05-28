package main

import (
	"os"
	"path/filepath"
	"voicepack/internal/audio"
	"voicepack/internal/config"
	"voicepack/internal/ui"

	"fyne.io/fyne/v2/app"
)

// 使用项目目录下的 ali.ttf 中文字体
func init() {
	// Check current directory for ali.ttf
	if _, err := os.Stat("ali.ttf"); err == nil {
		absPath, _ := filepath.Abs("ali.ttf")
		os.Setenv("FYNE_FONT", absPath)
	}
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Create view model with MVVM bindings and auto-save
	vm := config.NewViewModel(cfg)

	// Initialize audio player with volume from config
	player, err := audio.NewPlayer(cfg.Volume)
	if err != nil {
		panic(err)
	}

	// Create Fyne app with Chinese theme
	a := app.NewWithID("com.oneclick.voicepack")

	w := ui.NewMainWindow(a, vm, player, func() {
		player.Close()
		a.Quit()
	})

	// Show window and run
	w.Show()
	a.Run()
}
