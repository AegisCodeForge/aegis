//go:build ignore

package main

/*
   template-generating utility.  collects all the related files within
   a directory (referred to as "the directory" below) and generates a
   `templates.go` file and a `models.go` file which then could be
   used. the directory is provided as the only command line
   argument. when it's a relative path, it's considered to be relative
   to the working directory when this utility is executed.  currently
   used for `go generate` thru invoking `go run`.

   currently does the 3 following things:

   1.  all files that ends with `.template.html` would be read and
       generated as templates with the following extra:
       +  any content between a line that only contains "<!-- model"
          and a line that only contains "-->" (in this order) is
          considered to be the definition of a "model" useful for this
          "view". the content within is directly taken out and written
          as go code for `models.go`.
       the name of the file (without the ".template.html" part) would
       be the name of the template, and is thus required to follow
       whatever rules and/or conventions that might exist.
   2.  all files that ends with `.func.go` would be read and generated
       as functions provided to the template. see the example in the
       following url:
       https://pkg.go.dev/text/template#example-Template-Func
       the name of the file (without the ".func.go" part) would be the
       name of the function for the template, and is thus required to
       follow whatever rules and/or conventions that might exist.
   3.  all files that ends with `.model.go` would be read and written
       to `models.go`.

   import statements are respected and would be inserted into the resulting
   files (i.e. `models.go` and `templates.go`), but only one kind of
   import syntax of golang is currently supported. Each imported module
   must have their own separate import statement, and qualified import is
   not supported (i.e. it should only be a sequence of `import "[module]"`)

*/

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
)

type template struct {
	name string
	modelDefinition string
	modelImports []*importSpec
	view string
}

type importSpec struct {
	k string
	source string
}

type model struct {
	imports []*importSpec
	content []string
}

type templateFunc struct {
	imports []*importSpec
	content []string
}

// note that according to the definition of go, this happens:
//     Import declaration          Local name of Sin
//     
//     import   "lib/math"         math.Sin
//     import m "lib/math"         m.Sin
//     import . "lib/math"         Sin
//     import _ "lib/math"         (none.)
// we must keep the qualified *module* names from clashing with each other
// but we should allow any amount of "import _" and ignore possible clash
// from "import .". for this reason we need to handle the three cases
// differently.
type importList struct {
	plain map[string]*importSpec  // key is module name, not module path
	dot []string
	underscore []string
}

func parseImport(s string) (*importSpec, error) {
	re, err := regexp.Compile("\\s*import\\s*([^\\s]*)?\\s*(.*)\\s*")
	if err != nil { return nil, err }
	ss := re.FindSubmatch([]byte(s))
	k := ""
	source := string(ss[1])
	if len(ss[2]) > 0 {
		k = string(ss[1])
		source = string(ss[2])
	}
	source = strings.TrimSpace(source)
	source = source[1:len(source)-1]
	return &importSpec{k: k, source: source}, nil
}

