package message

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	topic   = "1337"
	message = "message"
)

func TestMagic_mem_Subscribe(t *testing.T) {
	t.Run("subscribe 3 handlers to the same topic", func(t *testing.T) {
		c := NewPubsubMem()

		_, err0 := c.Subscribe(topic, func(msg []int) {})
		_, err1 := c.Subscribe(topic, func(msg []byte) {})
		_, err2 := c.Subscribe(topic, func(msg []string) {})

		assert.NoError(t, err0)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, 3, len(c.subs[topic]))
	})

	t.Run("ensure HandlerFunc is func and signature has one argument and no returns", func(t *testing.T) {
		c := NewPubsubMem()

		_, err := c.Subscribe(topic, 0)
		assert.Error(t, err)

		_, err = c.Subscribe(topic, func() {})
		assert.Error(t, err)

		_, err = c.Subscribe(topic, func(msg []byte, str string) {})
		assert.Error(t, err)

		_, err = c.Subscribe(topic, func(msg []byte) int { return 0 })
		assert.Error(t, err)
	})

	t.Run("subscribe safely concurrently", func(t *testing.T) {
		const wantedSubscribers = 1000
		c := NewPubsubMem()

		var wg sync.WaitGroup

		wg.Add(wantedSubscribers)
		for i := 0; i < wantedSubscribers; i++ {
			go func(wg *sync.WaitGroup) {
				_, _ = c.Subscribe(topic, func(msg []byte) {})
				wg.Done()
			}(&wg)
		}
		wg.Wait()

		assert.Equal(t, wantedSubscribers, len(c.subs[topic]))
	})
}

