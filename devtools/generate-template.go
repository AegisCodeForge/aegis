//go:build ignore

package main

/*
   template-generating utility.
   collects all the related files within a directory (referred to as
   "the directory" below) and generates a `templates.go` file which
   then could be used. the directory is provided as the only command
   line argument. when it's a relative path, it's considered to be
   relative to the working directory when this utility is executed.
   currently used for `go generate` thru invoking `go run`.

   currently does the 3 following things:

   1.  the file `import-list`, when exists within the directory, is
       read with its content being written to the generated .go file's
       `import` part. the content of this file should be package names
       without double-quotes and each line should only contain one
       package. import identifiers are not supported.
   2.  all files that ends with `.func` would be read and generated
       as functions provided to the template. see the example in the
       following url:
       https://pkg.go.dev/text/template#example-Template-Func
       the name of the file (without the ".func.go" part) would be the
       name of the function for the template, and is thus required to
       follow whatever rules and/or conventions that might exist.
   3.  all files that ends with `.template.html` would be read and
       generated as templates with the following extra:
       +  any content between a line that only contains "<!-- model"
          and a line that only contains "-->" (in this order) is
          considered to be the definition of a "model" useful for this
          "view". the content within is directly taken out and written
          as go code for `templates.go`.
       the name of the file (without the ".template.html" part) would
       be the name of the template, and is thus required to follow
       whatever rules and/or conventions that might exist.
   
*/

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

type template struct {
	name string
	modelDefinition string
	view string
}

func parseTemplate(name string, f io.Reader) (*template, error) {
	data, err := io.ReadAll(f)
	if err != nil { return nil, err }
	rbegin, err := regexp.Compile("^<!--\\s*(?i:model)$")
	if err != nil { return nil, err }
	rend, err := regexp.Compile("^-->$")
	modelSource := make([]string, 0)
	viewSource := make([]string, 0)
	readingModelDefinition := false
	for item := range strings.SplitSeq(string(data), "\n") {
		trimmed := strings.TrimSpace(item)
		if readingModelDefinition {
			if rend.Match([]byte(trimmed)) {
				readingModelDefinition = false
			} else {
				modelSource = append(modelSource, item)
			}
		} else {
			if rbegin.Match([]byte(trimmed)) {
				readingModelDefinition = true
			} else {
				viewSource = append(viewSource, item)
			}
		}
	}
	if readingModelDefinition != false {
		return nil, errors.New("Unfinished model definition reading")
	}
	return &template{
		name: name,
		modelDefinition: strings.Join(modelSource, "\n"),
		view: strings.Join(viewSource, "\n"),
	}, nil
}

func isTemporaryFile(s string) bool {
	return strings.HasPrefix(s, ".#")
}

func main() {
	if len(os.Args) < 2 { log.Fatal("Source directory required.") }
	sourceDir := os.Args[1]
	fileList, err := os.ReadDir(sourceDir)
	if err != nil { log.Fatal(err) }
	
	templateList := make([]*template, 0)
	for _, item := range fileList {
		fileName := item.Name()
		if !strings.HasSuffix(fileName, ".template.html") { continue }
		if isTemporaryFile(fileName) { continue }
		templateName := fileName[:len(fileName)-len(".template.html")]
		p := path.Join(sourceDir, fileName)
		thisTemplateF, err := os.Open(p)
		if err != nil {
			log.Panicf(
				"Failed to open template file %s: %s\n",
				fileName,
				err.Error(),
			)
		}
		thisTemplate, err := parseTemplate(templateName, thisTemplateF)
		if err != nil {
			log.Panicf(
				"Failed to parse template file %s: %s\n",
				item.Name(),
				err.Error(),
			)
		}
		thisTemplateF.Close()
		templateList = append(templateList, thisTemplate)
	}

	functionList := make(map[string]string, 0)
	for _, item := range fileList {
		fileName := item.Name()
		if !strings.HasSuffix(fileName, ".func") { continue }
		if isTemporaryFile(fileName) { continue }
		funcName := fileName[:len(fileName)-len(".func")]
		p := path.Join(sourceDir, fileName)
		ff, err := os.Open(p)
		if err != nil {
			log.Panicf(
				"Failed to open function file %s: %s\n",
				fileName,
				err.Error(),
			)
		}
		ffsource, err := io.ReadAll(ff)
		if err != nil {
			log.Panicf(
				"Failed to read function file %s: %s\n",
				fileName,
				err.Error(),
			)
		}
		functionList[funcName] = string(ffsource)
	}
	
	targetFilePath := path.Join(sourceDir, "templates.go")
	os.Remove(targetFilePath)
	f, err := os.OpenFile(
		targetFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_EXCL,
	    0644,
	)
	if err != nil { log.Panic(err) }
	defer f.Close()
	importListPath := path.Join(sourceDir, "import-list")
	importListFile, err := os.ReadFile(importListPath)
	var importList []string = nil
	if err == nil {
		s := string(importListFile)
		importList = strings.Split(s, "\n")
	}
	
	_, err = f.WriteString(`// generated by devtools/generate-template.go. DO NOT EDIT

package templates

import (
  "html/template"
  "log"

`)
	if err != nil { log.Panic(err) }
	if importList != nil {
		for _, lib := range importList {
			if len(strings.TrimSpace(lib)) <= 0 { continue }
			_, err = f.WriteString("\"" + lib + "\"\n")
			if err != nil {
				log.Fatalf("Failed to write import \"%s\"", lib)
			}
		}
	}
	_, err = f.WriteString(`
)
`)
	if err != nil { log.Panic(err) }
	for _, item := range templateList {
		_, err = f.WriteString(item.modelDefinition)
		if err != nil { log.Fatal(err) }
	}
	_, err = f.WriteString(`

func LoadTemplate() *template.Template {
  var err error = nil
  masterTemplate := template.New("").Funcs(template.FuncMap{
`)
	if err != nil { log.Panic(err) }
	for key, value := range functionList {
		_, err = f.WriteString(
			fmt.Sprintf(
				"\"%s\": %s,\n",
				key, strings.TrimSpace(value),
			),
		)
		if err != nil { log.Panic(err) }
	}
	_, err = f.WriteString(`
  })
`)
	if err != nil { log.Fatal(err) }
	for _, item := range templateList {
		_, err = f.WriteString(
			fmt.Sprintf(
				"  _, err = masterTemplate.New(\"%s\").Parse(`%s`)\n",
				item.name,
				item.view,
			),
		)
		if err != nil { log.Fatal(err) }
		_, err = f.WriteString("  if err != nil { log.Fatal(err) }\n")
		if err != nil { log.Fatal(err) }
	}
	_, err = f.WriteString("\n  return masterTemplate\n}\n")
	if err != nil { log.Fatal(err) }
	fmt.Println("Done.")
}

