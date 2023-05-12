# Concurrency
* Effective Go [Concurrency](https://go.dev/doc/effective_go#concurrency)
* go.dev/ref/spec [Go statements](https://go.dev/ref/spec#Go_statements)
* go.dev/ref/spec [Select statements](https://go.dev/ref/spec#Select_statements)
* go.dev/tour [Concurrency](https://go.dev/tour/concurrency/1)

# Goroutines
* A __goroutine__ is a lightweight thread managed by the Go runtime.
* Any function can be launched as a goroutine using the __`go` keyword__.
  Parameters can be passed in, but any __return values are ignored__.
<pre><code><b>func say(s string)</b> {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func main() {
	<b>go say("world")</b>
	say("hello")
}
</code></pre>

# Channels
* Goroutines communicate using __channels__.
* __Create a channel__ using the built-in [func make](https://pkg.go.dev/builtin#make).
<pre><code>ch := make(chan int)
</code></pre>
* __Channels are reference types__, similart to slices. Passing a channel to
  a function passes a pointer to the channel.
* __Read from a channel__ using the `<-` operator, with the channel on the
  right-hand side of the operator.
<pre><code>a := <-ch
</code></pre>
* __Write to a channel__ using the `<-` operator, with the channel on the
  left-hand side of the operator.
<pre><code>ch <- b
</code></pre>
* By default channels are __unbuffered__. Writing to an unbuffered channel will block
  until the previously written value has been read. Reading from an unbuffered channel 
  will block until the previously written value has been read.
* Channels can be __buffered__. The buffer capacity is specified when creating the channel.
<pre><code>ch := make(chan int, 10)
</code></pre>
* The built-in [func cap](https://pkg.go.dev/builtin#cap) returns the capacity of the channel.
* The built-in [func len](https://pkg.go.dev/builtin#len) returns how many value are currently in the channel.
* A `for range` loop can be used to read from a channel.
<pre><code>for v := range ch {
    fmt.Println(v)
}
</code></pre>
* Call [close](https://pkg.go.dev/builtin#close) on a channel to communicate to
  receivers that there are no more values to read.
<pre><code>func main() {
    jobs := make(chan int, 5)
    done := make(chan bool)

    // Process jobs. "more" will be set to false when the "jobs" channel has
    // been closed.
    go func() {
        for {
            j, <b>more</b> := <-jobs
            if <b>more</b> {
                fmt.Println("received job", j)
            } else {
                fmt.Println("received all jobs")
                done <- true
                return
            }
        }
    }()

    // Send jobs.
    for j := 1; j <= 3; j++ {
        jobs <- j
        fmt.Println("sent job", j)
    }
    <b>close(jobs)</b>
    fmt.Println("sent all jobs")

    // Wait for jobs to be processed.
    <-done
}
</code></pre>
* Calling `close()` on a channel that's already been closed causes a panic, as
  will writing to a closed channel.
* Reading from a closed channel always works. Unread values are returned if the
  channel still has values to read.  Otherwise, the zero value is returned. Use
  the __comma ok idiom__ to determine if there are still values to read.
<pre><code>v, ok := -< ch
</code></pre>

# The `select` Statement
* The `select` statement lets a goroutine __wait on multiple communication
  operations__. It __blocks until one of its cases can run__, then it executes
  that case. It __chooses one at random if multiple are ready__.  The optional
  __`default` case__ is run if no other case is ready. 
<pre><code>select {
case v := <-ch:
    fmt.Println(v)
case v := <-ch2:
    fmt.Println(v)
case ch3 <- x:
    fmt.Println("wrote", x)
case <-ch4:
    fmt.Println("got value on ch4, but ignored it")
}
</code></pre>
* The `default` case can be used to implement and __non-blocking read or write__.
<pre><code>messages := make(chan string)

select {
case msg := <-messages:
    fmt.Println("received message", msg)
default:
    fmt.Println("no message received")
}
</code></pre>
* `select` is often embedded in a `for` statement. This is often called
  a "for-select" statement.  Here we use a separate channel called done to
  signal when to stop looping, and then `return` to exit the loop.
<pre><code>for {
    select {
    case <-done:
        return
    case v := <-ch:
        fmt.Println(v)
    }
}
</code></pre>
* Having a `default` case inside a for-select loop is almost always the wrong
  thing to do since the default case will run through the loop when there's
  nothing to do, and consume lots of CPU. (LGO p 336)
* __Turn off a case in a select that reads__ by setting the channel to `nil`
  when it has nothing left to do. Otherwise, reads will continue to be done,
  with `ok` always set to false. Reading from `nil` channel blocks, which
  effectively turns off the case.
<pre><code>for {
    select {
    case v, ok := <-in:
        if !ok {
            <b>in = nil</b> // the case will never succeed again!
            continue
        }
        // process the v that was read from in
     case v, ok := <-in2:
        if !ok {
            <b>in2 = nil</b> // the case will never succeed again!
            continue
        }
        // process the v that was read from in2
     case <-done:
        return
}
</code></pre>

# Channel Syncronization
* Use a "done channel" to wait for a goroutine to finish.
<pre><code>func worker(<b>done chan bool</b>) {
    fmt.Print("working...")
    time.Sleep(time.Second)
    fmt.Println("done")

    <b>done <- true</b>
}

func main() {
    <b>done := make(chan bool, 1)</b>
    go worker(<b>done</b>)
    <b><-done</b>
}
</code></pre>
* Or, to wait on more than one goroutine to finish use a [WaitGroup](https://pkg.go.dev/sync#WaitGroup).
<pre><code>var wg sync.WaitGroup
<b>wg.Add(3)</b>
go func() {
    defer <b>wg.Done()</b>
	doThing1()
}()
go func() {
    defer <b>wg.Done()</b>
	doThing2()
}()
go func() {
    defer <b>wg.Done()</b>
    doThing3()
}()
<b>wg.Wait()</b>
</code></pre>

# Timeouts
* Use [time.After()](https://pkg.go.dev/time#After) to implement a timeout.
<pre><code>func main() {
    // Channel c1 will return a result after 2s. Note that the channel is
    // buffered, so the send in the goroutine is nonblocking. This is a common
    // pattern to prevent goroutine leaks in case the channel is never read.
	c1 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		c1 <- "result 1"
	}()

    // Here we wait on c1 but timeout after 1s.
	select {
	case res := <-c1:
		fmt.Println(res)
	<b>case <-time.After(1 * time.Second)</b>:
		fmt.Println("timeout 1")
	}

	// Let's do this again with a new channel, but not timeout by waiting longer.
	c2 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		c2 <- "result 2"
	}()
	select {
	case res := <-c2:
		fmt.Println(res)
	<b>case <-time.After(3 * time.Second)</b>:
		fmt.Println("timeout 2")
	}
}
</code></pre>

# Mutexes
* Use [Mutex](https://pkg.go.dev/sync#Mutex) to lock and unlock shared memory.
<pre><code>// Define a struct that has shared data and control access with a Mutex.  Note
// that Mutexes must not be copied, so if this `struct` is passed around, it
// should be done by pointer.
type Container struct {
	<b>mu       sync.Mutex</b>
	counters map[string]int
}

func (c &ast;Container) inc(name string) {
    // Lock the mutex before accessing `counters`; unlock it at the end of the
    // function using a [defer](defer) statement.
	<b>c.mu.Lock()</b>
	<b>defer c.mu.Unlock()</b>
	c.counters[name]++
}

func main() {
	c := Container{
        // Note that the zero value of a mutex is usable as-is, so no
        // initialization is required here.
		counters: map[string]int{"a": 0, "b": 0},
	}

	var wg sync.WaitGroup

	// This function increments a named counter in a loop.
	doIncrement := func(name string, n int) {
		for i := 0; i < n; i++ {
			c.inc(name)
		}
		wg.Done()
	}

    // Run several goroutines concurrently; note that they all access the same
    // `Container`, and two of them access the same counter.
	wg.Add(3)
	go doIncrement("a", 10000)
	go doIncrement("a", 10000)
	go doIncrement("b", 10000)

	// Wait for the goroutines to finish
	wg.Wait()
	fmt.Println(c.counters)
}
</code></pre>
* Use [RWMutex](https://pkg.go.dev/sync#RWMutex) to allow must one writer but multiple readers.

# Best Practices
* In general __keep your APIs concurrency-free__. Don't expose channels or
  mutexes as parameters. (LGO p336)
* __Specify whether a `chan` parameter or return value will be used for input or
  output__ by using the `<-` operator as part of its type.
<pre><code>// The return value here is for reading from.
func countTo(max int) <-chan int { ...

// This function has a channel to read from and a second to write to.
func runThingsConcurrently(in <-chan int, out chan<- int) { ...
</code></pre>
* __Avoid "goroutine leaks"__ by making sure goroutines end when done. Otherwise
  the scheduler give them time to run.  For example, this goroutine never
  ends because not all values are read:
<pre><code>func countTo(max int) <-chan int {
	ch := make(chan int)
	go func() {
		for i := 0; i < max; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

func main() {
	for i := range countTo(10) {
		if i > 5 {
			break
		}
		fmt.Println(i)
	}
}
</code></pre>

# Timers
* __Create a `Timer`__ with [time.NewTimer()](https://pkg.go.dev/time#NewTimer).
  __Cancel it__ with [func (&ast;Timer) Stop](https://pkg.go.dev/time#Timer.Stop).
<pre><code>// Create a timer that will fire in 2 seconds.
timer1 := <b>time.NewTimer(2 * time.Second)</b>

// Wait on the timer.
<-timer1.C
fmt.Println("Timer 1 fired")

// Create another timer but cancel it before it fires.
timer2 := <b>time.NewTimer(time.Second)</b>
go func() {
    <-timer2.C
	fmt.Println("Timer 2 fired")
}()
stop2 := <b>timer2.Stop()</b>
if stop2 {
    fmt.Println("Timer 2 stopped")
}

// Give the `timer2` enough time to fire, if it ever was going to, to show it
// is in fact stopped.
<b>time.Sleep(2 * time.Second)</b>
</code></pre>
* Or just use __[time.Sleep()](https://pkg.go.dev/time#Sleep) if the timer
  doesn't need to be canceled__.
* Timers fire just once. __To repeatedly fire use a `Ticker`__. __Create a `Ticker`__
  with [time.NewTicker()](https://pkg.go.dev/time#NewTicker). __Stop a `Ticker`__
  with [func (t &ast;Ticker) Stop](https://pkg.go.dev/time#Ticker.Stop).
<pre><code>// Create a Ticker that will fire every 500ms.
ticker := <b>time.NewTicker(500 * time.Millisecond)</b>
done := make(chan bool)

// Read timestamps from the Ticker as it fires.
go func() {
    for {
        select {
        case <-done:
            return
        case t := <-ticker.C:
            fmt.Println("Tick at", t)
        }
    }
}()

// Stop the Ticker.
time.Sleep(1600 * time.Millisecond)
<b>ticker.Stop()</b>
done <- true
fmt.Println("Ticker stopped")
</code></pre>

# Atomic Counters
* Call the sync module's
  [atomic.AddUint64()](https://pkg.go.dev/sync/atomic#AddUint64) to do an
  __atomic add__ on an `uint64`.
* Call [atomic.LoadUint64()](https://pkg.go.dev/sync/atomic#LoadUint64) to read
  the `uint64`.

# Context
* pkg.go.dev [context](https://pkg.go.dev/context)
* go.dev/blog [Go Concurrency Patterns: Context](https://go.dev/blog/context)
* go.dev/blog [Contexts and structs](https://go.dev/blog/context-and-structs)
* __Prevent a goroutine leak by using a cancelable context__, created with the
  function [context.WithCancel()](https://pkg.go.dev/context#WithCancel). In
  the example below only some of the generated value are read, but the
  goroutine doesn't leak because it's canceled.
<pre><code>// gen generates integers in a separate goroutine and
// sends them to the returned channel.
// The callers of gen need to cancel the context once
// they are done consuming generated integers not to leak
// the internal goroutine started by gen.
<b>gen := func(ctx context.Context)</b> <-chan int {
	dst := make(chan int)
	n := 1
	go func() {
		for {
			select {
			<b>case <-ctx.Done():</b>
				return // returning not to leak the goroutine
			case dst <- n:
				n++
			}
		}
	}()
	return dst
}

<b>ctx, cancel := context.WithCancel(context.Background())</b>
<b>defer cancel()</b> // cancel when we are finished consuming integers

for n := range <b>gen(ctx)</b> {
	fmt.Println(n)
	if n == 5 {
		break
	}
}
</code></pre>
* __Abandon a blocking function after a deadline has passed__ by creating
  a Context with the function function
  [context.WithDeadline()](https://pkg.go.dev/context#WithDeadline).
<pre><code>const shortDuration = 1 * time.Millisecond
d := time.Now().Add(shortDuration)
<b>ctx, cancel := context.WithDeadline(context.Background(), d)</b>

// Even though ctx will be expired, it is good practice to call its
// cancellation function in any case. Failure to do so may keep the context and
// its parent alive longer than necessary.
defer cancel()

select {
case <-time.After(1 * time.Second):
	fmt.Println("overslept")
<b>case &lt;-ctx.Done()</b>:
	fmt.Println(ctx.Err())
}

// Prints:
// context deadline exceeded
</code></pre>
* Or, a variation on this is to use
  [context.WithTimeout()](https://pkg.go.dev/context#WithTimeout) instead of
  `context.WithDeadline()`, to __abandon a blocking function after a timeout
  has expired__.
* __Pass a value to a Context and retrieve it__ by creating a context using
  [context.WithValue()](https://pkg.go.dev/context#WithValue) and doing the lookup
  by calling `Context.Value()`.
<pre><code>type favContextKey string

f := func(ctx context.Context, k favContextKey) {
	if v := <b>ctx.Value(k)</b>; v != nil {
		fmt.Println("found value:", v)
		return
	}
	fmt.Println("key not found:", k)
}

k := favContextKey("language")
<b>ctx := context.WithValue(context.Background(), k, "Go")</b>

f(ctx, k)
f(ctx, favContextKey("color"))

// Prints:
// found value: Go
// key not found: color
</code></pre>

