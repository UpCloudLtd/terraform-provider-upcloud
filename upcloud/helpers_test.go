package upcloud

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// renderTerraformConfig renders Terraform config (*.tf) from path using provided variables.
// If path is directory then all *.tf files are combined before applying variables.
func renderTerraformConfig(path string, vars map[string]string) (string, error) {
	tmpl, err := readConfigTemplates(path)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	t, err := template.New("default").Parse(tmpl)
	if err != nil {
		return "", err
	}
	err = t.Execute(&b, vars)
	return b.String(), err
}

func readConfigTemplates(path string) (string, error) {
	f, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !f.IsDir() {
		d, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return fmt.Sprint(d), nil
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return "", nil
	}
	var b bytes.Buffer
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".tf" {
			continue
		}
		d, err := os.ReadFile(filepath.Join(path, f.Name()))
		if err != nil {
			return "", err
		}
		b.Write(d)
	}
	return b.String(), nil
}
