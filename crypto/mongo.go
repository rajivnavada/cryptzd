package crypto

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//----------------------------------------
// CONSTANTS AND VARIABLES
//----------------------------------------

var (
	NilSessionError error = errors.New("A mgo.Session instance is required but nil provided.")
)

// PUBLIC INTERFACE

type Indexer interface {
	AddUniqueIndex(collectionName string, fieldNames ...string) error

	Close()
}

type Identifiable interface {
	ObjectId() bson.ObjectId
}

type Session interface {
	HostName() string

	DbName() string

	Close()

	Save(doc interface{}, collectionName string) error

	Update(doc Identifiable, collectionName string) error

	Find(doc interface{}, selector bson.M, collectionName string) error

	FindAll(docs interface{}, pageLength uint8, pageNum int, selector bson.M, collectionName string, sortFields ...string) error
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

func (ms *mongoSession) AddUniqueIndex(collectionName string, fieldNames ...string) error {
	if err := ms.init(); err != nil {
		return err
	}
	return ms.DB(ms.dbName).C(collectionName).EnsureIndex(mgo.Index{
		Key:        fieldNames,
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     true,
	})
}

func (ms *mongoSession) Save(doc interface{}, collectionName string) error {
	if err := ms.init(); err != nil {
		return err
	}
	return ms.DB(ms.dbName).C(collectionName).Insert(doc)
}

func (ms *mongoSession) Update(doc Identifiable, collectionName string) error {
	if err := ms.init(); err != nil {
		return err
	}
	return ms.DB(ms.dbName).C(collectionName).UpdateId(doc.ObjectId(), doc)
}

func (ms *mongoSession) Find(container interface{}, selector bson.M, collectionName string) error {
	if err := ms.init(); err != nil {
		return err
	}
	return ms.DB(ms.dbName).C(collectionName).Find(selector).One(container)
}

func (ms *mongoSession) FindAll(container interface{}, pageLength uint8, pageNum int, selector bson.M, collectionName string, sortFields ...string) error {
	if err := ms.init(); err != nil {
		return err
	}
	if len(sortFields) == 0 {
		sortFields = []string{"$natural"}
	}

	skipNum := int(pageLength) * pageNum
	return ms.DB(ms.dbName).C(collectionName).Find(selector).Sort(sortFields...).Skip(skipNum).Limit(int(pageLength)).All(container)
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