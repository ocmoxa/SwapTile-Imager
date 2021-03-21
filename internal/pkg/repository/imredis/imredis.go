// Package imredis contains implementations of the storage interfaces.
package imredis

import "github.com/gomodule/redigo/redis"

const (
	keyImageMeta     = "ocmoxa:image_meta"
	keyPrefixImageID = "ocmoxa:image_id:"
)

// pipeline helps to handle send error.
type pipeline struct {
	kv redis.Conn

	err error
}

func newPipeline(kv redis.Conn) *pipeline {
	return &pipeline{
		kv: kv,

		err: nil,
	}
}

func (p *pipeline) Send(commandName string, args ...interface{}) {
	if p.err != nil {
		return
	}

	p.err = p.kv.Send(commandName, args...)
}

func (p *pipeline) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if p.err != nil {
		return nil, p.err
	}

	return p.kv.Do(commandName, args...)
}
