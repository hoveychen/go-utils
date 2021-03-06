// Package beanstalkd is a higher level wrapper to utilize beanstalkd as a simple message queue.
// It removes some good features like reserve-delete mechanism, reserve with timeout, etc.
// Offical beanstalk libary should be used for such features.
// NOTE: This client doesn't support concurrent Get/Put operations.
package beanstalkd

import (
	"fmt"
	"time"

	goutils "github.com/hoveychen/go-utils"
	"github.com/kr/beanstalk"
)

type Client struct {
	conn     *beanstalk.Conn
	tubeSet  *beanstalk.TubeSet
	tube     *beanstalk.Tube
	addr     string
	tubeName string
}

func Dial(addr string, tubeName string) *Client {
	c := &Client{}
	c.addr = addr
	c.tubeName = tubeName
	c.Reconnect()
	return c
}

func (c *Client) Reconnect() error {
	c.Close()

	conn, err := beanstalk.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	c.conn = conn
	c.tubeSet = beanstalk.NewTubeSet(conn, c.tubeName)
	c.tube = &beanstalk.Tube{conn, c.tubeName}

	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) Len() int {
	stats, err := c.tube.Stats()
	if err != nil {
		return -1
	}
	var ugent, ready int
	fmt.Sscanf(stats["current-jobs-urgent"], "%d", &ugent)
	fmt.Sscanf(stats["current-jobs-ready"], "%d", &ready)
	return ugent + ready
}

func (c *Client) isTimedOut(err error) bool {
	if err != nil {
		connErr, ok := err.(beanstalk.ConnError)
		if ok && connErr.Err == beanstalk.ErrTimeout {
			return true
		}
	}
	return false
}

func (c *Client) isConnectionLost(err error) bool {
	if err != nil {
		connErr, ok := err.(beanstalk.ConnError)
		if ok && connErr.Err.Error() == "EOF" {
			return true
		}
	}
	return false
}

// Get returns a job from a tube, blocking.
func (c *Client) Get() []byte {
	for {
		id, data, err := c.tubeSet.Reserve(time.Minute)
		if err != nil {
			if c.isTimedOut(err) {
				// Simply timed out. Retry again.
				continue
			}
			// Failed to connect to server.
			goutils.LogError("Beanstalk Connection:", err)
			// Holds for several seconds to wait for server recover.
			if c.isConnectionLost(err) {
				time.Sleep(5 * time.Second)
				c.Reconnect()
			}
			continue
		}

		if err := c.conn.Delete(id); err != nil {
			goutils.LogError(err, id)
		}

		return data
	}
}

// Get returns a job from a tube. Non-blocking. When no jobs, return nil.
func (c *Client) GetOrNull() []byte {
	id, data, err := c.tubeSet.Reserve(0)
	if err != nil {
		if c.isConnectionLost(err) {
			c.Reconnect()
		} else if !c.isTimedOut(err) {
			goutils.LogError(err)
		}
		return nil
	}

	if err := c.conn.Delete(id); err != nil {
		goutils.LogError(err, id)
	}

	return data
}

// Put the data into tube.
func (c *Client) Put(d []byte) error {
	return c.PutWithPriority(d, 1024)
}

func (c *Client) PutWithPriority(d []byte, pri uint32) error {
	_, err := c.tube.Put(d, pri, 0, time.Minute)
	return err
}
