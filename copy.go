package main

import (
	"io"
	"os"
	"path/filepath"
)

func CopyDir(from, to string) error {

	fi, err := os.Stat(from)
	if err != nil {
		return err
	}

	err = os.MkdirAll(to, fi.Mode())
	if err != nil {
		return err
	}

	dir, err := os.Open(from)
	if err != nil {
		return err
	}

	ents, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, ent := range ents {

		srcName := filepath.Join(from, ent.Name())
		destName := filepath.Join(to, ent.Name())

		if ent.IsDir() {
			err = CopyDir(srcName, destName)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcName, destName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}
