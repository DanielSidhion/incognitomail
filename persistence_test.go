package incognitomail_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/danielsidhion/incognitomail"
)

const (
	accountSecret   = "testsecret"
	accountTarget   = "testtarget@example.com"
	accountHandle1  = "testhandle1"
	accountHandle2  = "testhandle2"
	neverUsedHandle = "testneverusedhandle"
)

// handleInsideList is a helper function to check if handles are contained in list of handles.
func handleInsideList(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}

	return false
}

// newDBFileName creates a new temporary file to force a brand new DB.
func newDBFileName(t *testing.T) {
	f, err := ioutil.TempFile("", "incognitomail_test_")
	if err != nil {
		t.Log("could not create temporary file")
		t.Fatal(err)
	}

	incognitomail.Config.Persistence.DatabasePath = f.Name()
	f.Close()
}

// removeCurrDB removes the temporary file created with newDBFileName().
func removeCurrDB(t *testing.T) {
	err := os.Remove(incognitomail.Config.Persistence.DatabasePath)
	if err != nil {
		t.Log("could not remove temporary file used for database")
	}
}

// commonSetup should be called at the beginning of each test to ensure a clean DB.
func commonSetup(t *testing.T) *incognitomail.IncognitoData {
	newDBFileName(t)
	data, err := incognitomail.OpenIncognitoData()
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// commonTeardown should be called at the end of each test to clean up the generated DB.
func commonTeardown(t *testing.T, data *incognitomail.IncognitoData) {
	data.Close()
	removeCurrDB(t)
}

// Ensure a new account can be created without errors.
func TestPersistence_NewAccount(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret, accountTarget)
	if err != nil {
		t.Fatal(err)
	}

	res := data.HasAccount(accountSecret)
	if !res {
		t.Fatal("account created is not present")
	}

	commonTeardown(t, data)
}

// Ensure an account's target is successfully retrieved after creating.
func TestPersistence_CheckTarget(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret, accountTarget)
	if err != nil {
		t.Fatal(err)
	}

	target, err := data.GetAccountTarget(accountSecret)
	if err != nil {
		t.Fatal(err)
	}

	if target != accountTarget {
		t.Fatal("retrieved account target is not the same as inserted")
	}

	commonTeardown(t, data)
}

// Ensure deleting an account actually deletes its secret from the DB.
func TestPersistence_DeleteAccount(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret, accountTarget)
	if err != nil {
		t.Fatal(err)
	}

	data.DeleteAccount(accountSecret)

	res := data.HasAccount(accountSecret)
	if res {
		t.Fatal("deleted account is still present")
	}

	commonTeardown(t, data)
}

// Ensure a new handle can be created without errors.
func TestPersistence_NewHandle(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret, accountTarget)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	commonTeardown(t, data)
}

// Ensure an account's handles are listed successfully.
func TestPersistence_ListHandles(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret, accountTarget)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret, accountHandle2)
	if err != nil {
		t.Fatal(err)
	}

	handles, err := data.ListAccountHandles(accountSecret)
	if err != nil {
		t.Fatal(err)
	}

	if len(handles) != 2 {
		t.Fatal("list of handles differ from amount of handles inserted")
	}

	if !handleInsideList(accountHandle1, handles) {
		t.Fatal("list of handles does not contain ", accountHandle1)
	}

	if !handleInsideList(accountHandle2, handles) {
		t.Fatal("list of handles does not contain ", accountHandle2)
	}

	commonTeardown(t, data)
}

// Ensure an account's handles make it into the global handle list.
func TestPersistence_CheckHandlesGlobal(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret, accountTarget)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	res := data.HasHandleGlobal(accountHandle1)

	if !res {
		t.Fatal("global handle check did not identify inserted handle")
	}

	commonTeardown(t, data)
}

// Ensure a deleted handle is removed from the account's handle list and the global handle list.
func TestPersistence_DeleteHandle(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret, accountTarget)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	data.DeleteAccountHandle(accountSecret, accountHandle1)

	handles, err := data.ListAccountHandles(accountSecret)
	if err != nil {
		t.Fatal(err)
	}

	if handleInsideList(accountHandle1, handles) {
		t.Fatal("list of handles still contains deleted handle ", accountHandle1)
	}

	res := data.HasHandleGlobal(accountHandle1)

	if res {
		t.Fatal("global handle check still identifies deleted handle ", accountHandle1)
	}

	commonTeardown(t, data)
}
