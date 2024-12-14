package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/output"
)

func baseExecute(
	ctx context.Context,
	ctr *container.Container,
	filepath string,
	store *sync.Map,
	threadOnlyStore *sync.Map,
	outputRoot string,
	outputCtr output.OutputContainer,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// selfLoopCount := ""
	// if v, exists := threadOnlyStore.Load("loopCount"); exists {
	// 	selfLoopCount = v.(string)
	// }

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
		buffer.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	yamlTemplate := buffer.String()

	tmpl, err := template.New("yaml").Funcs(sprig.TxtFuncMap()).Parse(yamlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse yaml: %v", err)
	}

	data := map[string]interface{}{
		"DataCount": 10,
		"Name":      "John Doe",
		"Number":    42,
		"Items":     []string{"one", "two", "three"},
		"MapItems": []map[string]any{
			{"Name": "one", "Value": 1},
			{"Name": "two", "Value": 2},
			{"Name": "three", "Value": 3},
		},
		"Release": map[string]string{
			"Name": "releaseName",
			"Date": "2024-01-01",
			"Time": "12:00:00",
		},
	}
	var yamlBuf *bytes.Buffer = new(bytes.Buffer)
	if err := tmpl.Execute(yamlBuf, data); err != nil {
		return fmt.Errorf("failed to execute yaml: %v", err)
	}

	var rawData bytes.Buffer
	reader := io.TeeReader(yamlBuf, &rawData)

	var conf Runner
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&conf); err != nil {
		return fmt.Errorf("failed to decode yaml: %v", err)
	}

	validRunner, err := conf.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate runner: %v", err)
	}

	switch validRunner.Kind {
	case RunnerKindStoreValue:
	}

	fmt.Println(validRunner)
	fmt.Println(rawData.String())

	return nil
}
