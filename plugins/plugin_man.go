package plugins

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/gofrs/uuid"

	"github.com/robertkrimen/otto"
)

var manager = &Manager{
	dir:     "./plugins",
	plugins: []Plugin{},
}

type Plugin struct {
	uuid     string
	filePath string
	src      string
	vm       *otto.Otto
}

func (p *Plugin) UUID() string { return p.uuid }

func (p *Plugin) ParseFile() error {
	if p.filePath != "" && p.filePath != "-" {
		data, err := ioutil.ReadFile(p.filePath)
		if err != nil {
			return err
		}
		p.src = string(data)
	}
	return nil
}

func (p *Plugin) Call(funcName string, this interface{}, argumentList ...interface{}) (otto.Value, error) {
	if err := p.ParseFile(); err != nil {
		return otto.Value{}, err
	}

	if _, err := p.vm.Run(p.src); err != nil {
		return otto.Value{}, err
	}

	return p.vm.Call(funcName, this, argumentList)
}

// Manager contains plugin collection and add utility and concurrent protection for executing
type Manager struct {
	sync.Mutex
	dir     string
	plugins []Plugin
}

// NewManager retrieves pointer to only single instance plugin manager
func NewManager() *Manager {
	return manager
}

// Load finds all plugins in provided directory and loads then into manager
func (m *Manager) Load() error {
	m.Unload()

	if err := m.loadFromDir(m.dir); err != nil {
		return err
	}

	return nil
}

func (m *Manager) Unload() {
	m.Lock()
	defer m.Unlock()
	m.plugins = []Plugin{}
}

func (m *Manager) Plugins() *[]Plugin {
	return &m.plugins
}

func (m *Manager) loadFromDir(dir string) error {
	if files, err := ioutil.ReadDir(m.dir); err == nil {
		for i := range files {
			file := files[i]
			fileFullPath := fmt.Sprintf("%s%s%s", dir, string(os.PathSeparator), file.Name())
			// if found directory, call this function to process that directory too
			if file.IsDir() {
				m.loadFromDir(fileFullPath)
			}
			fileNameParts := strings.Split(file.Name(), ".")
			if len(fileNameParts) > 1 {
				if fileNameParts[len(fileNameParts)-1] == "js" {
					m.loadPlugin(fileFullPath)
				}
			}
		}
	} else {
		return err
	}
	return nil
}

func (m *Manager) loadPlugin(fileFullPath string) error {
	m.Lock()
	defer m.Unlock()

	if uuidV4, err := uuid.NewV4(); err == nil {
		plugin := Plugin{
			uuid:     uuidV4.String(),
			vm:       otto.New(),
			filePath: fileFullPath,
		}

		if err := plugin.ParseFile(); err != nil {
			return err
		}

		plugin.vm.Set("UUID", plugin.uuid)
		plugin.vm.Set("InfoLog", PluginInfoLog)
		plugin.vm.Set("DebugLog", PluginDebugLog)
		plugin.vm.Set("ErrorLog", PluginErrorLog)
		plugin.vm.Run(plugin.src)

		m.plugins = append(m.plugins, plugin)
	} else {
		return err
	}

	return nil
}
