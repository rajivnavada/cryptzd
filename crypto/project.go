package crypto

import (
	"database/sql"
	"time"
)

const (
	ACCESS_LEVEL_ADMIN = "admin"
	ACCESS_LEVEL_WRITE = "write"
	ACCESS_LEVEL_READ  = "read"
)

type projectCore struct {
	Id                 int       `db:"id"`
	Name               string    `db:"name"`
	Environment        string    `db:"environment"`
	DefaultAccessLevel string    `db:"default_access_level"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

type project struct {
	*projectCore
}

func (p project) Id() int {
	return p.projectCore.Id
}

func (p project) Name() string {
	return p.projectCore.Name
}

func (p project) Environment() string {
	return p.projectCore.Environment
}

func (p project) DefaultAccessLevel() string {
	return p.projectCore.DefaultAccessLevel
}

func (p project) CreatedAt() time.Time {
	return p.projectCore.CreatedAt
}

func (p project) UpdatedAt() time.Time {
	return p.projectCore.UpdatedAt
}

func (p project) Members(dbMap DataMapper) ([]ProjectMember, error) {
	var ret []ProjectMember
	var members []*projectMemberCore
	_, err := dbMap.Select(&members, "SELECT * FROM project_members WHERE project_id = ? ORDER BY created_at ASC", p.Id())
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		ret = append(ret, &projectMember{m})
	}
	return ret, nil
}

func (p project) AddMember(userId int, dbMap DataMapper) (ProjectMember, error) {
	// Use p.Id() and userId to locate a member record.
	pm, err := FindProjectMemberWithUserId(userId, p.Id(), dbMap)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	// If it does not exist, create and save the record
	if err == sql.ErrNoRows {
		pm = NewProjectMember(userId, p.Id())
		err = pm.Save(dbMap)
		if err != nil {
			return nil, err
		}
	}
	return pm, nil
}

func (p project) RemoveMember(userId int, dbMap DataMapper) error {
	// Find the member using p.Id() and userId
	pm, err := FindProjectMemberWithUserId(userId, p.Id(), dbMap)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		return nil
	}
	// If it exists, delete it
	_, err = dbMap.Delete(pm)
	return err
}

func (p project) Credentials(dbMap DataMapper) ([]ProjectCredentialKey, error) {
	var ret []ProjectCredentialKey
	var creds []*projectCredentialKeyCore
	_, err := dbMap.Select(&creds, "SELECT * FROM project_credentials WHERE project_id = ? ORDER BY created_at ASC", p.Id())
	if err != nil {
		return nil, err
	}
	for _, c := range creds {
		ret = append(ret, &projectCredentialKey{c})
	}
	return ret, nil
}

func (p project) AddCredential(key, value string, dbMap DataMapper) (ProjectCredentialKey, error) {
	// Figure out if the combo of key & p.Id exists
	// If it exists, update the value record
	// Else create a new credential key record
	// use the key record to add a value record
	return nil, NotImplementedError
}

func (p project) UpdateCredential(key, value string, dbMap DataMapper) (ProjectCredentialKey, error) {
	// Get the key record using key & p.Id combo
	// if key does not exist, return error
	// if key exists, create a value record for each project member
	return nil, NotImplementedError
}

func (p project) RemoveCredential(key string, dbMap DataMapper) error {
	// Find the member using p.Id() and userId
	pk, err := FindProjectCredentialKey(key, p.Id(), dbMap)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		return nil
	}
	// If it exists, delete it
	_, err = dbMap.Delete(pk)
	return err
}

func (p project) Save(dbMap DataMapper) error {
	if p.Id() > 0 {
		_, err := dbMap.Update(p.projectCore)
		return err
	}
	return dbMap.Insert(p.projectCore)
}

func NewProject(name, environment, defaultAccessLevel string) Project {
	if defaultAccessLevel == "" {
		defaultAccessLevel = ACCESS_LEVEL_READ
	}
	return &project{&projectCore{
		Name:               name,
		Environment:        environment,
		DefaultAccessLevel: defaultAccessLevel,
	}}
}
