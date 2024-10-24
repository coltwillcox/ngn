package conductor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cmdBufSize     = 10
	maxNestedCalls = 10
)

var (
	goModRe = regexp.MustCompile(`^(git.sr.ht/~blallo/|/.*/)conductor/[a-z_]+\.go$`)
	testRe  = regexp.MustCompile(`^(git.sr.ht/~blallo/|/.*/)conductor/[a-z_]+\_test.go$`)
)

type simple[T any] struct {
	listeners map[string]chan T
	mu        sync.RWMutex
	ctx       context.Context
	logFile   *os.File
}

/* Implement context.Context */

var _ context.Context = &simple[struct{}]{}

func (c *simple[T]) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

func (c *simple[T]) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *simple[T]) Err() error {
	return c.ctx.Err()
}

func (c *simple[T]) Value(key any) any {
	return c.ctx.Value(key)
}

/* Implement Conductor[T] */

func (c *simple[T]) Cmd() <-chan T {
	return c.cmd(2)
}

func (c *simple[T]) WithContext(ctx context.Context) Conductor[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ctx = ctx
	return c
}

func (c *simple[T]) WithContextPolicy(policy Policy[T]) Conductor[T] {
	go func() {
		<-c.ctx.Done()
		if cmd, ok := policy.Decide(); ok {
			c.send(cmd)
		}
	}()

	return c
}

/* Internal functions */

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func (c *simple[T]) cmd(level int, discriminator ...any) <-chan T {
	pc := make([]uintptr, maxNestedCalls)
	n := runtime.Callers(level, pc)
	if n == 0 {
		fmt.Fprintln(c.logFile, "Cannot find caller")
		// XXX: we return a closed channel, as we are not able to properly return
		// a valid channel without leaking it for each case statement evaluation.
		ch := make(chan T)
		close(ch)
		return ch
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)
	var file string
	var line int
	var progCounter uintptr
	// XXX: here we unwind the stack until we exit the conductor/ package.
frameLoop:
	for {
		frame, more := frames.Next()
		//fmt.Fprintf(c.logFile, "unwinding stack -> %s:%d:%d\n", frame.File, frame.Line, frame.PC)

		if testRe.MatchString(frame.File) || !goModRe.MatchString(frame.File) {
			file = frame.File
			line = frame.Line
			progCounter = frame.PC
			break frameLoop
		}

		if !more {
			fmt.Fprintln(c.logFile, "Cannot find caller")
			// XXX: we return a closed channel, as we are not able to properly return
			// a valid channel without leaking it for each case statement evaluation.
			ch := make(chan T)
			close(ch)
			return ch
		}
	}

	key := fmt.Sprintf("%s:%d:%d:%d", file, line, progCounter, goid())
	for _, dis := range discriminator {
		key = fmt.Sprintf("%s:%s", key, fmt.Sprint(dis))
	}

	fmt.Fprintln(c.logFile, "cmd key ->", key)

	c.mu.RLock()
	if ch, ok := c.listeners[key]; ok {
		c.mu.RUnlock()
		return ch
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	lis := make(chan T, cmdBufSize)
	c.listeners[key] = lis
	return lis
}

func (c *simple[T]) send(cmd T) {
	c.mu.RLock()
	for k, ch := range c.listeners {
		fmt.Fprintf(c.logFile, "Sending %s to %s listener\n", fmtCmd(cmd), k)
		ch <- cmd
	}
	c.mu.RUnlock()
}

func (c *simple[T]) notify(cmd T, signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	for {
		select {
		case <-c.Done():
			return
		case <-ch:
			go c.send(cmd)
		}
	}
}

/* Public functions */

// Simple creates a [Conductor] with a single type of listener.
func Simple[T any]() Conductor[T] {
	return &simple[T]{
		ctx:       context.TODO(),
		logFile:   initLogFile(),
		listeners: make(map[string]chan T),
	}
}

// SimpleFromContext creates a Simple [Conductor] from a given [context.Context].
func SimpleFromContext[T any](parent context.Context) Conductor[T] {
	c := Simple[T]()
	c.(*simple[T]).ctx = parent

	return c
}
