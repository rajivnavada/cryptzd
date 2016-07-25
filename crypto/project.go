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

func (p project) AddMember(userId int, accessLevel string, dbMap DataMapper) (ProjectMember, error) {
	// Use p.Id() and userId to locate a member record.
	pm, err := FindProjectMemberWithUserId(userId, p.Id(), dbMap)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	// If it does not exist, create and save the record
	if err == sql.ErrNoRows {
		pm = NewProjectMember(userId, p.Id(), accessLevel)
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

func (p project) GetCredential(key string, publicKeyId int, dbMap DataMapper) (ProjectCredentialValue, error) {
	pk, err := FindProjectCredentialKey(key, p.Id(), dbMap)
	if err != nil {
		return nil, err
	}

	return pk.ValueForPublicKey(publicKeyId, dbMap)
}

func (p project) SetCredential(key, value string, dbMap DataMapper) (ProjectCredentialKey, error) {
	// Figure out if the combo of key & p.Id exists
	pk, err := FindProjectCredentialKey(key, p.Id(), dbMap)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		pk = NewProjectCredentialKey(key, p.Id(), dbMap)
		if err := pk.Save(dbMap); err != nil {
			return nil, err
		}
	}

	// Now we need to go over each member of the project and create credential values for each member
	members, err := p.Members(dbMap)
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		u, err := m.User(dbMap)
		if err != nil {
			return nil, err
		}
		keys, err := u.ActivePublicKeys(dbMap)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		for _, k := range keys {
			cipher, err := k.Encrypt(value)
			if err != nil {
				return nil, err
			}
			pv, err := FindProjectCredentialValueForPublicKey(k.Id(), pk.Id(), dbMap)
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}
			if err == sql.ErrNoRows {
				currentTime := time.Now().UTC()
				pv = &projectCredentialValue{&projectCredentialValueCore{
					CredentialId: pk.Id(),
					MemberId:     m.Id(),
					PublicKeyId:  k.Id(),
					Cipher:       []byte(cipher),
					CreatedAt:    currentTime,
					UpdatedAt:    currentTime,
					ExpiresAt:    currentTime.AddDate(0, 3, 0),
				}}
			} else {
				pv.SetCipher([]byte(cipher))
			}
			if err := pv.Save(dbMap); err != nil {
				return nil, err
			}
		}
	}
	// Now we just return the key created
	return pk, nil
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

func FindProjectWithId(projectId int, dbMap DataMapper) (Project, error) {
	pc := &projectCore{Id: projectId}
	err := dbMap.SelectOne(pc, "SELECT * FROM projects WHERE id = ?", pc.Id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &project{pc}, nil
}

func NewProject(name, environment, defaultAccessLevel string) Project {
	if defaultAccessLevel == "" {
		defaultAccessLevel = ACCESS_LEVEL_READ
	}
	currentTime := time.Now().UTC()
	return &project{&projectCore{
		Name:               name,
		Environment:        environment,
		DefaultAccessLevel: defaultAccessLevel,
		CreatedAt:          currentTime,
		UpdatedAt:          currentTime,
	}}
}
