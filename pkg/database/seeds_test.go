package database

import (
	"errors"
	"testing"
)

func TestRunSeed_HashError(t *testing.T) {
	db := testDB
	
	origHash := hashPassword
	defer func() { hashPassword = origHash }()
	
	// Force HashPassword to fail
	hashPassword = func(password string) (string, error) {
		return "", errors.New("mock hashing error")
	}
	
	// RunSeed should log the error and continue
	RunSeed(db)
}
