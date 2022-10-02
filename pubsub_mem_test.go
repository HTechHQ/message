package message_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/HTechHQ/message"
)

func TestMagic_mem_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("fail, if EventOrMessage is invalid", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		// invalid EventOrMessage parameters
		_, err := p.Subscribe(nil, validEmptyHandlerFunc)
		assert.Error(t, err)
		assert.Equal(t, message.ErrInvalidEvent, err)

		_, err = p.Subscribe("", validEmptyHandlerFunc)
		assert.Error(t, err)
		assert.Equal(t, message.ErrInvalidEvent, err)

		_, err = p.Subscribe(0, validEmptyHandlerFunc)
		assert.Error(t, err)
		assert.Equal(t, message.ErrInvalidEvent, err)

		_, err = p.Subscribe([]int{}, validEmptyHandlerFunc)
		assert.Error(t, err)
		assert.Equal(t, message.ErrInvalidEvent, err)

		_, err = p.Subscribe(struct{}{}, validEmptyHandlerFunc)
		assert.Error(t, err)
		assert.Equal(t, message.ErrInvalidEvent, err)

		_, err = p.Subscribe(struct{ name string }{}, validEmptyHandlerFunc)
		assert.Error(t, err)
		assert.Equal(t, message.ErrInvalidEvent, err)

		// valid EventOrMessage parameters
		_, err = p.Subscribe(simpleEvent{}, validEmptyHandlerFunc)
		assert.NoError(t, err)

		_, err = p.Subscribe(eventOrTopicStructEvent{}, validEmptyHandlerFunc)
		assert.NoError(t, err)

		_, err = p.Subscribe(validSliceEvent{}, validEmptyHandlerFunc)
		assert.NoError(t, err)
	})

	t.Run("subscribe via an implemented EventOrTopic-interface name", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		_, err := p.Subscribe(eventOrTopicStructEvent{}, validEmptyHandlerFunc)
		assert.NoError(t, err)
		assert.Equal(t, 1, p.NumSubs(eventOrTopicStructEvent{}))
	})

	t.Run("subscribe via an struct type as name", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		_, err := p.Subscribe(simpleEvent{}, validEmptyHandlerFunc)
		assert.NoError(t, err)
		assert.Equal(t, 1, p.NumSubs(simpleEvent{}))
	})

	t.Run("ensure HandlerFunc is func and signature has two arguments with first ctx and no returns", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		_, err := p.Subscribe(simpleEvent{}, 0)
		assert.Error(t, err)

		_, err = p.Subscribe(simpleEvent{}, "")
		assert.Error(t, err)

		_, err = p.Subscribe(simpleEvent{}, func() {})
		assert.Error(t, err)

		_, err = p.Subscribe(simpleEvent{}, func(msg []byte, str string) {})
		assert.Error(t, err)

		_, err = p.Subscribe(simpleEvent{}, func(i io.Reader, e simpleEvent) int { return 0 })
		assert.Error(t, err)

		_, err = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) int { return 0 })
		assert.Error(t, err)

		_, err = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {})
		assert.NoError(t, err)
	})

	t.Run("subscribe 3 handlers to the same topic", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		_, err0 := p.Subscribe(simpleEvent{}, func(ctx context.Context, msg []int) {})
		_, err1 := p.Subscribe(simpleEvent{}, func(ctx context.Context, msg []byte) {})
		_, err2 := p.Subscribe(simpleEvent{}, func(ctx context.Context, msg []string) {})

		assert.NoError(t, err0)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, 3, p.NumSubs(simpleEvent{}))
	})

	t.Run("subscribe safely concurrently", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		p := message.NewPubsubMem()

		wg.Add(wantedSubscribers)
		for i := 0; i < wantedSubscribers; i++ {
			go func(wg *sync.WaitGroup) {
				_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, msg []byte) {})
				wg.Done()
			}(&wg)
		}
		wg.Wait()

		assert.Equal(t, wantedSubscribers, p.NumSubs(simpleEvent{}))
	})
}

