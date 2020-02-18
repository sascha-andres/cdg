package main

import (
	"fmt"
	"github.com/apoorvam/goterminal"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	rootPath  = kingpin.Flag("path", "Path to scan for git repositories").Required().String()
	cacheFile = kingpin.Flag("cache-file", "Path to cachefile").Required().String()
)

func main() {
	var (
		err         error
		directories = make([]string, 0)
	)
	_ = kingpin.Parse()

	// initialize channels
	c := make(chan string, 2)
	d := make(chan bool)
	defer close(c)
	defer close(d)

	// walk through directory
	go func() {
		err = filepath.Walk(*rootPath, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			if info.IsDir() && info.Name() == ".git" {
				p := strings.TrimSuffix(path, info.Name())
				c <- p
			}
			return nil
		})
		d <- true
	}()

	writer := goterminal.New(os.Stdout)
	for {
		exit := false
		select {
		case gitDirectory := <-c:
			writer.Clear()
			directories = append(directories, gitDirectory)
			fmt.Fprintf(writer, "found: %s\n", gitDirectory)
			writer.Print()
		case <-d:
			exit = true
		}
		if exit {
			break
		}
	}

	if err != nil {
		os.Exit(1)
	}
	writer.Clear()

	var sb strings.Builder

	for _, dir := range directories {
		_, _ = sb.WriteString(fmt.Sprintf("%s\n", dir))
	}

	err = ioutil.WriteFile(*cacheFile, []byte(sb.String()), 0600)
	if err != nil {
		os.Exit(1)
	}

	fmt.Fprintf(writer, "wrote %d directories to cache", len(directories))
	writer.Print()

	os.Exit(0)
}
