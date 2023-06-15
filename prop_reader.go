package statepro

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var stateProYml = "statepro.yml"
var changePropPathOnce sync.Once

// SetDefinitionPath sets the path to the configuration yml file.
// This method should be called before InitMachines.
func SetDefinitionPath(path string) {
	changePropPathOnce.Do(func() {
		stateProYml = path
	})
}

// Props is the statepro wrapper configuration filled from the yml file.
type Props struct {
	*StateProProp `yaml:"statepro"`
}

// StateProProp is the statepro configuration filled from the yml file.
type StateProProp struct {
	FilePrefix *string  `yaml:"file-prefix"`
	Paths      []string `yaml:"paths"`
}

func (p *Props) getPrefix() string {
	if p.FilePrefix == nil {
		return ""
	}
	return *p.FilePrefix
}

func loadProp() *Props {
	p := &Props{}
	arr, exist := readYml()

	// if the file does not exist, return nil
	if !exist {
		log.Printf("[WARN] statepro yml file '%s' does not exist", stateProYml)
		return nil
	}

	// if the file exists, parse it
	env := []byte(os.ExpandEnv(string(arr)))
	err := yaml.Unmarshal(env, p)
	if err != nil {
		log.Fatalf("[FATAL] Error parsing statepro yml file '%s': %s", stateProYml, err)
	}
	return p
}

func readYml() ([]byte, bool) {
	// check if the file exists
	_, err := os.Stat(stateProYml)
	if os.IsNotExist(err) {
		return nil, false
	}

	filename, err := filepath.Abs(stateProYml)
	if err != nil {
		log.Fatalf("[FATAL] Error getting statepro yml file '%s'. %s", filename, err)
	}

	b, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("[FATAL] Error reading statepro yml file '%s'. %s", filename, err)
	}
	return b, true
}
