package mongo

import (
	"errors"
	"gopkg.in/mgo.v2"
)

//----------------------------------------
// CONSTANTS AND VARIABLES
//----------------------------------------

const (
	SEEK_START int = 0
	SEEK_CUR   int = 1
	SEEK_END   int = 2
)

var (
	NilSessionError error = errors.New("A mgo.Session instance is required but nil provided.")
)

// PUBLIC INTERFACE

type Indexer interface {
	AddUniqueIndex(collectionName, fieldName string) error

	Close()
}

type Session interface {
	HostName() string

	DbName() string

	SaveDocument(doc interface{}, collectionName string) error

	Close()
}

//----------------------------------------
// WRAP mongo session with our version so we can define new behavior on it
//----------------------------------------

type mongoSession struct {
	*mgo.Session

	hostName string

	dbName string
}

func (ms *mongoSession) init() error {
	if ms.Session == nil {
		var err error
		ms.Session, err = mgo.Dial(ms.hostName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *mongoSession) AddUniqueIndex(collectionName, fieldName string) error {
	if err := ms.init(); err != nil {
		return err
	}
	return ms.DB(ms.dbName).C(collectionName).EnsureIndex(mgo.Index{
		Key:        []string{fieldName},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     true,
	})
}

func (ms *mongoSession) SaveDocument(doc interface{}, collectionName string) error {
	if err := ms.init(); err != nil {
		return err
	}
	return ms.DB(ms.dbName).C(collectionName).Insert(doc)
}

func (ms *mongoSession) HostName() string {
	return ms.hostName
}

func (ms *mongoSession) DbName() string {
	return ms.dbName
}

func (ms *mongoSession) Close() {
	if ms.Session == nil {
		return
	}
	ms.Session.Close()
}

// PUBLIC FUNCTION

func NewSession(hostName, dbName string) Session {
	return &mongoSession{
		hostName: hostName,
		dbName:   dbName,
	}
}

func NewIndexer(hostName, dbName string) Indexer {
	return &mongoSession{
		hostName: hostName,
		dbName:   dbName,
	}
}
