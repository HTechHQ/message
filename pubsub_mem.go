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
	topic  EventOrTopicName
	h      HandlerFunc
}

// PubsubMem is a in memory implementation of the PubSub interface and uses goroutines
// to process the messages. It's methods are thread safe to use.
type PubsubMem struct {
	mu sync.Mutex
	wg sync.WaitGroup

	subs map[EventOrTopicName][]*Subscriber
}

var _ PubSub = (*PubsubMem)(nil)

// NewPubsubMem initialises a new instance of an in memory implementation of the PubSub interface.
func NewPubsubMem() *PubsubMem {
	m := &PubsubMem{
		mu:   sync.Mutex{},
		wg:   sync.WaitGroup{},
		subs: make(map[EventOrTopicName][]*Subscriber),
	}

	return m
}

func (mm *PubsubMem) NumSubs(eom EventOrMessage) int {
	eot, _ := getEventOrTopicName(eom)

	return len(mm.subs[eot])
}

func (mm *PubsubMem) Publish(ctx context.Context, msg EventOrMessage) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	topic, err := getEventOrTopicName(msg)
	if err != nil {
		return err
	}

	var mErr *multierror.Error

	for _, s := range mm.subs[topic] {
		handlerFuncType := reflect.TypeOf(s.h)
		paramType := handlerFuncType.In(1)

		if reflect.TypeOf(msg).ConvertibleTo(paramType) {
			mm.wg.Add(1)

			go func(h HandlerFunc, handlerFuncType reflect.Type, msg EventOrMessage, paramType reflect.Type) {
				fn := reflect.ValueOf(h).Convert(handlerFuncType)
				fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg).Convert(paramType)})
				mm.wg.Done()
			}(s.h, handlerFuncType, msg, paramType)

			continue
		}

		switch paramType.Kind() { //nolint:exhaustive
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
				fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(parsed).Elem()})
				mm.wg.Done()
			}(s.h, handlerFuncType, param)

			continue
		}

		mErr = multierror.Append(mErr, ErrParamConversionFailed)
	}

	return mErr.ErrorOrNil() //nolint:wrapcheck // allow the multierror to be returned.
}

func (mm *PubsubMem) Subscribe(eom EventOrMessage, h HandlerFunc) (*Subscriber, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	topic, err := getEventOrTopicName(eom)
	if err != nil {
		return nil, err
	}

	handlerFunc := reflect.TypeOf(h)

	// ensure handlerFunc is indeed a function
	if handlerFunc.Kind() != reflect.Func {
		return nil, ErrInvalidHandler
	}

	// ensure handlerFunc has two parameters
	if handlerFunc.NumIn() != 2 { //nolint:gomnd
		return nil, ErrInvalidHandler
	}

	// ensure the first parameter is of type context.Context
	if handlerFunc.In(0).Kind() != reflect.Interface {
		return nil, ErrInvalidHandler
	}

	if handlerFunc.In(0).PkgPath() != "context" && handlerFunc.In(0).Name() != "Context" {
		return nil, ErrInvalidHandler
	}

	// ensure handlerFunc has no return parameters
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

// getEventOrTopicName returns a name for an event or message topic.
// If the parameter implements the EventOrTopic interface, then that name is returned as the EventOrTopicName.
// Otherwise, it is expected, that the EventOrMessage is a struct and the name of that type is returned.
func getEventOrTopicName(eom EventOrMessage) (EventOrTopicName, error) {
	if eom == nil {
		return "", ErrInvalidEvent
	}

	eot, ok := eom.(EventOrTopic)
	if ok {
		return eot.EventOrTopicName(), nil
	}

	// ensure primitive data types like string or int are not allowed
	typeOf := reflect.TypeOf(eom)
	if typeOf.Kind() != reflect.Struct {
		return "", ErrInvalidEvent
	}

	structTypeName := reflect.TypeOf(eom).Name()
	if structTypeName == "" {
		return "", ErrInvalidEvent
	}

	return EventOrTopicName(structTypeName), nil
}
