package statepro

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const stateproYml = "statepro.yml"

type Props struct {
	*Scanner `yaml:"scanner"`
}

type Scanner struct {
	FilePrefix *string  `yaml:"file-prefix"`
	Paths      []string `yaml:"paths"`
}

func (p *Props) GetPrefix() string {
	if p.FilePrefix == nil {
		return ""
	}
	return *p.FilePrefix
}

func LoadProp() *Props {
	p := &Props{}
	arr := readYml()
	env := []byte(os.ExpandEnv(string(arr)))
	err := yaml.Unmarshal(env, p)
	if err != nil {
		log.Fatalf("Error parsing statepro yml file '%s': %s", stateproYml, err)
	}
	return p
}

func readYml() []byte {
	filename, err := filepath.Abs(stateproYml)
	if err != nil {
		log.Fatalf("Error getting statepro yml file '%s'. %s", filename, err)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading statepro yml file '%s'. %s", filename, err)
	}
	return b
}
