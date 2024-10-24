# Conductor

A generalization of the [Context][ctx]

## Motivation

The standard library [Context][ctx] is a great concept. It allows to propagate values
and allows to coordinate across API and goroutine boundaries. As much as I like it, when
dealing with application that show a lot of moving pieces, I want something more in
order to coordinate them. I want a [Conductor][cdct]. In the same spirit as the `Done`
method of a [Context][ctx], a [Conductor][cdct] `Cmd` method may be used to listen for
commands. These are represented by the ubiquitous generic parameter `T` across this
library, and may be everything, from simple `string`s to interfaces carrying a
complicated logic with them.

## How to use

Please, read also [the performance considerations](#performance).

There are two different [Conductor][cdct]s implemented right now, a [Simple][simple] and
a [Tagged][tagged] one.

The first is straightforward to use (we take `T` to be `string`, for the sake of
simplicity):

```go
simple := Simple[string]()

// we listen for commands

for {
	select {
	case cmd := <-simple.Cmd():
			// react to the commands
	case <-simple.Done():
			// a Conductor is also a Context, so we can use it the same way
			// for example, for controlling a clean exit
	}
}

// ...in another part of the control flow we can call Send to send a commands
// to all the listeners created with Cmd

Send[string](simple)("doit")

// This will fire in the above select statement, delivering "doit" to the first case
```

The second one might sound a bit more involved, but it's hopefully just a matter of
getting used to the syntax:

```go
tagged := Tagged[string]()

// we listen on many different possible tags

for {
	select {
	case cmd := <-WithTag[string](tagged, "tag1").Cmd():
		// React to a command in the "tag1" branch
	case cmd := <-WithTag[string](tagged, "tag2").Cmd():
		// React to a command in the "tag2" branch
	case <-tagged.Done():
		// As for the Simple, also the Tagged is a Context
	}
}


// We may selectively send a command, again using the Send function

Send[string](tagged, "tag1")("doitnow")

// We may also send a broadcast command

Send[string](tagged)("allhands")
```

### Performance

In the examples above and in those in the [examples/](./examples) folder, you can notice
that the typical usage of a conductor boils down to receiving from a channed produced by
the `Cmd` function in a case statement inside a `select`. This is done to echo the
syntax of [context.Context.Done][done]. To achieve this, the code tries to keep track
where it is called from, unwinding the stack using [runtime.Callers][callers] until we
exit this module and reach the calling site. This, coupled with the semantics of `case`
that if provided with a function calls it every time another statement fires, may
require the library to walk the stack very frequently. The following excerpt is taken
from the [performance][performance] example.

```go
func run(c conductor.Conductor[int], inst int) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Cmd():
			fmt.Println(inst, "received")
		case <-ticker.C:
			fmt.Println(inst, "tick")
		}
	}
}

func main() {
	c := conductor.Simple[int]()

	for i := 0; i < 5; i++ {
		go run(c, i)
	}

	time.Sleep(10 * time.Second)
	conductor.Send[int](c)(0)
	time.Sleep(500 * time.Millisecond)
}
```
Fortunately, this is easily avoidable, explicitly assigning the result of `Cmd` before
using it in a `case`:

```go
	lis := c.Cmd()

	for {
		select {
		case <-lis:
			fmt.Println(inst, "received")
		case <-ticker.C:
			fmt.Println(inst, "tick")
		}
	}
```

This is true both for `Simple` and `Tagged` conductors.

[ctx]: https://pkg.go.dev/context#Context
[done]: https://pkg.go.dev/context#Context.Done
[callers]: https://pkg.go.dev/runtime#Callers
[cdct]: ./conductor.go
[simple]: ./simple.go
[tagged]: ./tagged.go
[performance]: ./examples/performance/main.go


<!-- vim:set ft=markdown tw=88: -->