func parseTemplate(name string, f io.Reader) (*template, error) {
	data, err := io.ReadAll(f)
	if err != nil { return nil, err }
	rbegin, err := regexp.Compile("^<!--\\s*(?i:model)$")
	if err != nil { return nil, err }
	rend, err := regexp.Compile("^-->$")
	modelSource := make([]string, 0)
	modelImports := make([]*importSpec, 0)
	viewSource := make([]string, 0)
	readingModelDefinition := false
	for item := range strings.SplitSeq(string(data), "\n") {
		trimmed := strings.TrimSpace(item)
		if readingModelDefinition {
			if strings.HasPrefix(trimmed, "//") { continue }
			if strings.HasPrefix(trimmed, "import") {
				imp, _ := parseImport(trimmed)
				modelImports = append(modelImports, imp)
			} else if rend.Match([]byte(trimmed)) {
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
		modelImports: modelImports,
		view: strings.Join(viewSource, "\n"),
	}, nil
}

func parseModel(filename string, f io.Reader) (model, error) {
	data, err := io.ReadAll(f)
	if err != nil { log.Fatal(err) }
	content := make([]string, 0)
	content = append(content, "// from " + filename)
	imports := make([]*importSpec, 0)
	for item := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(item, "//") { continue }
		if strings.HasPrefix(item, "package") { continue }
		if strings.HasPrefix(item, "import") {
			imp, _ := parseImport(item)
			imports = append(imports, imp)
		} else {
			content = append(content, item)
		}
	}
	return model{imports:imports, content:content}, nil
}

func parseFunc(filename string, f io.Reader) (templateFunc, error) {
	data, err := io.ReadAll(f)
	if err != nil { log.Fatal(err) }
	content := make([]string, 0)
	content = append(content, "// from " + filename)
	imports := make([]*importSpec, 0)
	for item := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(item, "//") { continue }
		if strings.HasPrefix(item, "package") { continue }
		if strings.HasPrefix(item, "import") {
			imp, _ := parseImport(item)
			imports = append(imports, imp)
		} else {
			content = append(content, item)
		}
	}
	return templateFunc{imports:imports, content:content}, nil
}

func isTemporaryFile(s string) bool {
	return strings.HasPrefix(s, ".#")
}

func mkimportList() *importList {
	return &importList{
		plain: make(map[string]*importSpec, 0),
		dot: make([]string, 0),
		underscore: make([]string, 0),
	}
}

func collect(il *importList, a []*importSpec) *importList {
	for _, item := range a {
		if item.k == "_" {
			il.underscore = append(il.underscore, item.source)
		} else if item.k == "." {
			il.dot = append(il.dot, item.source)
		} else if len(item.k) > 0 {
			il.plain[item.k] = item
		} else {
			base := path.Base(item.source)
			il.plain[base] = item
		}
	}
	return il
}

func recursivelyCollectAllFile (base string, prefix string, suffix string) []string {
	// returns all file with a certain suffix under a directory (recursively).
	// the result would have no suffix.
	readBase := path.Join(base, prefix)
	s, err := os.ReadDir(readBase)
	if err != nil {
		log.Printf("Failed to read template dir: %s\n", base)
		return nil
	}
	res := make([]string, 0)
	dir := make([][]string, 0)
	for _, item := range s {
		p := path.Join(prefix, item.Name())
		if item.IsDir() {
			dir = append(dir, recursivelyCollectAllFile(base, p, suffix))
		} else {
			fileName := item.Name()
			if !strings.HasSuffix(fileName, suffix) { continue }
			if isTemporaryFile(fileName) { continue }
			templateName := fileName[:len(fileName)-len(suffix)]
			res = append(res, path.Join(prefix, templateName))
		}
	}
	res = slices.Concat(res, slices.Concat(dir...))
	return res
}


