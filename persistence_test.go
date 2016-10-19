package incognitomail_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/danielsidhion/incognitomail"
)

const (
	accountSecret1  = "testsecret1"
	accountSecret2  = "testsecret2"
	accountTarget1  = "testtarget1@example.com"
	accountTarget2  = "testtarget2@example.com"
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

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	res := data.HasAccount(accountSecret1)
	if !res {
		t.Fatal("account created is not present")
	}

	commonTeardown(t, data)
}

// Ensure a new account needs a non-empty secret.
func TestPersistence_NewAccount_SecretRequired(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount("", accountTarget1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != incognitomail.ErrEmptySecret {
		t.Fatal("expected ErrEmptySecret")
	}

	commonTeardown(t, data)
}

// Ensure a new account needs a non-empty target.
func TestPersistence_NewAccount_TargetRequired(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, "")
	if err == nil {
		t.Fatal("expected error")
	}

	if err != incognitomail.ErrEmptyTarget {
		t.Fatal("expected ErrEmptyTarget")
	}

	commonTeardown(t, data)
}

// Ensure an account's target is successfully retrieved after creating.
func TestPersistence_CheckTarget(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	target, err := data.GetAccountTarget(accountSecret1)
	if err != nil {
		t.Fatal(err)
	}

	if target != accountTarget1 {
		t.Fatal("retrieved account target is not the same as inserted")
	}

	commonTeardown(t, data)
}

// Ensure deleting an account actually deletes its secret from the DB.
func TestPersistence_DeleteAccount(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	data.DeleteAccount(accountSecret1)

	res := data.HasAccount(accountSecret1)
	if res {
		t.Fatal("deleted account is still present")
	}

	commonTeardown(t, data)
}

// Ensure a new handle can be created without errors.
func TestPersistence_NewHandle(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	commonTeardown(t, data)
}

// Ensure a repeated handle can't be created (same account).
func TestPersistence_RepeatedHandle_SameAccount(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != incognitomail.ErrHandleExists {
		t.Fatal("expected ErrHandleExists")
	}

	commonTeardown(t, data)
}

// Ensure a repeated handle can't be created (different accounts).
func TestPersistence_RepeatedHandle_DifferentAccounts(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccount(accountSecret2, accountTarget2)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret2, accountHandle1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != incognitomail.ErrHandleExists {
		t.Fatal("expected ErrHandleExists")
	}

	commonTeardown(t, data)
}

// Ensure an account's handles are listed successfully.
func TestPersistence_ListHandles(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle2)
	if err != nil {
		t.Fatal(err)
	}

	handles, err := data.ListAccountHandles(accountSecret1)
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

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
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

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	data.DeleteAccountHandle(accountSecret1, accountHandle1)

	handles, err := data.ListAccountHandles(accountSecret1)
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

// Ensure a deleted account also deletes the handles from the global list.
func TestPersistence_DeleteAccount_GlobalHandles(t *testing.T) {
	data := commonSetup(t)

	err := data.NewAccount(accountSecret1, accountTarget1)
	if err != nil {
		t.Fatal(err)
	}

	err = data.NewAccountHandle(accountSecret1, accountHandle1)
	if err != nil {
		t.Fatal(err)
	}

	data.DeleteAccount(accountSecret1)

	res := data.HasHandleGlobal(accountHandle1)

	if res {
		t.Fatal("global handle check still identifies deleted account's handle ", accountHandle1)
	}

	commonTeardown(t, data)
}
