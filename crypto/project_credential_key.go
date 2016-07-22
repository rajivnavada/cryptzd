package crypto

import (
	"time"
)

type projectCredentialKeyCore struct {
	Id        int       `db:"id"`
	ProjectId int       `db:"project_id"`
	Key       string    `db:"key"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type projectCredentialKey struct {
	*projectCredentialKeyCore
}

func (pk projectCredentialKey) Id() int {
	return pk.projectCredentialKeyCore.Id
}

func (pk projectCredentialKey) ProjectId() int {
	return pk.projectCredentialKeyCore.ProjectId
}

func (pk projectCredentialKey) Key() string {
	return pk.projectCredentialKeyCore.Key
}

func (pk projectCredentialKey) CreatedAt() time.Time {
	return pk.projectCredentialKeyCore.CreatedAt
}

func (pk projectCredentialKey) UpdatedAt() time.Time {
	return pk.projectCredentialKeyCore.UpdatedAt
}

func (pk projectCredentialKey) Save(dbMap *DataMapper) error {
	if pk.Id() > 0 {
		_, err := dbMap.Update(pk.projectCredentialKeyCore)
		return err
	}
	return dbMap.Insert(pk.projectCredentialKeyCore)
}

func (pk projectCredentialKey) ValueForUser(userId int, dbMap *DataMapper) (ProjectCredentialValue, error) {
	return nil, NotImplementedError
}

func FindProjectCredentialKey(key string, projectId int, dbMap *DataMapper) (ProjectCredentialKey, error) {
	pkc := &projectCredentialKeyCore{Key: key, ProjectId: projectId}
	err := dbMap.SelectOne(pkc, "SELECT * FROM project_credential_keys WHERE key = ? AND project_id = ?", pkc.Key, pkc.ProjectId)
	if err != nil {
		return nil, err
	}
	return &projectCredentialKey{pkc}, nil
}
