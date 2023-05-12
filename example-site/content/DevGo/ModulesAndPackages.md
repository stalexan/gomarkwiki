# Modules
* __Repositories__ contain __modules__ contain __packages__.
* Generally there should only be __one module per repository__.
* Intro:
  * go.dev/blog [Using Go Modules](https://go.dev/blog/using-go-modules) (2019)
  * github.com/golang/go/wiki [Modules](https://github.com/golang/go/wiki/Modules)
* From go.dev [Documentation](https://go.dev/doc/):
  * [Developing and publishing modules](https://go.dev/doc/modules/developing)
  * [Module release and versioning workflow](https://go.dev/doc/modules/release-workflow)
  * [Managing module source](https://go.dev/doc/modules/managing-source)
  * [Developing a major version update](https://go.dev/doc/modules/major-version)
  * [Publishing a module](https://go.dev/doc/modules/publishing): A module's
    version number is created by tagging a Git commit with, for example, `v0.1.0`.
  * [Module version numbering](https://go.dev/doc/modules/version-numbers)
  * [Managing dependencies](https://go.dev/doc/modules/managing-dependencies)
  * [go.mod file reference](https://go.dev/doc/modules/gomod-ref)
* Reference:
  * go.dev/ref [Go Modules Reference](https://go.dev/ref/mod)
* Misc:
  * go.dev/blog [Go Modules: v2 and Beyond](https://go.dev/blog/v2-go-modules):
    How to create a new major version of a module.
* The __root of a module__ is a directory that has a __`go.mod` file__. Create `go.mod`
  using [go mod init](https://go.dev/ref/mod#go-mod-init) from the module's
  root directory. Here `example.com/hello` is the module path for the new module.
<pre><code>go mod init example.com/hello
</code></pre>
* __Dependencies to other modules are added by adding an `import` statement__ to
  a `.go` code file.
<pre><code>import "rsc.io/quote"
</code></pre>
* __Download the dependency__ with [go get](https://go.dev/ref/mod#go-get). This will get the latest version.
<pre><code>go get rsc.io/quote
</code></pre>
* The next time the file is built with `go build` or `go test`, __a `require`
  statement will be added to the `go.mod` file__ listing the dependency along
  with any indirect depencencies and their version numbers.
<pre><code>require (
        golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c // indirect
        rsc.io/quote v1.5.2 // indirect
        rsc.io/sampler v1.3.0 // indirect
)
</code></pre>
* Use [go list -m](https://go.dev/ref/mod#go-list-m) to __list the current module and all its dependencies__.
<pre><code>$ gdr go list -m all
example.com/hello
golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c
rsc.io/quote v1.5.2
rsc.io/sampler v1.3.0
</code></pre>
* __The file `go.sum`__ is created and lists the hashes for what was downloaded; e.g.
<pre><code>golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c h1:qgOY6WgZOaTkIIMiVjBQcw93ERBE4m30iBm00nkL0i8=
golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
rsc.io/quote v1.5.2 h1:w5fcysjrx7yqtD/aO+QwRjYZOKnaM9Uh2b40tElTs3Y=
rsc.io/quote v1.5.2/go.mod h1:LzX7hefJvL54yjefDEDHNONDjII0t9xZLPXsUe+TKr0=
rsc.io/sampler v1.3.0 h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/QiW4=
rsc.io/sampler v1.3.0/go.mod h1:T1hPZKmBbMNahiBKFy5HrXp6adAjACjK9JXDnKaTXpA=
</code></pre>
* __Upgrade a module to its latest version__ with [go get](https://go.dev/ref/mod#go-get).
  Currently `golang.org/x/text`, a dependency downloaded indirectly, is using
  an untagged version (`v0.0.0-20170915032832-14c0d48ead0c`).
<pre><code>
$ gdr go get golang.org/x/text
go: upgraded golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c => v0.8.0

$ gdr go list -m all
example.com/hello
golang.org/x/mod v0.8.0
golang.org/x/sys v0.5.0
golang.org/x/text v0.8.0
golang.org/x/tools v0.6.0
rsc.io/quote v1.5.2
rsc.io/sampler v1.3.0

$ cat go.mod
module example.com/hello

go 1.20

require (
        golang.org/x/text v0.8.0 // indirect
        rsc.io/quote v1.5.2 // indirect
        rsc.io/sampler v1.3.0 // indirect
)
</code></pre>
* __Switch to a specific version__ by specifying the version number. Use 
  [go list -m](https://go.dev/ref/mod#go-list-m) to __view available version numbers__.
<pre><code>$ go list -m -versions rsc.io/sampler
rsc.io/sampler v1.0.0 v1.2.0 v1.2.1 v1.3.0 v1.3.1 v1.99.99

$ go get rsc.io/sampler<b>@v1.3.1</b>
go: upgraded rsc.io/sampler v1.3.0 => v1.3.1
</code></pre>
* Go uses [Semantic Versioning](https://semver.org/) for version numbers.
  A change in the major version number indicates breaking changes and is
  imported independently from other major versions of the module. This allows
  more than one major version to be used at the same time.  Import statements
  must be added or change because the module path for each major version is
  unique. Each major version after the first must have the version number
  appended to it's path, with `v2` for version 2, `v3` for version 3, and so on.
* __Import a new major version by adding a new import statement__. Here we alias the
  module name so be able to differentiate between the two major versions.
<pre><code>import (
    "rsc.io/quote"
    <b>quoteV3 "rsc.io/quote/v3"</b>
)

func Hello() string {
    return quote.Hello()
}

func Proverb() string {
    return <b>quoteV3</b>.Concurrency()
}
</code></pre>
* __Get the new major version__ using the new module path.
<pre><code>go get rsc.io/quote/v3
</code></pre>
* Later when all calls have been upgraded to use the new major version, __the import statement
  for the old version can be removed__, along with the alias to the new version.
<pre><code>import "rsc.io/quote/v3"

func Hello() string {
    return quote.HelloV3()
}

func Proverb() string {
    return quote.Concurrency()
}
</code></pre>
* __Prune versions no longer needed__ with [go mod tidy](https://go.dev/ref/mod#go-mod-tidy).
  This will also download any missing dependencies.
<pre><code>$ go mod tidy

$ gdr go list -m all
example.com/hello
golang.org/x/mod v0.8.0
golang.org/x/sys v0.5.0
golang.org/x/text v0.8.0
golang.org/x/tools v0.6.0
rsc.io/quote/v3 v3.1.0
rsc.io/sampler v1.3.1

$ cat go.mod
module example.com/hello

go 1.20

require rsc.io/quote/v3 v3.1.0

require (
        golang.org/x/text v0.8.0 // indirect
        rsc.io/sampler v1.3.1 // indirect
)
</code></pre>
* [Minimal version selection (MVS)](https://go.dev/ref/mod#minimal-version-selection):
  A module will only be imported once. If there are depencencies on more than
  one version, Go will import the lowest version that is declared to work
  across all modules that need it. (LGO p 309)
* Use `go get -u` to __update dependencies to their latest versions__.  Here
  `./...` causes the updates to happen not just for the current directory but
  recusively for for all packages in any subdirectories as well.
<pre><code>go get -u ./...
go mod tidy
</code></pre>

# Publishing Modules
* [pkg.go.dev](https://pkg.go.dev/) is an index of all open source Go projects.
* Public modules are automatically placed in the index when their source is
  made available in a public repo that has a `go.mod` and `go.sum` file. Make
  sure to include a `LICENSE` file too, in the root of your repo. (LGO p 316)
* The Go community favors permissive licenses such as BSD, MIT, and Apache. This is since
  Go compiles in all 3rd-party code directly and a non-permissive license would require
  any application that uses Go code to be open source as well. (LGO p 316)

# Module Proxy Servers
* __By default `go get` does not fetch directly from the original source code
  repo for a module but from a proxy server run by Google__. (LGO p 317)
* proxy.golang.org [Go Module Mirror, Index, and Checksum Database](https://proxy.golang.org/)
* If a module isn't present in the proxy server, the proxy server downloads it
  and stores a copy. The checksums from `go.sum` are stored as well, to help
  protect against modifications to a module (either malicious or inadvertent.) (LGO p 318)
* For those that object to sending librar requests to Google, you can __disable
  proxying__ entirely by setting the `GOPROXY` environment variable to
  `direct`.  Note, though, if the version you depend on is removed, you won't
  have access to it. (LGO p 318)
* Or, another option is to __run your own proxy server__. [The Athens Project](https://docs.gomods.io/)
  is an open source proxy server.
* Go modules can come from __private repositories__ as well. This requires not using 
  `proxy.golang.org`, using either of the above methods. (LGO p 320)

# Packages, Imports, and Exports
* go.dev/ref/spec [Packages](https://go.dev/ref/spec#Packages)
* Effective Go [Package names](https://go.dev/doc/effective_go#package-names): Keep them short.
* Executable commands must always use package `main`. 
* Package names should be one word all lower case; see Effective Go [Package names](https://go.dev/doc/effective_go#package-names).
* __Capitalization exports__ an identifier. An identifier whose name with an uppercase letter is exported.
* The __`package` clause__ defines a package.
<pre><code>package math
</code></pre>
* __Every Go file in a given directory must have an identical package clause.__
* __Import a package__ with the `import` clause, and a package path. The
  __package path__ is the __module path with the package path appended to it__.
  Any package not in the standard library should use the full module path. This
  is not required for packages within a module, it's a recommended practice, to
  clarify what's being imported and to make refactoring easier. (LGO p285)
<pre><code>package main

<b>import (
	"fmt"
	"github.com/learning-go-book/package_example/formatter"
	"github.com/learning-go-book/package_example/math"
)</b>

func main() {
	num := <b>math</b>.Double(2)
	output := <b>print</b>.Format(num)
	fmt.Println(output)
}
</code></pre>
* The go.mod for this is:
<pre><code>module "github.com/learning-go-book/package_example"

go 1.15
</code></pre>
* __The name of a package is determined by its `package` clause, and not its
  directory path__.  Although, its a __best practice to make the directory path
  the same as the package path__. The above call to the `Format()` method
  is `print.Format()` and not `formatter.Format()` because the package
  clause in `package_example/formatter/formatter.go` is `package print`.
* The import statement can be used to __rename a package__ when there's a naming conflict.
<pre><code>import (
	<b>crand</b> "crypto/<b>rand</b>"
	"encoding/binary"
	"fmt"
	"math/<b>rand</b>"
)

func seedRand() &ast;<b>rand</b>.Rand {
	var b [8]byte
	_, err := <b>crand</b>.Read(b[:])
	if err != nil {
		panic("cannot seed with cryptographic random number generator")
	}
	r := <b>rand</b>.New(rand.NewSource(int64(binary.LittleEndian.Uint64(b[:]))))
	return r
}
</code></pre>
* To share code between packages within a module create a package called `internal`. (LGO p 294)
* Circular dependencies are not allowed.

# Init Functions
* Effective Go [The init function](https://go.dev/doc/effective_go#init)
* Each package can have any number of `init()` functions. They're called after 
  package variable declarations. Order of calls is not guaranteed.
<pre><code>func init() {
    if user == "" {
        log.Fatal("$USER not set")
    }
    if home == "" {
        home = "/home/" + user
    }
 }
</code></pre>
* `main.main()` is called after all `init()` functions have been run.
