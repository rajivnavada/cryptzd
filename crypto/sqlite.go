package crypto

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
)

type DataMapper struct {
	*gorp.DbMap
}

func NewDataMapper() (*DataMapper, error) {
	db, err := sql.Open("sqlite3", SqliteFilePath)
	if err != nil {
		return nil, err
	}

	dbMap := &gorp.DbMap{
		Db:      db,
		Dialect: gorp.SqliteDialect{},
	}

	// Add the tables
	dbMap.AddTableWithName(userCore{}, "users").SetKeys(true, "Id")
	dbMap.AddTableWithName(publicKeyCore{}, "public_keys").SetKeys(true, "Id")
	dbMap.AddTableWithName(encryptedMessageCore{}, "encrypted_messages").SetKeys(true, "Id")
	dbMap.AddTableWithName(projectCore{}, "projects").SetKeys(true, "Id")
	dbMap.AddTableWithName(projectMemberCore{}, "project_members").SetKeys(true, "Id")
	dbMap.AddTableWithName(projectCredentialKeyCore{}, "project_credential_keys").SetKeys(true, "Id")
	dbMap.AddTableWithName(projectCredentialValueCore{}, "project_credential_values").SetKeys(true, "Id")

	return &DataMapper{dbMap}, nil
}

func (d *DataMapper) Close() {
	d.DbMap.Db.Close()
}
