package crypto

import (
	"time"
)

type projectMemberCore struct {
	Id          int       `db:"id"`
	ProjectId   int       `db:"project_id"`
	UserId      int       `db:"user_id"`
	AccessLevel string    `db:"access_level"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type projectMember struct {
	*projectMemberCore
}

func (pm projectMember) Id() int {
	return pm.projectMemberCore.Id
}

func (pm projectMember) ProjectId() int {
	return pm.projectMemberCore.ProjectId
}

func (pm projectMember) UserId() int {
	return pm.projectMemberCore.UserId
}

func (pm projectMember) AccessLevel() string {
	return pm.projectMemberCore.AccessLevel
}

func (pm projectMember) CreatedAt() time.Time {
	return pm.projectMemberCore.CreatedAt
}

func (pm projectMember) UpdatedAt() time.Time {
	return pm.projectMemberCore.UpdatedAt
}

func (pm projectMember) User(dbMap DataMapper) (User, error) {
	return FindUserWithId(pm.UserId(), dbMap)
}

func (pm projectMember) Save(dbMap DataMapper) error {
	if pm.Id() > 0 {
		_, err := dbMap.Update(pm.projectMemberCore)
		return err
	}
	return dbMap.Insert(pm.projectMemberCore)
}

func (pm projectMember) Delete(dbMap DataMapper) error {
	_, err := dbMap.Delete(pm.projectMemberCore)
	return err
}

func FindProjectMemberWithUserId(userId, projectId int, dbMap DataMapper) (ProjectMember, error) {
	pm := &projectMemberCore{UserId: userId, ProjectId: projectId}
	err := dbMap.SelectOne(pm, "SELECT * FROM project_members WHERE user_id = ? AND project_id = ?", pm.UserId, pm.ProjectId)
	if err != nil {
		return nil, err
	}
	return &projectMember{pm}, nil
}

func FindProjectMemberWithId(memberId int, dbMap DataMapper) (ProjectMember, error) {
	pm := &projectMemberCore{Id: memberId}
	err := dbMap.SelectOne(pm, "SELECT * FROM project_members WHERE id = ?", memberId)
	if err != nil {
		return nil, err
	}
	return &projectMember{pm}, nil
}

func NewProjectMember(userId, projectId int, accessLevel string) ProjectMember {
	currentTime := time.Now().UTC()
	return &projectMember{&projectMemberCore{
		UserId:      userId,
		ProjectId:   projectId,
		AccessLevel: accessLevel,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}}
}
