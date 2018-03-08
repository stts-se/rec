package admin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// TODO Return error? Doesn't check whether path is dir or file if it exists
func userExists(audioDir, userName string) bool {
	userName = strings.ToLower(userName)
	userDirName := filepath.Join(audioDir, userName)

	fi, err := os.Stat(userDirName)
	//if os.IsNotExist(err) {
	//	return false
	//} else {
	if os.IsNotExist(err) {
		return false
	}

	if !fi.IsDir() {
		return false
	}

	return true
}

func createBaseDir(d string) error {
	_, err := os.Stat(d)
	if !os.IsNotExist(err) {
		return fmt.Errorf("dir already exists: '%s'", d)
	}

	err = os.MkdirAll(d, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create base dir '%s' : ", d, err)
	}

	return nil
}

func subDirs(dirPath string) []os.FileInfo {
	var res []os.FileInfo

	fs, _ := ioutil.ReadDir(dirPath)
	for _, f := range fs {
		if f.IsDir() {
			res = append(res, f)
		}
	}

	return res
}

func addUser(baseDir, userName string) error {
	userName = strings.ToLower(userName)
	userDirName := filepath.Join(baseDir, userName)
	_, err := os.Stat(userDirName)
	if !os.IsNotExist(err) {
		return fmt.Errorf("user already exists '%s'", userName)
	}

	err = os.MkdirAll(userDirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to add user '%s'", userName)
	}

	return nil
}

func deleteUser(baseDir, userName string) error {
	userName = strings.ToLower(userName)
	userDirName := filepath.Join(baseDir, userName)
	_, err := os.Stat(userDirName)
	if os.IsNotExist(err) {
		return fmt.Errorf("no such user '%s'", userName)
	}

	err = os.RemoveAll(userDirName)
	if err != nil {
		fmt.Errorf("failed to delete user '%s' : %v", userName, err)
	}

	return nil
}
