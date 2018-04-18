package admin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/stts-se/rec"
)

// For writing files
var mutex sync.Mutex

const uttSuffix = ".utt"

// TODO Return error? Doesn't check whether path is dir or file if it exists
func userExists(baseDir, userName string) bool {
	userName = strings.ToLower(userName)
	userDirName := filepath.Join(baseDir, userName)

	fi, err := os.Stat(userDirName)
	if os.IsNotExist(err) {
		return false
	}

	if fi == nil || !fi.IsDir() {
		return false
	}

	return true
}

func createBaseDir(d string) error {
	_, err := os.Stat(d)
	if !os.IsNotExist(err) {
		return fmt.Errorf("dir already exists: '%s'", d)
	}

	mutex.Lock()
	defer mutex.Unlock()
	err = os.MkdirAll(d, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create base dir '%s' : %v", d, err)
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

var defaultTestUtterances = `testutt_0001_is	is
testutt_0002_bi	bi
testutt_0003_rose	rose
testutt_0004_blaes	blæs
testutt_0005_mus	mus
testutt_0006_sne	sne
testutt_0007_e	e
testutt_0008_i	i
testutt_0009_o	o
testutt_0010_u	u
testutt_0011_å	å
`

func AddUser(baseDir, userName string) error {
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

	mutex.Lock()
	defer mutex.Unlock()
	err = os.MkdirAll(userDirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to add user '%s'", userName)
	}
	log.Printf("[admin] Created folder %s for user %s", userDirName, userName)

	// create default utterance list
	uttFile := filepath.Join(userDirName, "test_utterances.utt")
	err = ioutil.WriteFile(uttFile, []byte(defaultTestUtterances), 0644)
	if err != nil {
		return fmt.Errorf("couldn't create default utterance file %s for user '%s'", uttFile, userName)
	}

	return nil
}

func UserExists(baseDir, userName string) (bool, error) {
	_, err := os.Stat(baseDir)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("dir does not exist '%s'", baseDir)
	}
	userName = strings.ToLower(userName)
	userDirName := filepath.Join(baseDir, userName)
	_, err = os.Stat(userDirName)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func DeleteUser(baseDir, userName string) error {
	userName = strings.ToLower(userName)
	userDirName := filepath.Join(baseDir, userName)
	_, err := os.Stat(userDirName)
	if os.IsNotExist(err) {
		return fmt.Errorf("no such user '%s'", userName)
	}

	mutex.Lock()
	defer mutex.Unlock()
	err = os.RemoveAll(userDirName)
	if err != nil {
		fmt.Errorf("failed to delete user '%s' : %v", userName, err)
	}

	return nil
}

func ListUsers(baseDir string) ([]string, error) {
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

func ListUtts(baseDir, userName string) ([]rec.UttList, error) {
	var res []rec.UttList

	userName = strings.ToLower(userName)
	userDir := filepath.Join(baseDir, userName)
	//TODO Check if user exists, and return error if not?

	ls := filepath.Join(userDir, "*"+uttSuffix)

	files, err := filepath.Glob(ls)
	if err != nil {
		return res, fmt.Errorf("failed to list utterances for '%s' : %v", userName, err)
	}

	for _, f := range files {
		// TODO frail: ".utt" or "." + "utt" ...?
		fn := strings.TrimSuffix(filepath.Base(f), uttSuffix)
		utts, err := ReadUttFile(baseDir, userName, fn)
		if err != nil {
			return res, fmt.Errorf("failed to read utterance file '%s' : %v", fn, err)

		}
		res = append(res, rec.UttList{Name: fn, Utts: utts})
	}

	return res, nil
}

func WriteSimpleUttFile(baseDir, userName, baseFileName string, utts []rec.Utterance) error {

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

		err := AddUser(baseDir, userName)
		if err != nil {
			return fmt.Errorf("failed to create user '%s' : %v", userName, err)
		}
	}

	var lines [][]byte
	for _, u := range utts {
		l := fmt.Sprintf("%s\t%s", u.RecordingID, u.Text)
		lines = append(lines, []byte(l))
	}

	var data = bytes.Join(lines, []byte("\n"))
	data = append(data, []byte("\n")...)

	uttFileName := filepath.Join(userDirName, baseFileName+uttSuffix)

	mutex.Lock()
	defer mutex.Unlock()

	err = ioutil.WriteFile(uttFileName, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create utterance file '%s' : %v", uttFileName, err)
	}

	return nil
}

func ReadUttFile(baseDir, userName, baseFileName string) ([]rec.Utterance, error) {
	var res0, res []rec.Utterance
	userName = strings.ToLower(userName)

	mutex.Lock()
	defer mutex.Unlock()
	fn := filepath.Join(baseDir, userName, baseFileName+uttSuffix)
	bts, err := ioutil.ReadFile(fn)
	if err != nil {
		return res, fmt.Errorf("failed to read utt file '%s'", fn)
	}
	lines := strings.Split(string(bts), "\n")

	n := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		fs := strings.SplitN(l, "\t", 2)
		// TODO error check
		n++
		res0 = append(res0, rec.Utterance{
			UserName:    userName,
			RecordingID: fs[0],
			Text:        fs[1],
			//TODO maybe we don't need the Num and Of fields
			Num: n,
		})
	}

	//TODO maybe we don't need the Num and Of fields
	for _, u := range res0 {
		u.Of = n
		res = append(res, u)
	}

	return res, nil
}
