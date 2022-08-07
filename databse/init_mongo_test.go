package db

import (
	"testing"

	"github.com/hrygo/gosmsn/bootstrap"
)

func TestInitDB(t *testing.T) {
	InitDB(bootstrap.ConfigYml, "AuthClient.Mongo")
}
