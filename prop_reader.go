package statepro

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var defaultStateProYml = "statepro.yml"
var changePropPathOnce sync.Once

// SetDefinitionPath sets the path to the configuration yml file.
// This method should be called before InitMachines.
func SetDefinitionPath(path string) {
	changePropPathOnce.Do(func() {
		defaultStateProYml = path
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
	arr := readYml()
	env := []byte(os.ExpandEnv(string(arr)))
	err := yaml.Unmarshal(env, p)
	if err != nil {
		log.Fatalf("Error parsing statepro yml file '%s': %s", defaultStateProYml, err)
	}
	return p
}

func readYml() []byte {
	filename, err := filepath.Abs(defaultStateProYml)
	if err != nil {
		log.Fatalf("Error getting statepro yml file '%s'. %s", filename, err)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading statepro yml file '%s'. %s", filename, err)
	}
	return b
}
