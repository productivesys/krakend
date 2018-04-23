package plugin

import (
	"fmt"
	"io/ioutil"
	"plugin"
	"strings"
)

// Plugin is the interface of the loaded plugins
type Plugin interface {
	Lookup(name string) (plugin.Symbol, error)
}

type PluginDefinition interface {
	GetFolder() string
	GetPattern() string
}

// Load loads all the plugins in pluginFolder with pattern in its filename.
// It returns the number of plugins loaded and an error if something goes wrong.
func Load(cfg PluginDefinition, reg *Register) (int, error) {
	plugins, err := scan(cfg.GetFolder(), cfg.GetPattern())
	if err != nil {
		return 0, err
	}
	return load(plugins, reg)
}

func scan(folder, pattern string) ([]string, error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return []string{}, err
	}

	plugins := []string{}
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), pattern) {
			plugins = append(plugins, folder+file.Name())
		}
	}

	return plugins, nil
}

func load(plugins []string, reg *Register) (int, error) {
	errors := []error{}
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, reg); err != nil {
			errors = append(errors, fmt.Errorf("opening plugin %d (%s): %s", k, pluginName, err.Error()))
			continue
		}
		loadedPlugins++
	}

	if len(errors) > 0 {
		return loadedPlugins, loaderError{errors}
	}
	return loadedPlugins, nil
}

func open(pluginName string, reg *Register) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	var p Plugin
	p, err = pluginOpener(pluginName)
	if err != nil {
		return
	}
	err = reg.Register(p)
	return
}

// pluginOpener keeps the plugin open function in a var for easy testing
var pluginOpener = defaultPluginOpener

func defaultPluginOpener(name string) (Plugin, error) {
	return plugin.Open(name)
}

type loaderError struct {
	errors []error
}

// Error implements the error interface
func (l loaderError) Error() string {
	msgs := make([]string, len(l.errors))
	for i, err := range l.errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("plugin loader found %d error(s): \n%s", len(msgs), strings.Join(msgs, "\n"))
}
