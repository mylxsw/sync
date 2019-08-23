package utils

import (
	uuid "github.com/satori/go.uuid"
)

// UUID create a uuid
func UUID() string {
	return uuid.NewV4().String()
}
