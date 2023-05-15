# Go Language; 2023-02
* go.dev [The Go Programming Language Specification](https://go.dev/ref/spec)

# Primitive Types
* The "[zero value](https://go.dev/ref/spec#The_zero_value)" is the default value
  given to a variable when it's not given an initial value.

# Primitive Types: Strings
* The zero value for a string is the empty string.
* Strings are immutable.
* [String literals](https://go.dev/ref/spec#String_literals): Use __backquote__
  to define a __raw string literal__, and __double quote__ to define an
  __interpreted string literal__.

# Primitive Types: Numerics
* go.dev/ref/spec [Numeric types](https://go.dev/ref/spec#Numeric_types)
* `byte` is an alias for `uint8`, and makes code clearer. 
* Similarly, `rune` is alias for `int32`, and makes code clearer.
* `int` is `int32` on 32-bit CPUs and `int64` on 64-bit CPUs. Use `int` and let
  the compiler optimize, unless you need to be explicit. Ditto for `uint`,
  `uint32`, and `uint64`.
* Use `float64` instead of `float32` for the extra decimal place precision it gives.

# Variable Declarations
* Most verbose:
<pre><code>var x int = 10
</code></pre>
* The type can be left off if the right-hand side defines it:
<pre><code>var x = 10
</code></pre>
* Multiple variables can be defined on one line:
<pre><code>var x, y int = 10, 20
var a, b     = 10, "hello"
</code></pre>
* A declaration list can be used to define multiple variables at once:
<pre><code>var (
    x    int
    y        = 20
    z    int = 30
    d, e     = 40, "hello"
    f,g  string
)
</code></pre>
* The __short declaration format__ using `:=` can be used inside functions. It
  can be used to assign values to existing variables too:
<pre><code>var x = 10
x := 10
</code></pre>
* Be careful with `:=`, though. It will cause a variable to be shadowed if that
  variable is defined in an outer block, and won't use the existing variable. (LGO p107)
* The __shadow linter__ can be used to detect unexpected shadowing. (LGO p110)

# const
* go.dev/ref/spec [Constant declarations](https://go.dev/ref/spec#Constant_declarations)
* Consts in Go just give names to literals. A variable itself cannot be declared const.
* Consts do not need to be typed. Types only come into effect when assigning values.
<pre><code>const x = 10
var y int = x
var z float64 = x
var d tye = x
</code></pre>

# Arrays
* Effective Go [Arrays](https://go.dev/doc/effective_go#arrays)
* go.dev/ref/spec [Array types](https://go.dev/ref/spec#Array_types)
* Are rarely used directly, unless you know the exact size you need ahead of time.
  They're mainly used as the backing store for slices.
* Defining an array:
<pre><code>// With zero values
var x [3]int

// With initial values, an "array literal".
var x = [3]int{10, 20, 30}

// The array size can be left off with an array literal, and an ellipses used instead.
var x = [...]int{10, 20, 30}

// A "sparse array", with some initial values. Unset values are set to the zero value.
// The array that's created has:
//   [1, 0, 0, 0, 0, 4, 6, 0, 0, 0, 100, 15]
var x = [12]{1, 5: 4, 6, 10: 100, 15}
</code></pre>
* The `==` and `!=` operators can be used to compare arrays.
<pre><code>var x = [...]int{1, 2, 3}
var y = [3]int{1, 2, 3}
fmt.Println(x == y) // prints true
</code></pre>

# Slices
* Effective Go [Slices](https://go.dev/doc/effective_go#slices)
* go.dev/ref/spec [Slice types](https://go.dev/ref/spec#Slice_types)
* go.dev/blog [Go Slices: usage and internals](https://go.dev/blog/slices-intro) (2011)
* Slices are declared similarly to arrays, but no size is specified:
<pre><code>var x = []int{10, 20, 30}
</code></pre>
* The `==` and `!=` operators can't be used to compare slices, although can be used
  to check whether a slice is `nil`.
<pre><code>var x []int
fmt.Println(x == nil) // prints true
</code></pre>
* Use the built-in [func len](https://pkg.go.dev/builtin#len) to __determine the size of a slice__.
<pre><code>var x = []int{10, 20, 30}
fmt.Println(<b>len(x)</b>) // prints 3
</code></pre>
* Use the built-in [func append](https://pkg.go.dev/builtin#append) to __append to a slice__.
<pre><code>var x = []int{10, 20, 30}
x = append(x, 40, 50, 60)
</code></pre>
* Use the built-in [func cap](https://pkg.go.dev/builtin#cap) to see the __capacity__ of a slice, or 
  how much space is set aside for new elements.
<pre><code>var x = []int{10, 20, 30, 40}
fmt.Println(len(x), cap(x)) // prints "4 4"
x = append(x, 50)
fmt.Println(len(x), cap(x)) // prints "5 8"
</code></pre>
* The built-in [func make](https://pkg.go.dev/builtin#make) can be used to set
  the __initial size of a slice__. Values are set to the zero value.
<pre><code>var x = <b>make([]int, 5)</b>
fmt.Println(x) // prints "[0 0 0 0 0]"
</code></pre>
* The __initial capacity__ can be set as well.
<pre><code>var x = <b>make([]int, 5, 10)</b>
</code></pre>
* Use a __slice expression__ to create a slice from a slice. Starting and
  ending offsets are given, separated by a colon. The element at the ending index
  is not part of the returned slice. The default starting index is
  0 and the default ending is one past the last index, when not given. 
<pre><code>var x = []int{1, 2, 3, 4}
fmt.Println(x[:2])  // prints "[1 2]"
fmt.Println(x[1:])  // prints "[2 3 4]"
fmt.Println(x[1:3]) // prints "[2 3]"
fmt.Println(x[:])   // prints "[1 2 3 4]"
</code></pre>
* Slices of slices share memory. A change to one slice affects other slices. This
  can result in unexpected side-effects, especially when appending. (LGO p80)
* Use the built-in [func copy](https://pkg.go.dev/builtin#copy) to __make a copy of a slice__.
  The first parameter is the dest and second the source. The number or elements copied is
  returned. The number copied is the smaller of the size the source and dest. This is
  based on the length of the slices, and not their capacity.
<pre><code>// Basic copy
x := []int{1, 2, 3, 4}
y := make([]int, 4)
num := copy(y, x)
fmt.Println(y, num) // prints "[1 2 3 4] 4"

// Copy over values in the same array. Here the last 3 values are copied over
// the first 3 values.
x := []int{1, 2, 3, 4}
num := copy(x[:3], x[1:])
fmt.Println(x, num) // prints "[2 3 4 4] 3"
</code></pre>
* Use [reflect.DeepEqual](https://pkg.go.dev/reflect#DeepEqual) to compare slices.

# Strings
* go.dev/ref/spec [String types](https://go.dev/ref/spec#String_types)
* Strings are sequences of `byte`s.
* The `byte`s don't need to have a particular encoding, but many functions
  assume the encoding is UTF-8 code points.
* The `len()` function returns number of `byte`s, and not characters.

# Maps
* go.dev/ref/spec [Map types](https://go.dev/ref/spec#Map_types)
* Effective Go [Maps](https://go.dev/doc/effective_go#maps)
* Creating maps:
<pre><code>// The <b>zero value</b> for a map is <b>nil</b>. Here the map is of `int`s
// with a `string` index.
<b>var nilMap map[string]int</b>
fmt.Println(nilMap == nil) // prints true

// Create an <b>empty map</b> using a <b>map literal</b>.
<b>totalWins := map[string]int{}</b>
fmt.Println(totalWins == nil) // prints false

// Create a <b>non-empty map</b> using a <b>map literal</b>. Here the map is of
// slices of strings with a string index.
<b>teams := map[string][]string{
    "Orcas":   []string{"Fred", "Ralph", "Bijou"},
    "Lions":   []string{"Sarah", "Peter", "Billie"},
    "Kittens": []string{"Waldo", "Raul", "Ze"},
}</b>

// Create a map with a <b>predefined size</b>. Note len() here will return 1, and
// <b>cap() is not valid for a map</b>.
<b>ages := make(map[int][]string, 10)</b>
ages[40] = []string{"Bob", "Nancy", "Carol"}
fmt.Println(ages, len(ages))) // prints "map[40:[Bob Nancy Carol]] 1"
</code></pre>
* A map will return the zero value for entries not found. Use the __comma ok idiom__
  when you need to know if a value was not found. (LGO p 94)
<pre><code>m := map[string]int{
		"hello": 5,
		"world": 0,
}
<b>v, ok</b> := m["hello"]
fmt.Println(v, ok) // prints "5 true"

<b>v, ok</b> = m["goodbye"]
fmt.Println(v, ok) // prints "0 false"
</code></pre>
* Use the built-in [func delete](https://pkg.go.dev/builtin#delete) to <b>delete a map element</b>.
  If the key isn't in the map or the map is `nil` nothing happens.
<pre><code>m := map[string]int{
    "hello": 5,
    "world": 10,
}
delete(m, "hello")
</code></pre>

# Blocks
* go.dev/ref/spec [Blocks](https://go.dev/ref/spec#Blocks)
* Blocks influence scope.
* __Explicit blocks__ are defined by braces.
* There are also __implicit blocks__:
  * __Universe block__: is all code.
  * __Package block__: is the package.
  * __File block__: is the file.
  * `if`, `for`, and `switch` statements each define a block.
  * Each clause in a `switch` and `select` statement define a block.

# If Statements
* go.dev/ref/spec [If statements](https://go.dev/ref/spec#If_statements)
* Effective Go [If](https://go.dev/doc/effective_go#if)
* An `if` statement can define variables that exist for the the condition and
  all if and else blocks. The variable goes out of scope after the `if`
  statement; the last `Println()` below would cause a syntax error since `n` is
  no longer defined.
<pre><code>if n := rand.Intn(10); n == 0 {
    fmt.Println("That's too low")
} else if n > 5 {
    fmt.Println("That's too big:", n)
} else {
    fmt.Println("That's a good number:", n)
}
//fmt.Println(n) // Would cause a syntax error.
</code></pre>

# For Statements
* go.dev/ref/spec [For statements](https://go.dev/ref/spec#For_statements)
* Effective Go [For](https://go.dev/doc/effective_go#for)
* `for` is the only looping construct in Go.
* There are 4 types of `for` statements: 
  * The __complete C-style__ `for`.
  * A __condition-only__ `for`.
  * An __infinite__ `for`.
  * A `for range`.
* Like with `if`, a `for` statement can define variables that exist for the
  `for` statement.
* Examples:
<pre><code>// The <b>complete for</b> statement.
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// The <b>condition-only for</b> (like a while loop).
i := 1
for i < 100 {
    fmt.Println(i)
    i *= 2
}

// The <b>infinite for</b>
for {
    fmt.Println("Hello")
}

// The <b>for range</b>
evenVals := []int{2, 4, 6, 8, 10, 12}
for i, v := range evenVals {
    fmt.Println(i, v)
}
</code></pre>

# Switch Statements
* go.dev/ref/spec [Switch statements](https://go.dev/ref/spec#Switch_statements)
* Effective Go [Switch](https://go.dev/doc/effective_go#switch)
* Example:
<pre><code>words := []string{"a", "cow", "smile", "gopher", "octopus", "anthropologist"}
for &UnderBar;, word := range words {
    switch size := len(word); size {
	case 1, 2, 3, 4:
        fmt.Println(word, "is a short word!")
    case 5:
	    wordLen := len(word)
		fmt.Println(word, "is exactly the right length:", wordLen)
    case 6, 7, 8, 9:
    default:
        fmt.Println(word, "is a long word!")
    }
}
</code></pre>
* Switches can also be "__blank__", and not specify a value to compare on.
  Instead, __each `case` statement uses a boolean comparison__.
<pre><code>words := []string{"hi", "salutations", "hello"}
for &UnderBar;, word := range words {
	switch wordLen := len(word); {
	case wordLen < 5:
		fmt.Println(word, "is a short word!")
	case wordLen > 10:
		fmt.Println(word, "is a long word!")
	default:
		fmt.Println(word, "is exactly the right length.")
	}
}
</code></pre>

# Functions
* go.dev/ref/spec [Function declarations](https://go.dev/ref/spec#Function_declarations)
* Effective Go [Functions](https://go.dev/doc/effective_go#functions)
* Variable numbers of input parameters are possible ("variadic parameters") by
  using `...` before the type. This creates a slice of that type:
<pre><code>func addTo(base int, <b>vals ...int</b>) []int {
	out := make([]int, 0, len(vals))
	for _, v := range vals {
		out = append(out, base+v)
	}
	return out
}
</code></pre>
* Use `defer` to run clean-up code after a function returns.
<pre><code>// Contents returns the file's contents as a string.
func Contents(filename string) (string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    <b>defer f.Close()</b>  // f.Close will run when we're finished.
    ...
</code></pre>
* A function can have __multiple defers__ and they will be run in a last-in-first-out order.
* The function used by `defer` can be a __closure__, defined inline. The closure can refer
  to return value by using __named result parameters__.
<pre><code>func DoSomeInserts(ctx context.Context, db &ast;sql.DB, value1, value2 string) <b>(err error)</b> {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	<b>defer func() {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			tx.Rollback()
		}
	}()</b>
	_, err = tx.ExecContext(ctx, "INSERT INTO FOO (val) values $1", value1)
	if err != nil {
		return err
	}
	// use tx to do more database inserts here
	return nil
}
</code></pre>
* Go is __pass by value__, and values are copied to parameters for function
  calls; e.g. passing a `struct` instance will copy the values of the `struct`
  into a new instance of the `struct`.
* Go functions may be __closures__. A closure is a function value that references
  variables from outside its body. The function may access and assign to the
  referenced variables; in this sense the function is "bound" to the variables. 
  Here the function returned by `adder()` is a closure, and each instance of
  the closure has its own copy of `sum`.
<pre><code>func <b>adder()</b> func(int) int {
	<b>sum := 0</b>
	<b>return func(x int)</b> int {
		sum += x
		return sum
	}
}

func main() {
	pos, neg := <b>adder()</b>, <b>adder()</b>
	for i := 0; i < 5; i++ {
		fmt.Println(
			pos(i),
			neg(-2*i),
		)
	}
}

// Prints:
// 0 0
// 1 -2
// 3 -6
// 6 -12
// 10 -20
</code></pre>

# Pointers, new, and make
* go.dev/ref/spec [Pointer types](https://go.dev/ref/spec#Pointer_types)
* Effective Go [Pointers vs. Values](https://go.dev/doc/effective_go#pointers_vs_values)
* pkg.go.dev [new()](https://pkg.go.dev/builtin#new): Allocates memory for
  a specified type, zeros its memory, and returns a pointer.
* pkg.go.dev [make()](https://pkg.go.dev/builtin#make): Allocates memory and
  initializes a `slice`, `map`, or `chan` (only). Unlike `new()`, it returns
  the type and not a pointer to the type.
* The zero value for a pointer is `nil`.
* __Instances of `struct` are values versus references__, unlike in Python,
  Java, etc. Instances go on the stack whenever possible.
* Note, though, that `map`s are pointers, and passed to functions as pointers.
  The caller will see any changes made.
* __Instances of `slice` however, are structs passed by value and changes to
  the slice will not usually be seen. The exception to this is if the contents
  of the slice is changed without affecting the size. The contents array the
  slice points to changes, and this change is seen by callers.__ (LGO p192-7)
* Go implicitly dereferences a pointer used to access `struct` fields.
<pre><code>type Vertex struct {
	X int
	Y int
}

v := Vertex{1, 2}
p := &v

// Implicit dereference.
p.X = 1234

// Explicit dereference works too.
(&ast;p).X = 5678
}
</code></pre>
* A reference to local struct can be returned from a function. Go will place it
  on the heap, and the pointer remains valid. Normally it would go on the
  stack, for better performance since heap management takes longer.  This is
  done with what's called "__escape analysis__", to determine whether pointers
  can use memory on the stack versus heap, for better performance.  The same
  applies to data that might normally go on the heap. The Go compiler will put
  it on the stack if it can, using escape analysis. (LGO p 198-204)
<pre><code>func NewFile(fd int, name string) &ast;File {
    if fd < 0 {
        return nil
    }
    f := File{fd, name, nil, 0}
    <b>return &f</b> // Go sees this and puts f on the heap even though it's a local struct instance.
}
</code></pre>
* Both `new(MyStruct)` and `&MyStruct{}` do to the same thing.

# Structs
* go.dev/ref/spec [Struct types](https://go.dev/ref/spec#Struct_types)
* Define a struct and instantiate instances:
<pre><code>type person struct {
    name string
    age  int
    pet  string
}

// Create a zero value instance. Every field is set to the zero value.
// <b>Note that the a struct instance cannot be compared to nil</b>, and 
// that the expression "fred == nil" results in the compile time error 
// "invalid operation: fred == nil (mismatched types person and untyped nil)"
var fred person

// Printing fred just shows 0 for age since the two strings are empty and don't
// appear as anything.
fmt.Println(fred) // prints "{ 0 }"

// Create a zero value instance using a <b>struct literal</b>. This has the
// same effect as the previous declaration for fred.
bob := person{}

// Define a nonempty struct using a comma-separated list of values.
julia := person{
    "Julia",
    40,
    "cat",
}

// Define a nonempty struct using field names, similar to how a map is declared.
beth := person{
    age: 30,
    name: "Beth",
}
</code></pre>
* Structs can be compared if all their fields are comparable.

# Methods
* go.dev/ref/spec [Method declarations](https://go.dev/ref/spec#Method_declarations)
* Effective Go [Methods](https://go.dev/doc/effective_go#methods)
* Give a function a __receiver__ to make it a method. The receiver can be
  a __value receiver__ if the method doesn't modify the receiver, or
  a __pointer receiver__ if it modifies the receiver, or if it needs to handle
  `nil` instances.
<pre><code>type Counter struct {
	total       int
	lastUpdated time.Time
}

// <b>Pointer receiver</b> since the receiver is updated.
func <b>(c &ast;Counter)</b> Increment() {
	c.total++
	c.lastUpdated = time.Now()
}

// <b>Value receiver</b> since the receiver is not updated.
func <b>(c Counter)</b> String() string {
	return fmt.Sprintf("total: %d, last updated: %v", c.total, c.lastUpdated)
}

// Both methods are called the same way. (The call to the pointer receiver
// method does not need to dereference first.)
func main() {
	var c Counter
	fmt.Println(<b>c.String()</b>)
	<b>c.Increment()</b>
	fmt.Println(<b>c.String()</b>)
}
</code></pre>
* __The pointer passed to a pointer receiver method can be `nil`__. In the example
  below when this happens, the method creates an instance.
<pre><code>type IntTree struct {
	val         int
	left, right &ast;IntTree
}

func (it &ast;IntTree) Insert(val int) &ast;IntTree {
	<b>if it == nil {
		return &IntTree{val: val}
	}</b>
    ...
}
...
func main() {
	var it *IntTree
	it = it.Insert(5) // <b>"it" is null and so an instance of IntTree is created</b>.
    ...
}
</code></pre>
* Another reason to use a pointer receiver is to avoid copying the value on
  each method call. This can be more efficient if the receiver is a large struct.
* Name getters the same as the field they service, and don't use the term "Get"; e.g.
  `Owner()` would surface the field `owner`. 

# Construction
* Go does not have constructors.
* Instead, if a constructor is needed, for initialization, the convention is to
  create a factory method whose name begins with `New`, or just `New()` if the
  package only exports one type.
<pre><code>func NewJob(command string, logger &ast;log.Logger) &ast;Job {
    return &Job{command, logger}
}
</code></pre>

# Enumerations
* go.dev/ref/spec [Iota](https://go.dev/ref/spec#Iota)
* Create an enumeration by defining a type for the enumeration and then using
  the `iota` keyword in a `const` definition:
<pre><code>type MailCategory int

const (
    Uncategorized MailCategory = iota
    Personal
    Spam
    Social
    Advertisements
)
</code></pre>

# Embedding
* Effective Go [Embedding](https://go.dev/doc/effective_go#embedding)
* Go doesn't support inheritance, but does have embedding. All fields and
  methods on an embedded field can be called directly. An field is embedded by 
  just listing the type, and not giving the field a name.
<pre><code>type Employee struct {
	Name string
	ID   string
}

func (e Employee) Description() string {
	return fmt.Sprintf("%s (%s)", e.Name, e.ID)
}

type Manager struct {
	<b>Employee</b>
	Reports []Employee
}

func main() {
	m := Manager{
		Employee: Employee{
			Name: "Bob Bobson",
			ID:   "12345",
		},
		Reports: []Employee{},
	}
    fmt.Println(<b>m.Description()</b>) // Calls Description() on Employee.
}
</code></pre>

# Interfaces
* go.dev/ref/spec [Interface types](https://go.dev/ref/spec#Interface_types)
* Effective Go [Interfaces and other types](https://go.dev/doc/effective_go#interfaces_and_types)
* By convention interface names usually end in "er".
* Interfaces are implemented implicitly. The implementing type implements an interface by implementing
  all the methods on the interface, and doesn't explicitly refer to the interface.
* Just like you can embed a type in a struct, __an interface can be embedded in an interface__.
* One potential disadvantage with passing interfaces to and from methods is that it requires
  heap allocation, unlike passing structs with are passed by value. (LGO p 233)
* For an interface to be considered `nil` both the type and value must be `nil`. (LGO p 234)
<pre><code>var s &ast;string
fmt.Println(s == nil) // prints true
var i <b>interface{}</b>
fmt.Println(i == nil) // prints true
i = s
fmt.Println(i == nil) // prints false
</code></pre>
* `interface{}` denotes a type that can store anything. (LGO p 236)

# Type Conversions
* go.dev/ref/spec [Conversion](https://go.dev/ref/spec#Conversions)
* The expression `T(v)` converts the value `v` to the type `T`. 
<pre><code>var x, y int = 3, 4
var f float64 = math.Sqrt(float64(x&ast;x + y&ast;y))
var z uint = uint(f)
fmt.Println(x, y, z)
</code></pre>

# Type Assertions and Type Switches
* Effective Go [Type switch](https://go.dev/doc/effective_go#type_switch)
* Effective Go [Interface conversions and type assertions](https://go.dev/doc/effective_go#interface_conversions)
* go.dev/ref/spec [Type assertions](https://go.dev/ref/spec#Type_assertions)
* go.dev/ref/spec [Switch statements](https://go.dev/ref/spec#Switch_statements)
* A __type assertion__ casts an interface to a specific type. Note the interface must be of that type or a runtime panic occurs.
<pre><code>type MyInt int

func main() {
	var i interface{}
	var mine MyInt = 20
	i = mine
	i2 := <b>i.(MyInt)</b>
	fmt.Println(i2 + 1)
}
</code></pre>
* Use the comma ok idiom to avoid a panic.
<pre><code>type MyInt int
type MyFloat float64

func main() {
	var i interface{}
	var mine MyFloat = 1.234
	i = mine
	//f := i.(MyInt) // panics
	f, ok := i.(MyInt)
	if !ok {
		fmt.Println("unexpected type for", i)
	}
	fmt.Println(f + 1)
}
</code></pre>
* Use a __type switch__ instead when an interface can be of more than one type.
<pre><code>var t interface{}
t = functionOfSomeType()
<b>switch t := t.(type)</b> {
default:
    fmt.Printf("unexpected type %T\n", t)     // %T prints whatever type t has
case bool:
    fmt.Printf("boolean %t\n", t)             // t has type bool
case int:
    fmt.Printf("integer %d\n", t)             // t has type int
case &ast;bool:
    fmt.Printf("pointer to boolean %t\n", &ast;t) // t has type *bool
case &ast;int:
    fmt.Printf("pointer to integer %d\n", &ast;t) // t has type *int
}
</code></pre>

# Errors
* Effective Go [Errors](https://go.dev/doc/effective_go#errors)
* go.dev/ref/spec [Errors](https://go.dev/ref/spec#Errors)
* The `error` type is an interface that defines a single method:
<pre><code>type <b>error</b> interface {
    Error() string
}
</code></pre>
* Error messages should not be capitalized nor end with punctuation or a newline.
* Call [errors.New()](https://pkg.go.dev/errors#New) to create a `error` that has just a message.
<pre><code>err := <b>errors.New</b>("emit macho dwarf: elf header corrupted")
</code></pre>
* Another way to create a simple string `error` is by calling `fmt.Errorf()`. This
  allows formatting the string using the formatting package.
<pre><code>const name, id = "bimmler", 17
err := <b>fmt.Errorf</b>("user %q (id %d) not found", name, id)
</code></pre>
* You can define your own error type. Here's an example that has a status code, along with a message.
<pre><code>type Status int

const (
	InvalidLogin Status = iota + 1
	NotFound
)

type StatusErr struct {
	Status  Status
	Message string
}

func (se StatusErr) Error() string {
	return se.Message
}
</code></pre>
* By convention an `error` is returned as the last return value from
  a function. Return `nil` if there was no error.
<pre><code>func calcRemainderAndMod(numerator, denominator int) (int, int, error) {
	if denominator == 0 {
		return 0, 0, errors.New("denominator is 0")
	}
	return numerator / denominator, numerator % denominator, nil
}
</code></pre>
* Use `if` to check for and handle errors.
<pre><code>numerator := 20
denominator := 3
result, remainder, <b>err</b> := calcRemainderAndMod(numerator, denominator)
<b>if err != nil</b> {
	fmt.Println(err)
	os.Exit(1)
}
fmt.Println(result, remainder)
</code></pre>
* Errors can be __wrapped__ to create an error chain. Use `%w` to do this with `fmt.Errorf`, for
  a simple string `error` that wraps an `error`. By convention write `: %w` at the end of the format
  string and pass the `error` to wrap as the last parameter.
<pre><code>func fileChecker(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return <b>fmt.Errorf("in fileChecker: %w", err)</b>
	}
	f.Close()
	return nil
}
</code></pre>
* Call [errs.Unwrap()](https://pkg.go.dev/errors#Unwrap) to unwrap the error.
<pre><code>func main() {
	err := fileChecker("not_here.txt")
	if err != nil {
		fmt.Println(err)
		if wrappedErr := <b>errors.Unwrap(err)</b>; wrappedErr != nil {
			fmt.Println(wrappedErr)
		}
	}
}
</code></pre>
* Errors can be wrapped in a custom `error` by implementing the `Unwrap()` method.
<pre><code>type StatusErr struct {
	Status  Status
	Message string
    <b>Err     error</b>
}

<b>func (se StatusErr) Unwrap() error {
	return se.Err
}</b>
</code></pre>
* If all you want to include in your `error` is the message from another
  `error`, use `%v` instead of `%w`.
<pre><code>err := internalFunction()
if err != nil {
    return fmt.Errorf("internal failure <b>%v</b>", err)
}
</code></pre>
* Call [errors.Is()](https://pkg.go.dev/errors#Is) to determine if a particular error
  is in the error chain of a given error.
<pre><code>func fileChecker(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("in fileChecker: %w", err)
	}
	f.Close()
	return nil
}

func main() {
	err := fileChecker("not_here.txt")
	if err != nil {
		if <b>errors.Is(err, os.ErrNotExist)</b> {
			fmt.Println("That file doesn't exist")
		}
	}
}
</code></pre>
* `errors.Is()` uses the `==` operator to compare errors. If your `error` has fields that
  don't support `==` (e.g. a slice), implement the `Is()` method on the error to do
  a custom comparison.
<pre><code>type MyErr struct {
	Codes []int
}

func (me MyErr) Error() string {
	return fmt.Sprintf("codes: %v", me.Codes)
}

<b>func (me MyErr) Is(target error) bool</b> {
	if me2, ok := target.(MyErr); ok {
		return reflect.DeepEqual(me, me2)
	}
	return false
}
</code></pre>
* The `Is()` method could also be written to do a custom comparison that, for example,
  just looks at certain fields; e.g. `ResourceErr` on page 272 of LGO.
* Use [errors.As()](https://pkg.go.dev/errors#As) to find an `error` in the error chain.
  This is similar to `Is()`, but provides a reference to the matching `error`.
* A __panic__ causes the stack to unwind and any `defer` code to be run at each
  level in the stack. When `main()` is reached, a stack trace is printed and
  the program exits.
* Call [panic()](https://pkg.go.dev/builtin#panic) to generate a panic programmatically.
* Call [recover()](https://pkg.go.dev/builtin#recover) to recover from a panic. This must
  be done from within a `defer` function, since only `defer` functions are run once a panic happens.
<pre><code>func div60(i int) {
	defer func() {
		if <b>v := recover()</b>; v != nil {
			fmt.Println(v)
		}
	}()
	fmt.Println(60 / i)
}

func main() {
	for _, val := range []int{1, 2, 0, 6} {
		div60(val)
	}
}
</code></pre>
* To print the stack trace from an error use `fmt.Printf()` and the output verb `(%+v)`. (LGO p280)

# Documentation From Comments
* go.dev/doc [Go Doc Comments](https://go.dev/doc/comment)
* Place a comment directly before the item being documented.
* Start the comment with `//` followed by the name of the item.
* Use a blank comment to break your comment into multiple paragraphs.
* Comments before the package declaration create __packpage-level comments__.
  Although, if you have lengthy comments for a package, place them in a file
  called `doc.go`.
* Use `go doc` to view godocs.
  * `go doc PACKAGE_NAME` displays godocs for the specified package.
  * `go doc PACKAGE_NAME.IDENTIFIER` displays godocs for the specified identifier.
* Use [go list -m](https://go.dev/ref/mod#go-list-m) to __list the current module and all its dependencies__.

# Generics
* go.dev/ref/spec [Type parameter declarations](https://go.dev/ref/spec#Errors)
* Use __type parameters__ to make function __parameters generic__. 
<pre><code>// Index returns the index of x in s, or -1 if not found. The parameters can be
// of any type that fulfills the built-in constraint `comparable`.
func Index[T comparable](s []T, x T) int {
	for i, v := range s {
		if v == x {
			return i
		}
	}
	return -1
}

func main() {
	// Index works on a slice of ints
	si := []int{10, 20, 15, -10}
	fmt.Println(Index(si, 15))

	// Index also works on a slice of strings
	ss := []string{"foo", "bar", "baz"}
	fmt.Println(Index(ss, "hello"))
}

// Prints:
// 2
// -1
</code></pre>
* __Generic types__ are possible too.
<pre><code>// List represents a singly-linked list that holds values of any type.
type List[T any] struct {
	next &ast;List[T]
	val  T
}
</code></pre>