func main() {
	if len(os.Args) < 2 { log.Fatal("Source directory required.") }
	sourceDir := os.Args[1]
	packageName := path.Base(sourceDir)
	fileList := recursivelyCollectAllFile(sourceDir, "", ".template.html")

	modelImport := mkimportList()
	templateList := make([]*template, 0)
	for _, item := range fileList {
		fileName := fmt.Sprintf("%s.template.html", item)
		templateName := item
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
				item,
				err.Error(),
			)
		}
		thisTemplateF.Close()
		templateList = append(templateList, thisTemplate)
		modelImport = collect(modelImport, thisTemplate.modelImports)
	}
	
	modelList := make([]model, 0)
	fileList = recursivelyCollectAllFile(sourceDir, "", ".model.go")
	for _, item := range fileList {
		fileName := fmt.Sprintf("%s.model.go", item)
		p := path.Join(sourceDir, fileName)
		thisModelF, err := os.Open(p)
		if err != nil {
			log.Panicf(
				"Failed to open model file %s: %s\n",
				fileName,
				err.Error(),
			)
		}

		thisModel, err := parseModel(fileName, thisModelF)
		if err != nil {
			log.Panicf(
				"Failed to parse model file %s: %s\n",
				fileName,
				err.Error(),
			)
		}
		modelList = append(modelList, thisModel)
		modelImport = collect(modelImport, thisModel.imports)
	}

	funcImport := mkimportList()
	functionList := make(map[string]templateFunc, 0)
	fileList = recursivelyCollectAllFile(sourceDir, "", ".func.go")
	for _, item := range fileList {
		fileName := fmt.Sprintf("%s.func.go", item)
		if !strings.HasSuffix(fileName, ".func.go") { continue }
		if isTemporaryFile(fileName) { continue }
		funcName := fileName[:len(fileName)-len(".func.go")]
		p := path.Join(sourceDir, fileName)
		funcFile, err := os.Open(p)
		if err != nil {
			log.Panicf(
				"Failed to open function file %s: %s\n",
				fileName,
				err.Error(),
			)
		}
		funcObj, err := parseFunc(fileName, funcFile)
		if err != nil {
			log.Panicf(
				"Failed to read function file %s: %s\n",
				fileName,
				err.Error(),
			)
		}
		functionList[funcName] = funcObj
		funcImport = collect(funcImport, funcObj.imports)
	}
	
	templateTargetFilePath := path.Join(sourceDir, "templates.go")
	os.Remove(templateTargetFilePath)
	f, err := os.OpenFile(
		templateTargetFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_EXCL,
	    0644,
	)
	if err != nil { log.Panic(err) }
	defer f.Close()

	// start writing templates.go
	_, err = f.WriteString(`// generated by devtools/generate-template.go. DO NOT EDIT

package ` + packageName + `

import (
  "html/template"
  "log"
`)
	for _, spec := range funcImport.plain {
		f.WriteString(fmt.Sprintf("  %s \"%s\"\n", spec.k, spec.source))
	}
	for _, item := range funcImport.dot {
		f.WriteString(fmt.Sprintf("  . \"%s\"\n", item))
	}
	for _, item := range funcImport.underscore {
		f.WriteString(fmt.Sprintf("  _ \"%s\"\n", item))
	}

	_, err = f.WriteString(`
)

func LoadTemplate() *template.Template {
  var err error = nil
  masterTemplate := template.New("").Funcs(template.FuncMap{
`)
	if err != nil { log.Panic(err) }
	for key, value := range functionList {
		_, err = f.WriteString(
			fmt.Sprintf(
				"\"%s\": %s,\n",
				key, strings.TrimSpace(strings.Join(value.content, "\n")),
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

	// start writing models.go
	modelTargetFilePath := path.Join(sourceDir, "models.go")
	os.Remove(modelTargetFilePath)
	f2, err := os.OpenFile(
		modelTargetFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_EXCL,
		0644,
	)
	if err != nil { log.Panic(err) }
	defer f2.Close()

	_, err = f2.WriteString(`// generated by devtools/generate-template.go. DO NOT EDIT

package ` + packageName + `

import (
`)
	if err != nil { log.Panic(err) }
	for _, spec := range modelImport.plain {
		f2.WriteString(fmt.Sprintf("  %s \"%s\"\n", spec.k, spec.source))
	}
	for _, item := range modelImport.dot {
		f2.WriteString(fmt.Sprintf("  . \"%s\"\n", item))
	}
	for _, item := range modelImport.underscore {
		f2.WriteString(fmt.Sprintf("  _ \"%s\"\n", item))
	}
	_, err = f2.WriteString(`
)
`)
	if err != nil { log.Panic(err) }

	for _, item := range templateList {
		if len(strings.TrimSpace(item.modelDefinition)) > 0 {
			f2.WriteString(`
// from ` + item.name + ".template.html\n")
			f2.WriteString(item.modelDefinition)
			f2.WriteString("\n")
		}
	}

	for _, item := range modelList {
		content := strings.Join(item.content, "\n")
		if len(strings.TrimSpace(content)) > 0 {
			f2.WriteString(string(content))
			f2.WriteString("\n")
		}
	}
	
	fmt.Println("Done.")
}

