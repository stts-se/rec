package admin

import (
	"os"
	"testing"

	"github.com/stts-se/rec"
)

func TestUser(t *testing.T) {

	// set up
	d := "tmtTstUserDir"
	_, err := os.Stat(d)
	if !os.IsNotExist(err) {
		os.RemoveAll(d)
	}
	err = createBaseDir(d)
	if err != nil {
		t.Errorf("failed to create dir '%s': %v", d, err)
	}

	u := "userZero"

	ue1 := userExists(d, u)
	if ue1 {
		t.Errorf("expected false, got true")
	}

	err = AddUser(d, u)
	if err != nil {
		t.Errorf("failed to add user '%s'", u)
	}

	ue2 := userExists(d, u)
	if !ue2 {
		t.Errorf("expected true, got false")
	}

	// Shouldn't be able to add same user twice
	err = AddUser(d, u)
	if err == nil {
		t.Errorf("should not be able to add existing user '%s'", u)
	}

	err = DeleteUser(d, u)
	if err != nil {
		t.Errorf("failed to delete user '%s' : %v", u, err)
	}

	// Can't delete user twice
	err = DeleteUser(d, u)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	_, err = ListUsers("_?==)(#)(#(#")
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	us, err := ListUsers(d)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if w, g := 0, len(us); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	err = AddUser(d, u)
	us, err = ListUsers(d)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if w, g := 1, len(us); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	err = AddUser(d+"_no_such_dir", u)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	u2 := "second_user"
	err = AddUser(d, u2)
	if err != nil {
		t.Errorf("%v", err)
	}

	us, err = ListUsers(d)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if w, g := 2, len(us); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	err = DeleteUser(d, u)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	err = DeleteUser(d, u2)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	us, err = ListUsers(d)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if w, g := 0, len(us); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	// clean up
	_, err = os.Stat(d)
	if !os.IsNotExist(err) {
		os.RemoveAll(d)
	}
}

func TestWriteSimpleUttFile(t *testing.T) {
	// set up
	d := "tmtTstUserDir2"
	_, err := os.Stat(d)
	if !os.IsNotExist(err) {
		os.RemoveAll(d)
	}
	err = createBaseDir(d)
	if err != nil {
		t.Errorf("failed to create dir '%s': %v", d, err)
	}

	u := "user_u1"
	fn := "test_utts"

	utts := []rec.Utterance{
		{RecordingID: "utt001", Text: "test utterance one"},
		{RecordingID: "utt002", Text: "test utterance two"},
		{RecordingID: "utt003", Text: "test utterance three"},
	}
	err = WriteSimpleUttFile(d, u, fn, utts)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	_, err = ReadUttFile("no", "such", "file")
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	utts1, err := ReadUttFile(d, u, fn)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if w, g := 3, len(utts1); w != g {
		t.Errorf("wanted %d, got %d", w, g)
	}

	uttList, err := ListUtts(d, u)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if w, g := 2, len(uttList); w != g { // one utt list created as default when AddUser is called
		t.Errorf("wanted %d, got %d", w, g)
	}
	if w, g := 3, len(uttList[1].Utts); w != g { // we want to check the 2nd utt list, which is the one we created here
		t.Errorf("wanted %d, got %d", w, g)
	}

	// clean up
	_, err = os.Stat(d)
	if !os.IsNotExist(err) {
		os.RemoveAll(d)
	}
}
