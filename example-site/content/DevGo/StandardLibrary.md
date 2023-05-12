# Go Standard Library
* pkg.go.dev [Standard library](https://pkg.go.dev/std)

# Stringer Interface
* Implement the [Stringer](https://pkg.go.dev/fmt#Stringer) interface to print
  a type using a custom string.
<pre><code>type Person struct {
	Name string
	Age  int
}

<b>func (p Person) String() string</b> {
	return fmt.Sprintf("%v (%v years)", p.Name, p.Age)
}

func main() {
	a := Person{"Arthur Dent", 42}
	z := Person{"Zaphod Beeblebrox", 9001}
	fmt.Println(a)
	fmt.Println(z)
}

// Prints:
// Arthur Dent (42 years)
// Zaphod Beeblebrox (9001 years)
</code></pre>

# fmt Package
* pkg.go.dev [fmt](https://pkg.go.dev/fmt): "implements formatted I/O with
  functions analogous to C's printf and scanf."
* The name of the __print functions__ use these abbreviations:
  * `F`: Print to a specified file (e.g. `os.Stderr`).
  * `ln`: Add a trailing line feed.
  * `f`: Format the string. This is never combined with `ln`.
* The print functions are: `Fprint`, `Fprintf`, `Fprintln`, `Print`, `Printf`, and `Println`.
* There are corresponding methods to return a formatted string: `Sprint`, `Sprintf`, and `Sprintln`.
* Verbs:
| Verb | Output |
|------|--------|
| %v   | The value in a default format |
| %+v  | Prints field names for structs |
| %#v  | Go-syntax representation of the value |
| %T   | Go-syntax representation of the type of the value |

# log Package
* pkg.go.dev [log](https://pkg.go.dev/log): "[I]mplements a simple logging
  package. It defines a type, Logger, with methods for formatting output. It
  also has a __predefined 'standard' Logger...[t]hat logger writes to standard
  error__ and prints the date and time of each logged message...The Fatal
  functions call os.Exit(1) after writing the log message.  The Panic functions
  call panic after writing the log message."
* Log an error:
<pre><code>if err != nil {
    log.Printf("%v", err)
}
</code></pre>

# strconv Package
* [strconv](https://pkg.go.dev/strconv): "implements conversions to and from
  string representations of basic data types"

# text/tabwriter Package
* [text/tabwriter](https://pkg.go.dev/text/tabwriter): "implements a write
  filter (tabwriter.Writer) that translates tabbed columns in input into
  properly aligned text. The package is using the Elastic Tabstops algorithm
  described at <http://nickgravgaard.com/elastictabstops/index.html>"

# sort Package
* [sort](https://pkg.go.dev/sort): "provides primitives for sorting slices and
  user-defined collections"
* pkg.go.dev/sort [type Interface](https://pkg.go.dev/sort#Interface)
* pkg.go.dev/sort [func Sort](https://pkg.go.dev/sort#Sort)
<pre><code>type Item struct {
    Text     string
    Priority int
    position int
}

// ByPri implements sort.Interface for []Item based on Priority and position.
type ByPri []Item
func (s ByPri) <b>Len()</b> int { return len(s) }
func (s ByPri) <b>Swap(i, j int)</b> { s[i], s[j] = s[j], s[i] }
func (s ByPri) <b>Less(i, j int)</b> bool {
    if s[i].Priority == s[j].Priority {
        return s[i].position < s[j].position
    }
    return s[i].Priority < s[j].Priority
}

// Sort items.
items []Item = ...
<b>sort.Sort(ByPri(items))</b>
</code></pre>

# flag Package
* [flag](https://pkg.go.dev/flag): "implements command-line flag parsing."

# path/filepath Package
* [path/filepath](https://pkg.go.dev/path/filepath): "utility routines for manipulating filename paths"

# strings Package
* [strings](https://pkg.go.dev/strings): simple functions to manipulate UTF-8 encoded strings.
* Use [strings.Builder](https://pkg.go.dev/strings#Builder) to build strings using Write methods.
<pre><code>var <b>builder</b> strings.Builder
for i := 3; i >= 1; i-- {
	fmt.Fprintf(<b>&builder</b>, "%d...", i)
}
<b>builder.WriteString</b>("ignition")
fmt.Println(<b>builder.String()</b>)
</code></pre>

# exec Package
* [exec](https://pkg.go.dev/os/exec): "runs external commands".
* __Run a command and print stdout__, by calling
  [Command()](https://pkg.go.dev/os/exec#Command) to create
  a [Cmd](https://pkg.go.dev/os/exec#Cmd) and then calling
  [Output()](https://pkg.go.dev/os/exec#Cmd.Output) on it to run the command.
<pre><code>import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := <b>exec.Command("ls", "-l")</b>
	output, err := <b>cmd.Output()</b>
	if err != nil {
		fmt.Println("Error executing command:", err)
	}
	fmt.Println("Stdout output:")
	fmt.Println(string(output))
}
</code></pre>
* Run a command and print __both stderr and stdout__, by calling
  [Run()](https://pkg.go.dev/os/exec#Cmd.Run) instead of `Output()`, and attaching
  buffers to [Stdout and Stderr](https://pkg.go.dev/os/exec#Cmd.Stdout).
<pre><code>import bytes
...
cmd := exec.Command("ls", "-l", "/foo/bar")
var stdout, stderr bytes.Buffer
<b>cmd.Stdout = &stdout</b>
<b>cmd.Stderr = &stderr</b>
err := <b>cmd.Run()</b>
if err != nil {
	fmt.Println("Error executing command:", err)
}

fmt.Println("Stdout output:")
fmt.Println(string(stdout.Bytes()))

fmt.Println("Stderr output:")
fmt.Println(string(stderr.Bytes()))
</pre></code>

# os Package
* [os](https://pkg.go.dev/os): "provides a platform-independent interface to operating system functionality"
* Check for file existence with [os.Stat()](https://pkg.go.dev/os#Stat) and
  [os.IsNotExist()](https://pkg.go.dev/os#IsNotExist).
<pre><code>func pathDoesNotExist(path string) bool {
    if _, err := <b>os.Stat(path)</b>; err != nil {
        if <b>os.IsNotExist(err)</b> {
            return true
        }
    }
    return false
}
</code></pre>
* In the previous example the error returned is a pointer to
  a [fs.PathError](https://pkg.go.dev/io/fs#PathError) which in turn contains
  an `Err error` that is a [syscall.Errno](https://pkg.go.dev/syscall#Errno).
  To see details:
<pre><code>dir := "/path/to/directory"
&UnderBar;, err := os.Stat(dir)
fmt.Printf("err: %#v\n", err)
if pathError, ok := <b>err.(&ast;fs.PathError)</b>; ok {
    fmt.Printf("pathError: %#v\n", pathError)
    fmt.Printf("pathError.Err type: %T\n", pathError.Err)
    fmt.Printf("pathError.Err: %#v\n", pathError.Err)
    if errNo, ok := <b>pathError.Err.(syscall.Errno)</b>; ok {
       fmt.Printf("errNo: %#v\n", errNo)
    }
}

/&ast; Output generated:
err: &fs.PathError{Op:"stat", Path:"/path/to/directory", Err:0x2}
pathError: &fs.PathError{Op:"stat", Path:"/path/to/directory", Err:0x2}
pathError.Err type: syscall.Errno
pathError.Err: 0x2
errNo: 0x2
&ast;/
</code></pre>

