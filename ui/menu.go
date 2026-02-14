package ui

import (
	"lazypackage/node"
	"sort"
)

// InitialModel initializes the state of the application
func InitialModel() Model {
	cm := node.NewConfigManager()
	scripts, _ := cm.LoadScripts()

	mainStage := Stage{
		Title:    "Main Menu",
		Subtitle: "Available npm scripts:",
		Options:  []Option{},
		Cursor:   0,
	}

	// Sort scripts by name for consistent display
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].Name < scripts[j].Name
	})

	for _, s := range scripts {
		mainStage.Options = append(mainStage.Options, Option{
			Name:        s.Name,
			Icon:        s.Icon,
			Description: s.Command,
			Script:      s.Name,
		})
	}

	return Model{
		lines:  []string{},
		stack:  []Stage{mainStage},
		screen: "menu",
	}
}
