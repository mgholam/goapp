package myos

import "os"

func DirectoryExists(dir string) bool {

	_, e := os.Stat(dir)
	return e == nil
}

func FileExists(fn string) bool {

	_, e := os.Stat(fn)
	return e == nil
}

func CreateDir(dir string) {
	os.Mkdir(dir, 0755)
}
