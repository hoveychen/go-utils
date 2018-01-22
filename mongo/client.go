// Package mongo provides handful wrapper to access and/or modify the mongo database.
package mongo

import (
	"sync"
	"time"

	"github.com/hoveychen/go-utils"
	"github.com/hoveychen/go-utils/flags"
	"github.com/hoveychen/go-utils/gomap"
	"gopkg.in/mgo.v2"
)

var (
	clientCache     = gomap.New()
	numDbConcurrent = flags.Int("numDbConcurrent", 10, "Concurrent socket to db")
)

type DbClient struct {
	session      *mgo.Session
	addr         string
	dbConcurrent chan struct{}
	dbWaitGroup  sync.WaitGroup
}

type DbSession struct {
	client *DbClient
	*mgo.Session
}

func Dial(addr string) *DbClient {
	cacheClient := clientCache.GetOrCreate(addr, func() interface{} {
		s, err := mgo.Dial(addr)
		if err != nil {
			goutils.LogFatal(addr, err)
		}
		s.SetMode(mgo.Eventual, true)

		c := &DbClient{}
		c.addr = addr
		c.session = s
		c.dbConcurrent = make(chan struct{}, *numDbConcurrent)

		go func() {
			for range time.Tick(time.Minute * 5) {
				err := c.session.Ping()
				if err != nil {
					goutils.LogError("Connection to", c.addr, "lost", err)
					c.session.Refresh()
				}
			}
		}()
		return c
	})
	return cacheClient.(*DbClient)
}

func (c *DbClient) Open(db, collection string) (*mgo.Collection, *DbSession) {
	c.dbConcurrent <- struct{}{}
	c.dbWaitGroup.Add(1)

	mgoSession := c.session.Copy()
	ownSession := &DbSession{}
	ownSession.Session = mgoSession
	ownSession.client = c

	return ownSession.DB(db).C(collection), ownSession
}

func (c *DbClient) OpenWithLongTimeout(db, collection string) (*mgo.Collection, *DbSession) {
	col, session := c.Open(db, collection)
	session.SetSocketTimeout(time.Minute * 30)
	session.SetCursorTimeout(0)
	return col, session
}

// Wait blocks the goroutine until all sessions are closed.
func (c *DbClient) Wait() {
	c.dbWaitGroup.Wait()
}

func (d *DbSession) Close() {
	d.Session.Close()
	<-d.client.dbConcurrent
	d.client.dbWaitGroup.Done()
}
