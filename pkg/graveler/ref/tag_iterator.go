package ref

import (
	"context"

	"github.com/treeverse/lakefs/pkg/graveler"
	"github.com/treeverse/lakefs/pkg/kv"
)

type KVTagIterator struct {
	it            kv.MessageIterator
	err           error
	value         *graveler.TagRecord
	repoPartition string
	store         kv.Store
	ctx           context.Context
}

func NewKVTagIterator(ctx context.Context, store *kv.StoreMessage, repo *graveler.RepositoryRecord) (*KVTagIterator, error) {
	repoPartition := graveler.RepoPartition(repo)
	it, err := kv.NewPrimaryIterator(ctx, store.Store, (&graveler.TagData{}).ProtoReflect().Type(),
		graveler.RepoPartition(repo),
		[]byte(graveler.TagPath("")), kv.IteratorOptionsFrom([]byte("")))
	if err != nil {
		return nil, err
	}
	return &KVTagIterator{
		it:            it,
		store:         store.Store,
		repoPartition: repoPartition,
		ctx:           ctx,
	}, nil
}

func (i *KVTagIterator) Next() bool {
	if i.Err() != nil {
		return false
	}
	if !i.it.Next() {
		i.value = nil
		return false
	}
	e := i.it.Entry()
	if e == nil {
		i.err = graveler.ErrReadingFromStore
		return false
	}
	tag, ok := e.Value.(*graveler.TagData)
	if !ok {
		i.err = graveler.ErrReadingFromStore
		return false
	}
	i.value = graveler.TagFromProto(tag)
	return true
}

func (i *KVTagIterator) SeekGE(id graveler.TagID) {
	if i.Err() == nil {
		i.it.Close()
		it, err := kv.NewPrimaryIterator(i.ctx, i.store, (&graveler.TagData{}).ProtoReflect().Type(),
			i.repoPartition,
			[]byte(graveler.TagPath("")), kv.IteratorOptionsFrom([]byte(graveler.TagPath(id))))
		i.it = it
		i.value = nil
		i.err = err
	}
}

func (i *KVTagIterator) Value() *graveler.TagRecord {
	if i.Err() != nil {
		return nil
	}
	return i.value
}

func (i *KVTagIterator) Err() error {
	if i.err == nil {
		return i.it.Err()
	}
	return i.err
}

func (i *KVTagIterator) Close() {
	i.it.Close()
}