func TestMagic_mem_Publish(t *testing.T) {
	var wg sync.WaitGroup

	t.Run("single publish", func(t *testing.T) {
		c := NewPubsubMem()

		wg.Add(1)
		_, _ = c.Subscribe(topic, func(msg []byte) {
			assert.Equal(t, []byte(message), msg)
			wg.Done()
		})

		_ = c.Publish(topic, []byte(message))

		wg.Wait()
	})

	t.Run("multiple subscribers one publish", func(t *testing.T) {
		c := NewPubsubMem()

		wg.Add(2)
		_, _ = c.Subscribe(topic, func(msg []byte) {
			assert.Equal(t, []byte(message), msg)
			wg.Done()
		})
		_, _ = c.Subscribe(topic, func(msg []byte) {
			assert.Equal(t, []byte(message), msg)
			wg.Done()
		})

		_ = c.Publish(topic, []byte(message))

		wg.Wait()
	})

	t.Run("ensure automatic parameter casting", func(t *testing.T) {
		var b bytes.Buffer
		var mu sync.Mutex

		c := NewPubsubMem()

		wg.Add(2)
		_, _ = c.Subscribe(topic, func(msg []byte) {
			mu.Lock()
			defer mu.Unlock()

			b.Write(msg)
			wg.Done()
		})
		_, _ = c.Subscribe(topic, func(msg []byte) {})

		err0 := c.Publish(topic, []byte("."))
		err1 := c.Publish(topic, ".")
		err2 := c.Publish(topic, 1)

		wg.Wait()

		assert.NoError(t, err0)
		assert.NoError(t, err1)

		assert.Error(t, err2)
		assert.Contains(t, err2.Error(), "2 errors occurred")

		assert.Equal(t, "..", b.String())
	})

	t.Run("ensure structs are converted", func(t *testing.T) {
		c := NewPubsubMem()

		now, _ := time.Parse(time.RFC822, "01 Jan 15 10:00 UTC")

		wg.Add(4)
		_, _ = c.Subscribe(topic, func(e newUserRegisteredEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, "test@example.com", e.Email)
			assert.Equal(t, []string{"s0", "s1"}, e.Settings)
			wg.Done()
		})
		_, _ = c.Subscribe(topic, func(e newUserAuditLogEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, now, e.CreatedAt)
			wg.Done()
		})
		_, _ = c.Subscribe(topic, func(e newUserWelcomeEmailEvent) {
			assert.Equal(t, "user name", e.Username)
			assert.Equal(t, "test@example.com", e.Email)
			wg.Done()
		})
		_, _ = c.Subscribe(topic, func(e shareNothingWithEvents) {
			assert.Equal(t, shareNothingWithEvents{}, e)
			wg.Done()
		})

		err := c.Publish(topic, newUserRegisteredEvent{
			Username:  "user name",
			Email:     "test@example.com",
			CreatedAt: now,
			Settings:  []string{"s0", "s1"},
		})

		assert.Equal(t, nil, err)
		wg.Wait()
	})

	t.Run("ensure non primitiv conversion happens only between structs", func(t *testing.T) {
		var b bytes.Buffer

		c := NewPubsubMem()

		wg.Add(0)
		_, _ = c.Subscribe(topic, func(e newUserWelcomeEmailEvent) {
			b.Write([]byte("called subscriber which should not happen"))
			wg.Done()
		})

		err := c.Publish(topic, []byte(message))

		wg.Wait()
		assert.Error(t, err)
		assert.Equal(t, "", b.String())
	})

	t.Run("publish safely concurrently", func(t *testing.T) {
		const wantedPublisher = 10000
		var b safeBuffer

		c := NewPubsubMem()

		_, _ = c.Subscribe(topic, func(msg []byte) {
			_, _ = b.Write(msg)
			wg.Done()
		})

		wg.Add(wantedPublisher)
		for i := 0; i < wantedPublisher; i++ {
			go func() {
				_ = c.Publish(topic, []byte("."))
			}()
		}

		wg.Wait()
		assert.Equal(t, wantedPublisher, b.Len())
	})

	t.Run("publish safely concurrently while subscribers grow", func(t *testing.T) {
		const wantedPublisher = 1000
		const wantedSubscribers = 1000

		c := NewPubsubMem()

		go func() { // subscribe while also publishing
			for i := 0; i < wantedSubscribers; i++ {
				_, _ = c.Subscribe(topic, func(msg []byte) {
				})
			}
		}()

		wg.Add(wantedPublisher)
		for i := 0; i < wantedPublisher; i++ {
			go func() {
				_ = c.Publish(topic, []byte("."))
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func BenchmarkMagic_mem_Publish(b *testing.B) {
	c := NewPubsubMem()

	_, _ = c.Subscribe(topic, func(msg []byte) {
		time.Sleep(20 * time.Millisecond) // slow worker
	})

	for i := 0; i < b.N; i++ {
		_ = c.Publish(topic, []byte(""))
		_ = c.Publish(topic, newUserRegisteredEvent{})
	}
}

func TestMagic_mem_Shutdown(t *testing.T) {
	t.Run("wait on shutdown for worker to finish", func(t *testing.T) {
		var wg sync.WaitGroup

		c := NewPubsubMem()

		wg.Add(2)
		_, _ = c.Subscribe(topic, func(msg []byte) {
			time.Sleep(20 * time.Millisecond)
			assert.Equal(t, []byte(message), msg)
			wg.Done()
		})
		_, _ = c.Subscribe(topic, func(e newUserRegisteredEvent) {
			time.Sleep(20 * time.Millisecond)
			wg.Done()
		})

		_ = c.Publish(topic, []byte(message))
		_ = c.Publish(topic, newUserRegisteredEvent{Username: "user"})

		c.Shutdown(context.TODO())
		wg.Wait()
	})

	t.Run("shutdown with timeout", func(t *testing.T) {
		c := NewPubsubMem()

		hasProcessed := false
		_, _ = c.Subscribe(topic, func(msg []byte) {
			time.Sleep(20 * time.Second)
			hasProcessed = true
		})

		_ = c.Publish(topic, []byte(message))

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		c.Shutdown(ctxShutDown)

		assert.Equal(t, false, hasProcessed)
		cancel()
	})

	t.Run("call shutdown safely concurrently", func(t *testing.T) {
		c := NewPubsubMem()

		hasProcessed := false
		_, _ = c.Subscribe(topic, func(msg []byte) {
			time.Sleep(20 * time.Second)
			hasProcessed = true
		})

		_ = c.Publish(topic, []byte(message))

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
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	t.Run("unsubscribe", func(t *testing.T) {
		var b bytes.Buffer

		c := NewPubsubMem()

		wg.Add(1)
		sub, _ := c.Subscribe(topic, func(msg []byte) {
			b.Write(msg)
			wg.Done()
		})

		_ = c.Publish(topic, []byte(message))
		sub.Unsubscribe()
		_ = c.Publish(topic, []byte(message))

		wg.Wait()
		assert.Equal(t, message, b.String())
	})

	t.Run("unsubscribe safely concurrently", func(t *testing.T) {
		const wantedSubscribers = 1000

		subs := make(map[int]*Subscriber, wantedSubscribers)

		c := NewPubsubMem()

		wg.Add(wantedSubscribers)
		for i := 0; i < wantedSubscribers; i++ {
			go func(w *sync.WaitGroup, i int) {
				sub, _ := c.Subscribe(topic, func(msg []byte) {})

				mu.Lock()
				subs[i] = sub
				mu.Unlock()

				w.Done()
			}(&wg, i)
		}
		wg.Wait()

		wg.Add(wantedSubscribers)
		for _, sub := range subs {
			go func(wg *sync.WaitGroup, sub *Subscriber) {
				sub.Unsubscribe()
				wg.Done()
			}(&wg, sub)
		}
		wg.Wait()

		assert.Equal(t, 0, len(c.subs[topic]))
	})
}

func BenchmarkThrouputMagicMem(b *testing.B) {
	const workers = 100

	topic := getRandomTopic()

	c := NewPubsubMem()

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
	Username  string    `json:"user_name,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Settings  []string  `json:"settings,omitempty"`
}

type newUserAuditLogEvent struct {
	Username  string    `json:"user_name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Settings  []string  `json:"settings,omitempty"`
}

type newUserWelcomeEmailEvent struct {
	Username string `json:"user_name,omitempty"`
	Email    string `json:"email,omitempty"`
}

type shareNothingWithEvents struct {
	Name string `json:"name,omitempty"`
}

// safeBuffer is a goroutine safe bytes.Buffer
type safeBuffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.buffer.Write(p)
}

// Len returns the length of the buffer.
func (s *safeBuffer) Len() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.buffer.Len()
}
