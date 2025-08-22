// Copyright 2013, Chandra Sekar S. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package pubsub implements a simple multi-topic pub-sub
// library.
//
// A topic can have any number of subscribers. All subscribers receive messages
// published on the topic.
package pubsub

type operation int

const (
	sub operation = iota
	subOnce
	subOnceEach
	pub
	tryPub
	unsub
	unsubAll
	closeTopic
	shutdown
)

// PubSub is a collection of topics.
type PubSub[T comparable, M any] struct {
	cmdChan  chan cmd[T, M]
	capacity int
}

type cmd[T comparable, M any] struct {
	msg    M
	ch     chan M
	topics []T
	op     operation
}

// New creates a new PubSub and starts a goroutine for handling operations. Sub
// and SubOnce will create channels with the given capacity.
func New[T comparable, M any](capacity int) *PubSub[T, M] {
	ps := &PubSub[T, M]{make(chan cmd[T, M]), capacity}
	go ps.start()
	return ps
}

// Sub returns a channel from which messages published on the specified topics
// can be received.
func (ps *PubSub[T, M]) Sub(topics ...T) chan M {
	return ps.sub(sub, topics...)
}

// SubOnce is similar to Sub, but only the first message published, after subscription,
// on any of the specified topics can be received.
func (ps *PubSub[T, M]) SubOnce(topics ...T) chan M {
	return ps.sub(subOnce, topics...)
}

// SubOnceEach returns a channel on which callers receive, at most, one message
// for each topic.
func (ps *PubSub[T, M]) SubOnceEach(topics ...T) chan M {
	return ps.sub(subOnceEach, topics...)
}

func (ps *PubSub[T, M]) sub(op operation, topics ...T) chan M {
	ch := make(chan M, ps.capacity)
	ps.cmdChan <- cmd[T, M]{op: op, topics: topics, ch: ch}
	return ch
}

// AddSub adds subscriptions to an existing channel.
func (ps *PubSub[T, M]) AddSub(ch chan M, topics ...T) {
	ps.cmdChan <- cmd[T, M]{op: sub, topics: topics, ch: ch}
}

// AddSubOnceEach adds subscriptions to an existing channel with SubOnceEach
// behavior.
func (ps *PubSub[T, M]) AddSubOnceEach(ch chan M, topics ...T) {
	ps.cmdChan <- cmd[T, M]{op: subOnceEach, topics: topics, ch: ch}
}

// Pub publishes the given message to all subscribers of the specified topics.
func (ps *PubSub[T, M]) Pub(msg M, topics ...T) {
	ps.cmdChan <- cmd[T, M]{op: pub, topics: topics, msg: msg}
}

// TryPub publishes the given message to all subscribers of the specified topics
// if the topic has buffer space.
func (ps *PubSub[T, M]) TryPub(msg M, topics ...T) {
	ps.cmdChan <- cmd[T, M]{op: tryPub, topics: topics, msg: msg}
}

// Unsub unsubscribes the given channel from the specified topics. If no topic
// is specified, it is unsubscribed from all topics.
//
// Unsub must be called from a goroutine that is different from the subscriber.
// The subscriber must consume messages from the channel until it reaches the
// end. Not doing so can result in a deadlock.
func (ps *PubSub[T, M]) Unsub(ch chan M, topics ...T) {
	if len(topics) == 0 {
		ps.cmdChan <- cmd[T, M]{op: unsubAll, ch: ch}
		return
	}

	ps.cmdChan <- cmd[T, M]{op: unsub, topics: topics, ch: ch}
}

// Close closes all channels currently subscribed to the specified topics. If a
// channel is subscribed to multiple topics, some of which is not specified, it
// is not closed.
func (ps *PubSub[T, M]) Close(topics ...T) {
	ps.cmdChan <- cmd[T, M]{op: closeTopic, topics: topics}
}

// Shutdown closes all subscribed channels and terminates the goroutine.
func (ps *PubSub[T, M]) Shutdown() {
	ps.cmdChan <- cmd[T, M]{op: shutdown}
}

func (ps *PubSub[T, M]) start() {
	reg := registry[T, M]{
		topics:    make(map[T]map[chan M]subType),
		revTopics: make(map[chan M]map[T]bool),
	}

loop:
	for cmd := range ps.cmdChan {
		if cmd.topics == nil {
			switch cmd.op {
			case unsubAll:
				reg.removeChannel(cmd.ch)

			case shutdown:
				break loop
			}

			continue loop
		}

		for _, topic := range cmd.topics {
			switch cmd.op {
			case sub:
				reg.add(topic, cmd.ch, normal)

			case subOnce:
				reg.add(topic, cmd.ch, onceAny)

			case subOnceEach:
				reg.add(topic, cmd.ch, onceEach)

			case tryPub:
				reg.sendNoWait(topic, cmd.msg)

			case pub:
				reg.send(topic, cmd.msg)

			case unsub:
				reg.remove(topic, cmd.ch)

			case closeTopic:
				reg.removeTopic(topic)
			}
		}
	}

	for topic, chans := range reg.topics {
		for ch := range chans {
			reg.remove(topic, ch)
		}
	}
}

// registry maintains the current subscription state. It is not safe to access a
// registry from multiple goroutines concurrently.
type registry[T comparable, M any] struct {
	topics    map[T]map[chan M]subType
	revTopics map[chan M]map[T]bool
}

type subType int

const (
	onceAny subType = iota
	onceEach
	normal
)

func (reg *registry[T, M]) add(topic T, ch chan M, st subType) {
	if reg.topics[topic] == nil {
		reg.topics[topic] = make(map[chan M]subType)
	}
	reg.topics[topic][ch] = st

	if reg.revTopics[ch] == nil {
		reg.revTopics[ch] = make(map[T]bool)
	}
	reg.revTopics[ch][topic] = true
}

func (reg *registry[T, M]) send(topic T, msg M) {
	for ch, st := range reg.topics[topic] {
		ch <- msg
		switch st {
		case onceAny:
			for topic := range reg.revTopics[ch] {
				reg.remove(topic, ch)
			}
		case onceEach:
			reg.remove(topic, ch)
		}
	}
}

func (reg *registry[T, M]) sendNoWait(topic T, msg M) {
	for ch, st := range reg.topics[topic] {
		select {
		case ch <- msg:
			switch st {
			case onceAny:
				for topic := range reg.revTopics[ch] {
					reg.remove(topic, ch)
				}
			case onceEach:
				reg.remove(topic, ch)
			}
		default:
		}

	}
}

func (reg *registry[T, M]) removeTopic(topic T) {
	for ch := range reg.topics[topic] {
		reg.remove(topic, ch)
	}
}

func (reg *registry[T, M]) removeChannel(ch chan M) {
	for topic := range reg.revTopics[ch] {
		reg.remove(topic, ch)
	}
}

func (reg *registry[T, M]) remove(topic T, ch chan M) {
	if _, ok := reg.topics[topic]; !ok {
		return
	}

	if _, ok := reg.topics[topic][ch]; !ok {
		return
	}

	delete(reg.topics[topic], ch)
	delete(reg.revTopics[ch], topic)

	if len(reg.topics[topic]) == 0 {
		delete(reg.topics, topic)
	}

	if len(reg.revTopics[ch]) == 0 {
		close(ch)
		delete(reg.revTopics, ch)
	}
}
