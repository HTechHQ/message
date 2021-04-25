package message

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Subscriber struct {
	pubsub *PubsubMem
	topic  string
	h      HandlerFunc
}

// PubsubMem is a in memory implementation of the PubSub interface and uses goroutines
// to process the messages. It's methods are thread safe to use.
type PubsubMem struct {
	mu sync.Mutex
	wg sync.WaitGroup

	subs map[string][]*Subscriber
}

var _ PubSub = (*PubsubMem)(nil)

// NewPubsubMem initialises a new instance of a in memory implementation of the PubSub interface.
func NewPubsubMem() *PubsubMem {
	m := &PubsubMem{}
	m.subs = make(map[string][]*Subscriber)

	return m
}

func (mm *PubsubMem) Publish(topic string, msg Message) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	var err error

	for _, s := range mm.subs[topic] {
		handlerFuncType := reflect.TypeOf(s.h)
		paramType := handlerFuncType.In(0)

		if reflect.TypeOf(msg).ConvertibleTo(paramType) {
			mm.wg.Add(1)

			go func(h HandlerFunc, handlerFuncType reflect.Type, msg Message, paramType reflect.Type) {
				fn := reflect.ValueOf(h).Convert(handlerFuncType)
				fn.Call([]reflect.Value{reflect.ValueOf(msg).Convert(paramType)})
				mm.wg.Done()
			}(s.h, handlerFuncType, msg, paramType)

			continue
		}

		switch paramType.Kind() {
		case reflect.Struct:
			if reflect.TypeOf(msg).Kind() != reflect.Struct {
				break
			}

			b, err2 := json.Marshal(msg)
			if err2 != nil {
				break
			}

			param := reflect.New(paramType).Interface()

			err2 = json.Unmarshal(b, &param)
			if err2 != nil {
				break
			}

			mm.wg.Add(1)

			go func(h HandlerFunc, handlerFuncType reflect.Type, parsed interface{}) {
				fn := reflect.ValueOf(h).Convert(handlerFuncType)
				fn.Call([]reflect.Value{reflect.ValueOf(parsed).Elem()})
				mm.wg.Done()
			}(s.h, handlerFuncType, param)

			continue
		}

		err = multierror.Append(err, ErrParamConversionFailed)
	}

	return err
}

func (mm *PubsubMem) Subscribe(topic string, h HandlerFunc) (*Subscriber, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	handlerFunc := reflect.TypeOf(h)

	if handlerFunc.Kind() != reflect.Func {
		return nil, ErrInvalidHandler
	}

	if handlerFunc.NumIn() != 1 {
		return nil, ErrInvalidHandler
	}

	if handlerFunc.NumOut() != 0 {
		return nil, ErrInvalidHandler
	}

	s := &Subscriber{
		pubsub: mm,
		topic:  topic,
		h:      h,
	}

	mm.subs[topic] = append(mm.subs[topic], s)

	return s, nil
}

func (mm *PubsubMem) Shutdown(ctx context.Context) {
	waitCh := make(chan struct{})

	go func() {
		mm.wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
	case <-ctx.Done():
	}
}

// Unsubscribe will remove this subscriber from receiving new messages.
// Messages that are still processing remain untouched and continue to finish.
func (s *Subscriber) Unsubscribe() {
	s.pubsub.mu.Lock()
	defer s.pubsub.mu.Unlock()

	sTopic := s.topic
	for i, subscriber := range s.pubsub.subs[sTopic] {
		if subscriber == s {
			s.pubsub.subs[sTopic] = append(s.pubsub.subs[sTopic][:i], s.pubsub.subs[sTopic][i+1:]...)
		}
	}
}