func TestMagic_mem_Publish(t *testing.T) {
	t.Parallel()

	t.Run("invalid EventOrMessage", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		err := p.Publish(ctx, "")
		assert.Error(t, err)
	})

	t.Run("single publish", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		p := message.NewPubsubMem()

		wg.Add(1)
		_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			assert.Equal(t, simpleEvent{}, e)
			wg.Done()
		})

		err := p.Publish(ctx, simpleEvent{})
		assert.NoError(t, err)

		wg.Wait()
	})

	t.Run("multiple subscribers one publish", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		p := message.NewPubsubMem()

		wg.Add(2)
		_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			assert.Equal(t, simpleEvent{}, e)
			wg.Done()
		})
		_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			assert.Equal(t, simpleEvent{}, e)
			wg.Done()
		})

		err := p.Publish(ctx, simpleEvent{})
		assert.NoError(t, err)

		wg.Wait()
	})

	t.Run("ensure ctx is handed over from publisher to subscriber", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		p := message.NewPubsubMem()

		wg.Add(1)
		_, _ = p.Subscribe(simpleStruct{}, func(ctx context.Context, e simpleStruct) {
			assert.Equal(t, ctxVal, ctx.Value(ctxKey))
			wg.Done()
		})

		ctxWithVal := context.WithValue(context.Background(), ctxKey, ctxVal)
		_ = p.Publish(ctxWithVal, simpleStruct{})

		wg.Wait()
	})

	t.Run("ensure structs are converted, once they have the same EventOrTopic name", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		now, _ := time.Parse(time.RFC822, "01 Jan 15 10:00 UTC")

		p := message.NewPubsubMem()

		wg.Add(4)
		_, _ = p.Subscribe(newUserRegisteredEvent{}, func(ctx context.Context, e newUserRegisteredEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, "test@example.com", e.Email)
			assert.Equal(t, []string{"s0", "s1"}, e.Settings)
			wg.Done()
		})
		_, _ = p.Subscribe(newUserAuditLogEvent{}, func(ctx context.Context, e newUserAuditLogEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, now, e.CreatedAt)
			wg.Done()
		})
		_, _ = p.Subscribe(newUserWelcomeEmailEvent{}, func(ctx context.Context, e newUserWelcomeEmailEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, "test@example.com", e.Email)
			wg.Done()
		})
		_, _ = p.Subscribe(shareNothingWithEvents{}, func(ctx context.Context, e shareNothingWithEvents) {
			assert.Equal(t, shareNothingWithEvents{}, e)
			wg.Done()
		})

		err := p.Publish(ctx, newUserRegisteredEvent{
			Username:  "user name",
			Email:     "test@example.com",
			CreatedAt: now,
			Settings:  []string{"s0", "s1"},
		})

		assert.Equal(t, nil, err)
		wg.Wait()
	})

	t.Run("publish safely concurrently", func(t *testing.T) {
		t.Parallel()

		var (
			wg sync.WaitGroup
			mu sync.Mutex
			n  int
		)

		p := message.NewPubsubMem()

		_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			mu.Lock()
			defer mu.Unlock()

			n++
			wg.Done()
		})

		wg.Add(wantedPublishers)
		for i := 0; i < wantedPublishers; i++ {
			go func() {
				_ = p.Publish(ctx, simpleEvent{})
			}()
		}

		wg.Wait()
		assert.Equal(t, wantedPublishers, n)
	})

	t.Run("publish safely concurrently while subscribers grow", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		p := message.NewPubsubMem()

		go func() { // subscribe while also publishing
			for i := 0; i < wantedSubscribers; i++ {
				_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
				})
			}
		}()

		wg.Add(wantedPublishers)
		for i := 0; i < wantedPublishers; i++ {
			go func() {
				_ = p.Publish(ctx, simpleEvent{})
				wg.Done()
			}()
		}

		wg.Wait()
	})
}

