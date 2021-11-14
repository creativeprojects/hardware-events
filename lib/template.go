package lib

import (
	"path/filepath"

	"github.com/creativeprojects/hardware-events/cfg"
)

type Template struct {
	global *Global
	Name   string
	ID     string
}

func NewTemplate(global *Global, name string, config cfg.Template) *Template {
	return &Template{
		global: global,
		Name:   name,
		ID:     filepath.Base(config.Source),
	}
}
