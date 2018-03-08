package admin

import (
	"os"
	"testing"
)

func TestUser(t *testing.T) {

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

	err = addUser(d, u)
	if err != nil {
		t.Errorf("failed to add user '%s'", u)
	}

	ue2 := userExists(d, u)
	if !ue2 {
		t.Errorf("expected true, got false")
	}

	// Shouldn't be able to add same user twice
	err = addUser(d, u)
	if err == nil {
		t.Errorf("should not be able to add existing user '%s'", u)
	}

	err = deleteUser(d, u)
	if err != nil {
		t.Errorf("failed to delete user '%s' : %v", u, err)
	}

	// Can't delete user twice
	err = deleteUser(d, u)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

}
