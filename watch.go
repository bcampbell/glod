package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
)

// blocks until a file is changed
func waitForChanges(site Site) error {

	targs := []string{getStr(site, "_configfile"),
		getStr(site, "_templatesdir"),
		getStr(site, "_skeldir"),
		getStr(site, "_contentdir"),
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	for _, targ := range targs {
		err = addRecursive(watcher, targ)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case event := <-watcher.Events:
			// TODO:
			// - ignore swapfiles,
			// - ignore new dir creation
			fmt.Println("event:", event)
			return nil
		case err := <-watcher.Errors:
			return err
		}
	}
}

func addRecursive(watcher *fsnotify.Watcher, targ string) error {

	fi, err := os.Stat(targ)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		fmt.Println("add file ", targ)
		return watcher.Add(targ)
	}

	err = filepath.Walk(targ, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			fmt.Println("add dir ", path)
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
