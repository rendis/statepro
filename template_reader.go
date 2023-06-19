package statepro

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const jsonExt = ".json"

var deepFileRegex = regexp.MustCompile(`^.*\*/?$`)

func getDefinitionPaths(prefix string, paths []string) []string {
	defsSet := map[string]struct{}{}

	for _, path := range paths {
		deep := false
		if deepFileRegex.MatchString(path) {
			deep = true
			path = strings.Split(path, "**")[0]
		}

		if _, err := os.Stat(path); err != nil {
			log.Fatalf("[FATAL] Cannot find machine directory/file: %s", path)
		}

		subDefs := processPath(path, prefix, deep)
		for _, def := range subDefs {
			if _, ok := defsSet[def]; !ok {
				defsSet[def] = struct{}{}
			}
		}
	}

	defPaths := make([]string, 0, len(defsSet))
	for path := range defsSet {
		defPaths = append(defPaths, path)
	}

	return defPaths
}

func processPath(path, prefix string, deep bool) []string {
	if !isDirectory(path) {
		if isDefPath(path, prefix) {
			return []string{path}
		}
		log.Fatalf("[FATAL] Invalid machine path: %s", path)
	}
	if !deep {
		return plainScan(path, prefix)
	} else {
		return deepScan(path, prefix)
	}
}

func plainScan(path, prefix string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("[FATAL] Cannot find machine directory/info: %s", path)
	}
	var defs []string
	for _, info := range files {
		if !info.IsDir() && isDefPath(info.Name(), prefix) {
			defs = append(defs, filepath.Join(path, info.Name()))
		}
	}
	return defs
}

func deepScan(path, prefix string) []string {
	var defs []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("[FATAL] An error occurred while scanning '%s': %s", path, err)
			return err
		}
		if !info.IsDir() && isDefPath(info.Name(), prefix) {
			defs = append(defs, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("[FATAL] An error occurred while scanning directory '%s': %s", path, err)
	}
	return defs
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatalf("[FATAL] Cannot find machine directory/file: %s", path)
	}
	return fileInfo.IsDir()
}

func isDefPath(path, prefix string) bool {
	fileName := filepath.Base(path)
	return strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, jsonExt)
}