func TestMagic_mem_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("wait on shutdown for worker to finish", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		p := message.NewPubsubMem()

		wg.Add(2)
		_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			time.Sleep(20 * time.Millisecond)
			assert.Equal(t, simpleEvent{}, e)
			wg.Done()
		})
		_, _ = p.Subscribe(newUserRegisteredEvent{}, func(ctx context.Context, e newUserRegisteredEvent) {
			time.Sleep(20 * time.Millisecond)
			wg.Done()
		})

		_ = p.Publish(ctx, simpleEvent{})
		_ = p.Publish(ctx, newUserRegisteredEvent{Username: "user"})

		p.Shutdown(ctx)
		wg.Wait()
	})

	t.Run("shutdown with timeout", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		hasProcessed := false

		_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			time.Sleep(2 * time.Second)
			hasProcessed = true
		})

		_ = p.Publish(ctx, simpleEvent{})

		ctxShutDown, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
		p.Shutdown(ctxShutDown)

		assert.Equal(t, false, hasProcessed)
		cancel()
	})

	t.Run("call shutdown safely concurrently", func(t *testing.T) {
		t.Parallel()

		p := message.NewPubsubMem()

		hasProcessed := false

		_, _ = p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			time.Sleep(2 * time.Second)
			hasProcessed = true
		})

		_ = p.Publish(ctx, simpleEvent{})

		for i := 0; i < 5; i++ {
			go func(i int) {
				ctxShutDown, cancel := context.WithTimeout(context.Background(), time.Duration(i)*10*time.Millisecond)
				p.Shutdown(ctxShutDown)
				cancel()
			}(i)
		}

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		p.Shutdown(ctxShutDown)

		assert.Equal(t, false, hasProcessed)
		cancel()
	})
}

func TestSubscriber_Unsubscribe(t *testing.T) {
	t.Parallel()

	t.Run("unsubscribe", func(t *testing.T) {
		t.Parallel()

		var (
			wg sync.WaitGroup
			mu sync.Mutex
			n  int
		)

		p := message.NewPubsubMem()

		wg.Add(1)
		sub, _ := p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			mu.Lock()
			defer mu.Unlock()

			n++
			wg.Done()
		})

		_ = p.Publish(ctx, simpleEvent{})
		sub.Unsubscribe()
		_ = p.Publish(ctx, simpleEvent{})

		wg.Wait()
		assert.Equal(t, 1, n)
	})

	t.Run("unsubscribe safely concurrently", func(t *testing.T) {
		t.Parallel()

		var (
			wg sync.WaitGroup
			mu sync.Mutex
		)

		subs := make(map[int]*message.Subscriber, wantedSubscribers)

		p := message.NewPubsubMem()

		wg.Add(wantedSubscribers)
		for i := 0; i < wantedSubscribers; i++ {
			go func(w *sync.WaitGroup, i int) {
				mu.Lock()
				defer mu.Unlock()

				sub, _ := p.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {})
				subs[i] = sub

				w.Done()
			}(&wg, i)
		}
		wg.Wait()

		wg.Add(wantedSubscribers)
		for _, sub := range subs {
			go func(wg *sync.WaitGroup, sub *message.Subscriber) {
				sub.Unsubscribe()
				wg.Done()
			}(&wg, sub)
		}

		wg.Wait()
		assert.Equal(t, 0, p.NumSubs(simpleEvent{}))
	})
}

/*
	func BenchmarkMagic_mem_Publish(b *testing.B) {
		c := message.NewPubsubMem()

		_, _ = c.Subscribe(simpleEvent{}, func(ctx context.Context, e simpleEvent) {
			time.Sleep(20 * time.Millisecond) // slow worker
		})

		for i := 0; i < b.N; i++ {
			_ = c.Publish(ctx, simpleEvent{})
			// _ = c.Publish(ctx, newUserRegisteredEvent{})
		}
	}

	func BenchmarkThrouputMagicMem(b *testing.B) {
		const workers = 100

		topic := getRandomTopic()

		c := message.NewPubsubMem()

		for i := 0; i < workers; i++ {
			_, _ = c.Subscribe(topic(), func(msg []byte) {
				time.Sleep(200 * time.Millisecond) // slow worker
			})
		}

		for i := 0; i < b.N; i++ {
			go func() {
				_ = c.Publish(ctx, []byte(""))
			}()
		}
	}
*/
