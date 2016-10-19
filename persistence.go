package incognitomail

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

// IncognitoData holds a "connection" to the persistence layer. To create a valid IncognitoData object, call OpenIncognitoData().
type IncognitoData struct {
	db *bolt.DB
}

const (
	targetsBucketName  = "targets"
	accountsBucketName = "accounts"
	handlesBucketName  = "handles"
)

var (
	// ErrEmptySecret is used when an empty account secret is used.
	ErrEmptySecret = errors.New("empty secret")

	// ErrEmptyTarget is used when an empty account target is used.
	ErrEmptyTarget = errors.New("empty target")

	// ErrAccountNotFound is used when an action requires an account to exist, but it wasn't found.
	ErrAccountNotFound = errors.New("account not found")

	// ErrAccountExists is used when trying to create an account with a given secret, but it already exists.
	ErrAccountExists = errors.New("account already exists")

	// ErrHandleNotFound is used when an action requires a handle to exist, but it wasn't found.
	ErrHandleNotFound = errors.New("handle not found")

	// ErrHandleExists is used when trying to create a handle, but it already exists.
	ErrHandleExists = errors.New("handle already exists")
)

// OpenIncognitoData returns an IncognitoData object with a successful "connection" to the persistence layer, ready to be used.
func OpenIncognitoData() (*IncognitoData, error) {
	db, err := bolt.Open(Config.Persistence.DatabasePath, 0600, nil)

	if err != nil {
		return nil, err
	}

	// Create "static" buckets that are used for persistence
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(targetsBucketName))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(accountsBucketName))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(handlesBucketName))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &IncognitoData{
		db: db,
	}, nil
}

// NewAccount generates a new account with the given secret and target email address.
func (a *IncognitoData) NewAccount(secret, target string) error {
	if secret == "" {
		return ErrEmptySecret
	}

	if target == "" {
		return ErrEmptyTarget
	}

	err := a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(secret))

		if b != nil {
			return ErrAccountExists
		}

		_, err := tx.CreateBucket([]byte(secret))
		if err != nil {
			return err
		}

		b = tx.Bucket([]byte(targetsBucketName))
		err = b.Put([]byte(secret), []byte(target))
		if err != nil {
			return err
		}

		now, err := time.Now().GobEncode()

		if err != nil {
			return err
		}

		b = tx.Bucket([]byte(accountsBucketName))
		err = b.Put([]byte(secret), now)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// DeleteAccount deletes all information related to the account with the given secret. If no account with that secret exists, it does nothing.
func (a *IncognitoData) DeleteAccount(secret string) {
	if secret == "" {
		return
	}

	// Delete all handles associated with this account first
	handles, err := a.ListAccountHandles(secret)
	if err != nil {
		return
	}

	for _, v := range handles {
		a.DeleteAccountHandle(secret, v)
	}

	a.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(secret))

		b := tx.Bucket([]byte(targetsBucketName))
		b.Delete([]byte(secret))

		b = tx.Bucket([]byte(accountsBucketName))
		b.Delete([]byte(secret))

		return nil
	})
}

// NewAccountHandle stores the given handle for the account with the given secret.
func (a *IncognitoData) NewAccountHandle(secret, handle string) error {
	if secret == "" {
		return ErrEmptySecret
	}

	err := a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(secret))
		if b == nil {
			return ErrAccountNotFound
		}

		hb := tx.Bucket([]byte(handlesBucketName))
		h := hb.Get([]byte(handle))
		if h != nil {
			return ErrHandleExists
		}

		now, err := time.Now().GobEncode()
		if err != nil {
			return err
		}

		err = b.Put([]byte(handle), now)
		if err != nil {
			return err
		}

		err = hb.Put([]byte(handle), now)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// DeleteAccountHandle deletes the given handle from the account with the given secret. If either the account or the handle does not exist, this does nothing.
func (a *IncognitoData) DeleteAccountHandle(secret, handle string) {
	if secret == "" || handle == "" {
		return
	}

	// Note that we still return errors from the following func, but don't care about them
	a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(secret))
		if b == nil {
			return ErrAccountNotFound
		}

		// Also delete from the global handles name
		hb := tx.Bucket([]byte(handlesBucketName))

		b.Delete([]byte(handle))
		hb.Delete([]byte(handle))
		return nil
	})
}

// GetAccountTarget returns the target registered for the account with the given secret.
func (a *IncognitoData) GetAccountTarget(secret string) (string, error) {
	if secret == "" {
		return "", ErrEmptySecret
	}

	var target string

	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(targetsBucketName))
		t := b.Get([]byte(secret))
		if t == nil {
			return ErrAccountNotFound
		}

		// Note: boltdb only keeps the value of t until the transaction ends, so we must copy it somewhere else now.
		// However, the call to string(t) internally does that for us, as it will ultimately call copy() to copy the values to a new byte slice for the resulting string.
		target = string(t)
		return nil
	})

	if err != nil {
		return "", err
	}

	return target, nil
}

// HasAccount returns true if an account with the given secret exists, false otherwise.
func (a *IncognitoData) HasAccount(secret string) bool {
	if secret == "" {
		return false
	}

	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(accountsBucketName))
		t := b.Get([]byte(secret))
		if t == nil {
			return ErrAccountNotFound
		}

		return nil
	})

	return err == nil
}

// HasHandleGlobal returns true if the given handle exists for any account, false otherwise.
func (a *IncognitoData) HasHandleGlobal(handle string) bool {
	if handle == "" {
		return false
	}

	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(handlesBucketName))
		t := b.Get([]byte(handle))
		if t == nil {
			return ErrAccountNotFound
		}

		return nil
	})

	return err == nil
}

// ListAccountHandles returns an array with all handles from the account with the given secret.
func (a *IncognitoData) ListAccountHandles(secret string) ([]string, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}

	var result []string

	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(secret))

		b.ForEach(func(k, v []byte) error {
			// Note: boltdb only keeps the values of k and v until the transaction ends, so we must copy these values somewhere else now.
			// However, the call to string(k) internally does that for us, as it will ultimately call copy() to copy the values to a new byte slice for the resulting string.
			result = append(result, string(k))
			return nil
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Close closes the "connection" with the persistence layer.
func (a *IncognitoData) Close() {
	a.db.Close()
}
