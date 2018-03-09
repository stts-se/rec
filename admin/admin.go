package admin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/stts-se/rec"
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
	_, err := os.Stat(baseDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("dir does not exist '%s'", baseDir)
	}

	userName = strings.ToLower(userName)
	userDirName := filepath.Join(baseDir, userName)
	_, err = os.Stat(userDirName)
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

func listUsers(baseDir string) ([]string, error) {
	var res []string
	fi, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return res, fmt.Errorf("failed to list users : %v", err)
	}

	for _, f := range fi {
		if f.IsDir() {
			res = append(res, f.Name())
		}
	}
	return res, nil
}

func writeSimpleUttFile(baseDir, userName, baseFileName string, utts []rec.Utterance) error {
	_, err := os.Stat(baseDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("no such dir '%s'", baseDir)
	}

	userName = strings.ToLower(userName)
	userDirName := filepath.Join(baseDir, userName)
	fi, err := os.Stat(userDirName)
	// If user not already exists, create a dir of name userName
	if os.IsNotExist(err) {
		if fi != nil && !fi.IsDir() {
			return fmt.Errorf("failed to create user: non-dir file of same name already exists '%s'", userName)
		}

		err := addUser(baseDir, userName)
		if err != nil {
			return fmt.Errorf("failed to create user '%s' : %v", userName, err)
		}
	}

	var lines [][]byte
	for _, u := range utts {
		lines = append(lines, []byte(u.Text))
	}

	var data = bytes.Join(lines, []byte("\n"))
	data = append(data, []byte("\n")...)

	uttFileName := filepath.Join(userDirName, baseFileName+".utt")

	err = ioutil.WriteFile(uttFileName, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create utterance file '%s' : %v", uttFileName, err)
	}

	return nil
}
