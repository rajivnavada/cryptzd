package crypto

import (
	"time"
)

type projectCredentialValueCore struct {
	Id           int       `db:"id"`
	CredentialId int       `db:"credential_id"`
	MemberId     int       `db:"member_id"`
	PublicKeyId  int       `db:"public_key_id"`
	Cipher       []byte    `db:"cipher"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	ExpiresAt    time.Time `db:"expires_at"`
}

type projectCredentialValue struct {
	*projectCredentialValueCore
}

func (pv projectCredentialValue) Id() int {
	return pv.projectCredentialValueCore.Id
}

func (pv projectCredentialValue) CredentialId() int {
	return pv.projectCredentialValueCore.CredentialId
}

func (pv projectCredentialValue) MemberId() int {
	return pv.projectCredentialValueCore.MemberId
}

func (pv projectCredentialValue) PublicKeyId() int {
	return pv.projectCredentialValueCore.PublicKeyId
}

func (pv projectCredentialValue) Cipher() []byte {
	return pv.projectCredentialValueCore.Cipher
}

func (pv *projectCredentialValue) SetCipher(cipher []byte) {
	pv.projectCredentialValueCore.Cipher = cipher
	pv.projectCredentialValueCore.UpdatedAt = time.Now().UTC()
}

func (pv projectCredentialValue) CreatedAt() time.Time {
	return pv.projectCredentialValueCore.CreatedAt
}

func (pv projectCredentialValue) UpdatedAt() time.Time {
	return pv.projectCredentialValueCore.UpdatedAt
}

func (pv projectCredentialValue) ExpiresAt() time.Time {
	return pv.projectCredentialValueCore.ExpiresAt
}

func (pv projectCredentialValue) Save(dbMap *DataMapper) error {
	if pv.Id() > 0 {
		_, err := dbMap.Update(pv.projectCredentialValueCore)
		return err
	}
	return dbMap.Insert(pv.projectCredentialValueCore)
}

func NewProjectCredentialValue(credentialId, memberId, keyId int, cipher []byte) ProjectCredentialValue {
	currentTime := time.Now().UTC()
	return &projectCredentialValue{&projectCredentialValueCore{
		CredentialId: credentialId,
		MemberId:     memberId,
		PublicKeyId:  keyId,
		Cipher:       cipher,
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
		ExpiresAt:    currentTime.AddDate(1, 0, 0),
	}}
}

func FindProjectCredentialValueForPublicKey(publicKeyId, credentialId int, dbMap *DataMapper) (ProjectCredentialValue, error) {
	pkv := &projectCredentialValueCore{CredentialId: credentialId, PublicKeyId: publicKeyId}
	err := dbMap.SelectOne(pkv, "SELECT * FROM project_credential_values WHERE credential_id = ? AND public_key_id = ?", pkv.CredentialId, pkv.PublicKeyId)
	if err != nil {
		return nil, err
	}
	return &projectCredentialValue{pkv}, nil
}
