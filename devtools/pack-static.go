package main

// utility for packaging static assets into a zip archive.
// while the templates are compiled into the executable, the static assets
// (e.g. stylesheets) are not. i plan to have the executable to set up
// everything during the installation process but baking everything into
// the executable is bad feng shui; to provide one executable + one .zip
// seemes to be a good compromise.
// this file should not be run on its own but as a bigger building process.

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

func isTemporaryFile(f string) bool {
	if strings.HasPrefix(f, ".#") { return true }
	if strings.HasPrefix(f, "~") { return true }
	if strings.HasSuffix(f, "~") { return true }
	if strings.HasPrefix(f, "#") && strings.HasSuffix(f, "#") { return true }
	return false
}

func recursivelyWriteZip(w *zip.Writer, sourceBase string, prefix string) error {
	ls, err := os.ReadDir(sourceBase)
	if err != nil { return err }
	for _, item := range ls {
		if !item.IsDir() && isTemporaryFile(item.Name()) { continue }
		sourcePath := path.Join(sourceBase, item.Name())
		targetPath := fmt.Sprintf("%s%s", prefix, item.Name())
		if item.IsDir() {
			err = recursivelyWriteZip(w, sourcePath, targetPath+"/")
			if err != nil { return err }
		} else {
			wr, err := w.Create(targetPath)
			if err != nil { return err }
			f, err := os.Open(sourcePath)
			if err != nil { return err }
			_, err = io.Copy(wr, f)
			if err != nil { return err }
			err = f.Close()
			if err != nil { return err }
		}
	}
	return nil
}

func main() {
	staticDir := os.Args[1]
	f, err := os.OpenFile("static.zip", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { log.Panicln(err) }
	zipres := zip.NewWriter(f)
	err = recursivelyWriteZip(zipres, staticDir, "static/")
	if err != nil { log.Panicln(err) }
	zipres.Flush()
	zipres.Close()
	if err != nil { log.Panicln(err) }
}

