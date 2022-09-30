package message_test

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/HTechHQ/message"
)

const (
	validTopic   = "1337"
	validMessage = "message"
)

func TestMagic_mem_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("subscribe 3 handlers to the same topic", func(t *testing.T) {
		t.Parallel()

		c := message.NewPubsubMem()

		_, err0 := c.Subscribe(validTopic, func(msg []int) {})
		_, err1 := c.Subscribe(validTopic, func(msg []byte) {})
		_, err2 := c.Subscribe(validTopic, func(msg []string) {})

		assert.NoError(t, err0)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, 3, c.NumSubs(validTopic))
	})

	t.Run("ensure HandlerFunc is func and signature has one argument and no returns", func(t *testing.T) {
		t.Parallel()

		c := message.NewPubsubMem()

		_, err := c.Subscribe(validTopic, 0)
		assert.Error(t, err)

		_, err = c.Subscribe(validTopic, func() {})
		assert.Error(t, err)

		_, err = c.Subscribe(validTopic, func(msg []byte, str string) {})
		assert.Error(t, err)

		_, err = c.Subscribe(validTopic, func(msg []byte) int { return 0 })
		assert.Error(t, err)
	})

	t.Run("subscribe safely concurrently", func(t *testing.T) {
		t.Parallel()

		const wantedSubscribers = 1000

		c := message.NewPubsubMem()

		var wg sync.WaitGroup

		wg.Add(wantedSubscribers)
		for i := 0; i < wantedSubscribers; i++ {
			go func(wg *sync.WaitGroup) {
				_, _ = c.Subscribe(validTopic, func(msg []byte) {})
				wg.Done()
			}(&wg)
		}
		wg.Wait()

		assert.Equal(t, wantedSubscribers, c.NumSubs(validTopic))
	})
}

func TestMagic_mem_Publish(t *testing.T) {
	t.Parallel()

	t.Run("single publish", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		c := message.NewPubsubMem()

		wg.Add(1)
		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			assert.Equal(t, []byte(validMessage), msg)
			wg.Done()
		})

		_ = c.Publish(validTopic, []byte(validMessage))

		wg.Wait()
	})

	t.Run("multiple subscribers one publish", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		c := message.NewPubsubMem()

		wg.Add(2)
		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			assert.Equal(t, []byte(validMessage), msg)
			wg.Done()
		})
		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			assert.Equal(t, []byte(validMessage), msg)
			wg.Done()
		})

		_ = c.Publish(validTopic, []byte(validMessage))

		wg.Wait()
	})

	t.Run("ensure automatic parameter casting", func(t *testing.T) {
		t.Parallel()

		var (
			wg sync.WaitGroup
			mu sync.Mutex
			b  bytes.Buffer
		)

		c := message.NewPubsubMem()

		wg.Add(2)
		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			mu.Lock()
			defer mu.Unlock()

			b.Write(msg)
			wg.Done()
		})
		_, _ = c.Subscribe(validTopic, func(msg []byte) {})

		err0 := c.Publish(validTopic, []byte("."))
		err1 := c.Publish(validTopic, ".")
		err2 := c.Publish(validTopic, 1)

		wg.Wait()

		assert.NoError(t, err0)
		assert.NoError(t, err1)

		assert.Error(t, err2)
		assert.Contains(t, err2.Error(), "2 errors occurred")

		assert.Equal(t, "..", b.String())
	})

	t.Run("ensure structs are converted", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		c := message.NewPubsubMem()

		now, _ := time.Parse(time.RFC822, "01 Jan 15 10:00 UTC")

		wg.Add(4)
		_, _ = c.Subscribe(validTopic, func(e newUserRegisteredEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, "test@example.com", e.Email)
			assert.Equal(t, []string{"s0", "s1"}, e.Settings)
			wg.Done()
		})
		_, _ = c.Subscribe(validTopic, func(e newUserAuditLogEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, now, e.CreatedAt)
			wg.Done()
		})
		_, _ = c.Subscribe(validTopic, func(e newUserWelcomeEmailEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, "test@example.com", e.Email)
			wg.Done()
		})
		_, _ = c.Subscribe(validTopic, func(e shareNothingWithEvents) {
			assert.Equal(t, shareNothingWithEvents{}, e)
			wg.Done()
		})

		err := c.Publish(validTopic, newUserRegisteredEvent{
			Username:  "user name",
			Email:     "test@example.com",
			CreatedAt: now,
			Settings:  []string{"s0", "s1"},
		})

		assert.Equal(t, nil, err)
		wg.Wait()
	})

	t.Run("ensure non primitiv conversion happens only between structs", func(t *testing.T) {
		t.Parallel()

		var (
			wg sync.WaitGroup
			b  bytes.Buffer
		)

		c := message.NewPubsubMem()

		wg.Add(0)
		_, _ = c.Subscribe(validTopic, func(e newUserWelcomeEmailEvent) {
			b.Write([]byte("called subscriber which should not happen"))
			wg.Done()
		})

		err := c.Publish(validTopic, []byte(validMessage))

		wg.Wait()
		assert.Error(t, err)
		assert.Equal(t, "", b.String())
	})

	t.Run("publish safely concurrently", func(t *testing.T) {
		t.Parallel()

		var (
			wg sync.WaitGroup
			b  safeBuffer
		)

		const wantedPublisher = 1000

		c := message.NewPubsubMem()

		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			_, _ = b.Write(msg)
			wg.Done()
		})

		wg.Add(wantedPublisher)
		for i := 0; i < wantedPublisher; i++ {
			go func() {
				_ = c.Publish(validTopic, []byte("."))
			}()
		}

		wg.Wait()
		assert.Equal(t, wantedPublisher, b.Len())
	})

	t.Run("publish safely concurrently while subscribers grow", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		const wantedPublisher = 1000
		const wantedSubscribers = 1000

		c := message.NewPubsubMem()

		go func() { // subscribe while also publishing
			for i := 0; i < wantedSubscribers; i++ {
				_, _ = c.Subscribe(validTopic, func(msg []byte) {
				})
			}
		}()

		wg.Add(wantedPublisher)
		for i := 0; i < wantedPublisher; i++ {
			go func() {
				_ = c.Publish(validTopic, []byte("."))
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func BenchmarkMagic_mem_Publish(b *testing.B) {
	c := message.NewPubsubMem()

	_, _ = c.Subscribe(validTopic, func(msg []byte) {
		time.Sleep(20 * time.Millisecond) // slow worker
	})

	for i := 0; i < b.N; i++ {
		_ = c.Publish(validTopic, []byte(""))
		_ = c.Publish(validTopic, newUserRegisteredEvent{})
	}
}

func TestMagic_mem_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("wait on shutdown for worker to finish", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		c := message.NewPubsubMem()

		wg.Add(2)
		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			time.Sleep(20 * time.Millisecond)
			assert.Equal(t, []byte(validMessage), msg)
			wg.Done()
		})
		_, _ = c.Subscribe(validTopic, func(e newUserRegisteredEvent) {
			time.Sleep(20 * time.Millisecond)
			wg.Done()
		})

		_ = c.Publish(validTopic, []byte(validMessage))
		_ = c.Publish(validTopic, newUserRegisteredEvent{Username: "user"})

		c.Shutdown(context.TODO())
		wg.Wait()
	})

	t.Run("shutdown with timeout", func(t *testing.T) {
		t.Parallel()

		c := message.NewPubsubMem()

		hasProcessed := false
		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			time.Sleep(2 * time.Second)
			hasProcessed = true
		})

		_ = c.Publish(validTopic, []byte(validMessage))

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		c.Shutdown(ctxShutDown)

		assert.Equal(t, false, hasProcessed)
		cancel()
	})

	t.Run("call shutdown safely concurrently", func(t *testing.T) {
		t.Parallel()

		c := message.NewPubsubMem()

		hasProcessed := false
		_, _ = c.Subscribe(validTopic, func(msg []byte) {
			time.Sleep(2 * time.Second)
			hasProcessed = true
		})

		_ = c.Publish(validTopic, []byte(validMessage))

		for i := 0; i < 5; i++ {
			go func(i int) {
				ctxShutDown, cancel := context.WithTimeout(context.Background(), time.Duration(i)*10*time.Millisecond)
				c.Shutdown(ctxShutDown)
				cancel()
			}(i)
		}

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		c.Shutdown(ctxShutDown)

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
			b  bytes.Buffer
		)

		c := message.NewPubsubMem()

		wg.Add(1)
		sub, _ := c.Subscribe(validTopic, func(msg []byte) {
			b.Write(msg)
			wg.Done()
		})

		_ = c.Publish(validTopic, []byte(validMessage))
		sub.Unsubscribe()
		_ = c.Publish(validTopic, []byte(validMessage))

		wg.Wait()
		assert.Equal(t, validMessage, b.String())
	})

	t.Run("unsubscribe safely concurrently", func(t *testing.T) {
		t.Parallel()

		var (
			wg sync.WaitGroup
			mu sync.Mutex
		)

		const wantedSubscribers = 1000

		subs := make(map[int]*message.Subscriber, wantedSubscribers)

		c := message.NewPubsubMem()

		wg.Add(wantedSubscribers)
		for i := 0; i < wantedSubscribers; i++ {
			go func(w *sync.WaitGroup, i int) {
				sub, _ := c.Subscribe(validTopic, func(msg []byte) {})

				mu.Lock()
				subs[i] = sub
				mu.Unlock()

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

		assert.Equal(t, 0, c.NumSubs(validTopic))
	})
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
			_ = c.Publish(topic(), []byte(""))
		}()
	}
}

func getRandomTopic() func() string {
	mu := sync.Mutex{}
	t := []string{"s", "e", "e", "d"}

	return func() string {
		mu.Lock()

		/*r := rand.Int()
		neuTopic := strconv.Itoa(r)

		if r%100 == 0 {
			t = append(t, neuTopic)
		}*/
		elem := t[rand.Intn(len(t))] //nolint:gosec

		mu.Unlock()

		return elem
	}
}

type newUserRegisteredEvent struct {
	Username  string    `json:"userName,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	Settings  []string  `json:"settings,omitempty"`
}

type newUserAuditLogEvent struct {
	Username  string    `json:"userName,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	Settings  []string  `json:"settings,omitempty"`
}

type newUserWelcomeEmailEvent struct {
	Username string `json:"userName,omitempty"`
	Email    string `json:"email,omitempty"`
}

type shareNothingWithEvents struct {
	Name string `json:"name,omitempty"`
}

// safeBuffer is a goroutine safe bytes.Buffer.
type safeBuffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.buffer.Write(p) //nolint:wrapcheck // ignore for this simple test helper.
}

// Len returns the length of the buffer.
func (s *safeBuffer) Len() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.buffer.Len()
}
