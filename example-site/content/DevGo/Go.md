# Go; 2022-10
* Wikipedia [Go (programming language)](https://en.wikipedia.org/wiki/Go_(programming_language)):
  "Go is a statically typed, compiled programming language designed at Google by
  Robert Griesemer, Rob Pike, and Ken Thompson. It is syntactically similar to
  C, but with memory safety, garbage collection, structural typing, and
  CSP-style concurrency. It is often referred to as Golang because of its
  former domain name, golang.org, but its proper name is Go."
* Wikipedia [Communicating sequential processes](https://en.wikipedia.org/wiki/Communicating_sequential_processes): 
  "(CSP) is a formal language for describing patterns of interaction in
  concurrent systems. It is a member of the family of mathematical theories of
  concurrency known as process algebras, or process calculi, based on message
  passing via channels."
* Website: [The Go Programming Language](https://go.dev/)
* go.dev [Documentation](https://go.dev/doc/)
* go.dev [The Go Programming Language Specification](https://go.dev/ref/spec)
* go.dev [FAQ](https://go.dev/doc/faq)
* Source code: github.com/golang [go](https://github.com/golang/go)
  * Lanauges: 88% Go, 6% assembly, 6% HTML
* go.dev [The Go Playground](https://go.dev/play/): Create and run Go programs in a browser.
* Mailing lists (require a Google email address):
  * [golang-announce](https://groups.google.com/g/golang-announce): Security updates, etc.
  * [golang-nuts](https://groups.google.com/g/golang-nuts): General discussion.

# Community
* [Go Forum](https://forum.golangbridge.org/)

# Misc Resources
* [Awesome Go](https://awesome-go.com/)
* github.com golang/go [Projects](https://github.com/golang/go/wiki/Projects)

# Learning Go
* go.dev [Tour of Go](https://go.dev/tour/)
* go.dev [How to Write Go Code](https://go.dev/doc/code)
* go.dev [Effective Go](https://go.dev/doc/effective_go)
  "This document gives tips for writing clear, idiomatic Go code. It augments
  the [language specification](https://go.dev/ref/spec), the 
  [Tour of Go](https://go.dev/tour/), and 
  [How to Write Go Code](https://go.dev/doc/code.html), all of which you should
  read first."
* Book "[Learning Go: An Idiomatic Approach to Real-World Go Programming](https://www.amazon.com/Learning-Go-Idiomatic-Real-World-Programming/dp/1492077216/)" by Jon Bodner (LGO)
* exercism.org [Go](https://exercism.org/tracks/go): Doesn't appear to have much on concurrency.
* go.dev [Guided learning journeys](https://go.dev/learn/#guided-learning-journeys)
  * [Go by Example](https://gobyexample.com/)

# GopherCon 2022
* github.com/gophercon [2022-talks](https://github.com/gophercon/2022-talks)
* youtube.com [Gopher Academy](https://www.youtube.com/@GopherAcademy): "Gopher Academy is the creative force behind GopherCon"

# Code Style Guidelines
* github.com golang/go [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

# Exercises to Learn About Concurrency; 2023-03-21
* From ChatGPT:
<pre><code>Here are some specific coding exercises on websites for learning about concurrency in Go:

    Exercism: Exercism has several exercises on concurrency in Go, including
    "Clock," "Parallel Letter Frequency," and "Concurrent Tree Traversal."
    These exercises cover topics such as channels, goroutines, and
    synchronization. Each exercise comes with a set of test cases to verify
    your solution.

    HackerRank: HackerRank has a section dedicated to Go concurrency
    challenges, including "Go Channels," "Go Routines," and "Go Buffered
    Channels." These challenges test your knowledge of concurrency concepts
    such as communication, synchronization, and deadlock prevention.

    LeetCode: LeetCode has a variety of Go problems that involve concurrency,
    such as "Print in Order," "The Dining Philosophers," and "The H2O Problem."
    These problems test your understanding of concurrency patterns such as
    mutexes, semaphores, and condition variables.

    Codewars: Codewars has a few Go katas that focus on concurrency, such as
    "Parallel Computing in Go," "Mutex and Channels," and "Deadlock Detection."
    These katas cover topics such as race conditions, message passing, and
    deadlock prevention.

    Gophercises: Gophercises has a couple of exercises on Go concurrency,
    including "Concurrency Patterns," "Rate Limiting with Channels," and
    "Concurrent Word Count." These exercises focus on implementing common
    concurrency patterns and techniques, such as fan-out/fan-in, rate limiting,
    and context cancellation.
</code></pre>

# Installing Go; 2023-02
* Install is only needed on development machines. Go programs compile to
  a single binary and do not need any additional software to be installed in
  order to run them.
* go.dev [Download and install](https://go.dev/doc/install)
* go.dev [Downloads](https://go.dev/dl/)
* Latest install: <https://go.dev/dl/go1.20.1.linux-amd64.tar.gz>
* SHA256 checksum: 
<pre><code>000a5b1fca4f75895f78befeb2eecf10bfff3c428597f3f1e69133b63b911b02
</code></pre>
* As root:
<pre><code>$ rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.1.linux-amd64.tar.gz
</code></pre>
* Add to `~/.bashrc_custom`:
<pre><code># Add Go to PATH.
if [ -d "/usr/local/go" ] ; then
    PATH="$PATH:/usr/local/go/bin"
fi
</code></pre>
* Test install:
<pre><code>$ go version
go version go1.20.1 linux/amd64
</code></pre>

# Checking Signature on Install; 2023-05-08
* It turns out there's a signature on Go installs (e.g.
  <https://go.dev/dl/go1.20.4.linux-amd64.tar.gz> has a
  a <https://go.dev/dl/go1.20.4.linux-amd64.tar.gz.asc>).
* The Go website doesn't say anything about them, though, that I can see, including they don't say who signed them.
  It doesn't even list them on the releases page. You have to know they're there, and attempt to download them.
* I learned about them from the Restic release builder Docker image.
* I'd like to use, the signatures.
* The Restic installer has a `linux_signing_key.pub` file with the key used to sign, and imports it.
* When I look at the used to sign `go1.20.4.linux-amd64.tar.gz` I see:
```
$ gpg --verify go1.20.4.linux-amd64.tar.gz.asc
gpg: assuming signed data in 'go1.20.4.linux-amd64.tar.gz'
gpg: Signature made Wed 26 Apr 2023 12:56:03 PM CST                                                                                                   
gpg:                using RSA key 4EB27DB2A3B88B8B
gpg: Can't check signature: No public key     

$ gpg --recv-keys 4EB27DB2A3B88B8B
gpg: key 7721F63BD38B4796: new key but contains no user ID - skipped
gpg: Total number processed: 1
gpg:           w/o user IDs: 1                                             

$ gpg --keyserver keyserver.ubuntu.com --recv-keys 4EB27DB2A3B88B8B
gpg: key 7721F63BD38B4796: public key "Google Inc. (Linux Packages Signing Authority) <linux-packages-keymaster@google.com>" imported
gpg: Total number processed: 1
gpg:               imported: 1

$ gpg --with-subkey-fingerprint --list-keys 7721F63BD38B4796
pub   rsa4096 2016-04-12 [SC]
      EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796                           
uid           [ unknown] Google Inc. (Linux Packages Signing Authority) <linux-packages-keymaster@google.com>
sub   rsa4096 2023-02-15 [S] [expires: 2026-02-14]
      A5F483CD733A4EBAEA378B2AE88979FB9B30ACF2                                                                                                        
sub   rsa4096 2021-10-26 [S] [expires: 2024-10-25]
      8461EFA0E74ABAE010DE66994EB27DB2A3B88B8B    
```
* I see they key described here: google.com [Linux Software Repositories](https://www.google.com/linuxrepositories/)
> The Linux Package Signing Keys 
>
> Google’s Linux packages are signed with GNU Privacy Guard (GPG) keys.
> Google’s packages will automatically configure your package manager to verify
> product updates with the public signing key, but you may also install it
> separately if, for instance, you want to verify the integrity of an initial
> package download. Follow the instructions below to manually configure your
> package manager to use the keys. 
>
> Key Details 
> * Download: <https://dl.google.com/linux/linux_signing_key.pub>
> * Key ID: Google, Inc. Linux Package Signing Key <linux-packages-keymaster@google.com>\
>   Fingerprint: `4CCA 1EAF 950C EE4A B839 76DC A040 830F 7FAC 5991`
> * Google, Inc. (Linux Package Signing Authority) <linux-packages-keymaster@google.com>\
>   Fingerprint: `EB4C 1BFD 4F04 2F6D DDCC EC91 7721 F63B D38B 4796`
* Restic includes the `linux_signing_key.pub` in it's repo, which would be one way to do this, although
  when the subkey changes the file would need to be updated, and it's really more of just another check
  based on SSL since there's no key spelled out or fingerprint.
* The official [Go Docker image](https://hub.docker.com/_/golang) does a `--recv-keys` from `keyserver.ubuntu.com`, which is what
  I think I'll do too. It does a SHA256 checksum check as well:
```
wget -O go.tgz.asc "$url.asc"; \
wget -O go.tgz "$url" --progress=dot:giga; \
echo "$sha256 *go.tgz" | sha256sum -c -; \
\
# https://github.com/golang/go/issues/14739#issuecomment-324767697
GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
# https://www.google.com/linuxrepositories/
gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796'; \
# let's also fetch the specific subkey of that key explicitly that we expect "go.tgz.asc" to be signed by, just to make sure we definitely have it
gpg --batch --keyserver keyserver.ubuntu.com --recv-keys '2F52 8D36 D67B 69ED F998  D857 78BD 6547 3CB3 BD13'; \
gpg --batch --verify go.tgz.asc go.tgz; \
gpgconf --kill all; \
rm -rf "$GNUPGHOME" go.tgz.asc; \
```

# Configuring
* pkg.go.dev/cmd/go [Print Go environment information](https://pkg.go.dev/cmd/go#hdr-Print_Go_environment_information)
* pkg.go.dev/cmd/go [Environment variables](https://pkg.go.dev/cmd/go#hdr-Environment_variables)
* pkg.go.dev/cmd/go [GOPATH environment variable](https://pkg.go.dev/cmd/go#hdr-GOPATH_environment_variable)
* __Go environment variables have default values__ that will be used when they're
  not set directly in the shell. To see where these default values are stored:
<pre><code>go env GOENV
</code></pre>
* To see effective values of all environment variables:
<pre><code>go env
</code></pre>
* Set `GOPATH` to the location where you want dependencies installed:
<pre><code>export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
</code></pre>

# Uninstalling; 2023-03-06
* Delete go, as `root`:
<pre><code>rm -rf /usr/local/go/
</code></pre>
* Delete locally installed modules and apps, as `sean`:
<pre><code>rm -rf /home/sean/go
</code></pre>
* Remove settings from `~/.bashrc_custom`:
<pre><code># Add Go to PATH.
if [ -d "/usr/local/go" ] ; then
    PATH="$PATH:/usr/local/go/bin"
fi

# Configure GOPATH.
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
</code></pre>

# Create Dev Repo For Work On logan6; Wed 2023-02-15
* Create bare repo as beorn on elias:
<pre><code>$ mkdir /home/beorn/repos/dev/go.git 
$ cd !$
$ git init --bare
</code></pre>
* Clone repo as sean on logan5:
<pre><code>$ cd /home/sean/dev/elias
$ git clone ssh://beorn@elias:/home/beorn/repos/dev/go.git
</code></pre>

# Building and Running
* Run a go program without saving binary using [go run](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program).
  This creates the binary in a temp directory that is __deleted after the run__.
<pre><code>go run hello.go
</code></pre>
* Build an executable with [go build](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies).
  This creates an executable file named `hello`:
<pre><code>go build hello.go
</code></pre>

# Compiling and Installing Packages
* Install the `hey` executable using [go install](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies):
<pre><code>$ go install github.com/rakyll/hey@latest     
go: downloading github.com/rakyll/hey v0.1.4                                                             
go: downloading golang.org/x/net v0.0.0-20181017193950-04a2e542c03f                                      
go: downloading golang.org/x/text v0.3.0                                                                 

$ which hey
/home/sean/go/bin/hey
</code></pre>
* `go install` can be used to install the current module too. The following
  assumes we're in the directory `hello` which contains a `go.mod` with 
  `module example/user/hello` and a file `hello.go` with `package main`.
  (This example is from go.dev [How to Write Go Code](https://go.dev/doc/code).)
<pre><code># All are equivalent and will install the executable `hello` to ${GOPATH}/bin
go install example/user/hello
go install .
go install
</code></pre>

# Modules
* go.dev [Go Modules Reference](https://go.dev/ref/mod)
* Create a module using [go mod init](https://go.dev/ref/mod#go-mod-init). This
  will create the file `go.mod` in the `ch01` directory:
<pre><code>$ cd ch01
$ go mod init ch01
</code></pre>

# Formatting Code
* Format code using [go fmt](https://pkg.go.dev/cmd/gofmt) to format code. Here `./...` specifies the module defined in the current directory.
<pre><code>go fmt ./...
</code></pre>

# Linters
* Use [go vet](https://pkg.go.dev/cmd/vet) to check code.
<pre><code>go fmt ./...
</code></pre>
* go.dev [lint package](https://pkg.go.dev/golang.org/x/lint): "Golint is
  deprecated and frozen. There's no drop-in replacement for it, but tools such
  as Staticcheck and go vet should be used instead."
* [Staticcheck](https://staticcheck.io/):
  * "is a state of the art linter for the Go programming language".
  * [Getting started](https://staticcheck.io/docs/getting-started/)
* Install Staticcheck:
<pre><code>go install honnef.co/go/tools/cmd/staticcheck@latest
</code></pre>
* Check current package:
<pre><code>staticcheck .
</code></pre>
* Check all packages:
<pre><code>staticcheck ./...
</code></pre>

# Docker Container; 2023-03-03 <span id=docker=230303 />
* I just build [Hugo](../Linux/Hugo.html) and there were lots of dependencies.
* I think it probably makes more sense to development in a container, similar to what
  I've done with Rust, Python, etc.
* The [Hugo install instructions](https://gohugo.io/installation/linux/) point
  to hub.docker.com [klakegg/hugo](https://hub.docker.com/r/klakegg/hugo).
* I think I'd prefer a more generic solution, though, for general Go development.
  The actual Hugo build from source is straightforward.
* There are offical Docker images for Go based on Debian: hub.docker.com [golang](https://hub.docker.com/_/golang)
* Use this as a base image, and then customize it similar to what I've done for 
  Python and Rust.
* Repo: `~/dev/elias/docker/go-dev`
* The official Go Docker image has some suggestions on how to use it: <https://hub.docker.com/_/golang>
* The approach seems to be to run `docker run` from the host to do Go specific
  things: compile an app, run it, etc.
* So I would do most things from the host, and not within the container.
* I like this. I wouldn't need to write a bashig script for use inside the container.
* See how this goes.

# Running Go in a Container and ChatGPT Advice; 2023-03-07
* I'm having fun with ChatGPT and asked it how to run an Go version of "Hello
  World" inside the official Docker container for Go: "Run a Go version of
  "hello world" inside the official Docker container for Go, using only the
  command line of the host machine, and never running commands interactively
  inside the container.
* It has me create a directory called `myapp`, create the `hello.go` there and then 
  run:
<pre><code>docker run --rm -v "$(pwd)":/go/src/myapp -w /go/src/myapp golang go run hello.go
</code></pre>
* So that looks straight forward.
* Now there's the question of how to install modules and apps, and reuse installs
  between runs.
* To install it says create a hey directory and then run:
<pre><code>docker run --rm -v "$(pwd)":/go/src/hey -w /go/src/hey golang go get -u github.com/rakyll/hey
</code></pre>
* Run `hey` with:
<pre><code>docker run --rm -v "$(pwd)":/go/src/hey -w /go/src/hey golang hey http://example.com
</code></pre>
* It says this will preserve the install across runs, but it looks like `GOPATH` for
  the official docker images is set to `/go/bin` which I think gets recreated between runs?
* Give this a try. I think I'll need to create a separate volume for installs, though.
  Is there a best practice recommended for the official Go Docker container?
* Download the official Go Docker image: 
<pre><code>docker pull golang:1.20-bullseye
</code></pre>
* Create a local directory:
<pre><code>mkdir ~/tmp/try-hey
cd !$
</code></pre>
* Install:
<pre><code>$ docker run --rm -v "$(pwd)":/go/src/hey -w /go/src/hey golang:1.20-bullseye go get -u github.com/rakyll/hey
go: go.mod file not found in current directory or any parent directory.
        'go get' is no longer supported outside a module.
        To build and install a command, use 'go install' with a version,
        like 'go install example.com/cmd@latest'
        For more information, see https://golang.org/doc/go-get-install-deprecation
        or run 'go help get' or 'go help install'.
</code></pre>
* So it looks like the ChatGPT advice is out-of-date.
* `go install` works:
<pre><code>$ docker run --rm -v "$(pwd)":/go/src/hey -w /go/src/hey golang:1.20-bullseye go install github.com/rakyll/hey@latest
</code></pre>
* Running doesn't work, though, since there's no volume to preserve the install:
<pre><code>$ docker run --rm -v "$(pwd)":/go/src/hey -w /go/src/hey golang:1.20-bullseye go hey https://www.alexan.org
</code></pre>
* `GOPATH` in the official container is set to `/go/bin`, and `go` itself is installed to `/usr/local/go`.
* So it seems the solution is to create a volume for `/go/bin`.
<pre><code>docker volume create go-bin
</code></pre>
* Now the install is:
<pre><code>docker run --rm -v go-bin:/go/bin golang:1.20-bullseye go install github.com/rakyll/hey@latest
</code></pre>
* The run is:
<pre><code>docker run --rm -v go-bin:/go/bin golang:1.20-bullseye hey https://www.alexan.org
</code></pre>

# Running Inside Container As Regular User; 2023-03-07
* I see the default user is `root`:
<pre><code>$ docker run --rm -v go-bin:/go/bin golang:1.20-bullseye id
uid=0(root) gid=0(root) groups=0(root)
</code></pre>
* It would be safer to run as non-priviledged user.
* Create a Dockerfile for this, but with the assumption that commands won't be run
  interactively, and so `bashig` isn't needed.
* See `sean@logan6:./dev/elias/docker/go-dev`
* Create image:
<pre><code>./build
</code></pre>
* Create volume needed by `/home/sean/go/bin`:
<pre><code>$ docker volume create go-dev-bin
</code></pre>
* Run commands inside the container:
<pre><code>$ docker run --rm -v go-dev-bin:/home/sean/go/bin stalexan/go-dev go install github.com/rakyll/hey@latest
...
$ docker run --rm -v go-dev-bin:/home/sean/go/bin stalexan/go-dev which hey
/home/sean/go/bin/hey
$ docker run --rm -v go-dev-bin:/home/sean/go/bin stalexan/go-dev hey https://www.alexan.org
...
</code></pre>

# Run Sample Code Inside go-dev Docker Container; 2023-03-07
* So that allows me to run Go inside a Docker container, as a regular user, and
  I can install modules and applications, and have the persist across runs.
* Now I want to run sample code for learning.
* For example, I have `sean@logan6:./dev/current/go/learning-go-book/ch01`.
* It has a `Makefile`, and `go.mod` with:
<pre><code>module ch01

go 1.20
</code></pre>
* Install `staticcheck`, a dependency the Makefile has.
<pre><code>$ docker run --rm -v go-dev-bin:/home/sean/go/bin stalexan/go-dev go install honnef.co/go/tools/cmd/staticcheck@latest
</code></pre>
* I want to write the code, and use Git, on the host.
* Build `ch01`:
<pre><code>$ docker run --rm -v go-dev-bin:/home/sean/go/bin -v ~/dev/current/go:/home/sean/dev -w /home/sean/dev/learning-go-book/ch01 stalexan/go-dev make
go fmt ./...
go vet ./...
staticcheck ./...
go build hello.go
</code></pre>
* Run `hello`:
<pre><code>$ docker run --rm -v go-dev-bin:/home/sean/go/bin -v ~/dev/current/go:/home/sean/dev -w /home/sean/dev/learning-go-book/ch01 stalexan/go-dev ./hello
Hello, world!
</code></pre>
* Experiment some more with this and then wrap it in a bash script.
* The bash script could be simple to start.
* Have two variable that define the volume for code: `DIRH` for host directory and `DIRC` for container directory:
<pre><code>DIRH=/home/sean/dev/current/go
DIRC=/home/sean/dev
</code></pre>
* Then have another for the working directory, relative to DEV_DIR_CONTAINER:
<pre><code>WORKDIR=learning-go-book/ch01
</code></pre>
* The command to run would just be the parameters passed to the bash script.
* The earlier command would become:
<pre><code>$ docker run --rm -v go-dev-bin:/home/sean/go/bin -v $DIRH:$DIRC -w ${DIRC}/{WORKDIR} stalexan/go-dev $@
</code></pre>
* Call the script `goc`.
* To make:
<pre><code>goc make
</code></pre>
* Thu 2023-03-30: Never use `go build`, only `go run` and `go install`. Otherwise, the binary
  is placed in a host directory and it's too easy to accidently run, which kind of defeats the
  purpose of running in a container.

# Tree Shaking; 2023-03-22
* The Go linker only includes the subset of the standard library that's needed by a binary.
  It uses what's called "tree shaking" to eliminate dead code.

# Security Advisories; 2023-03-22
* ChatGPT's answer to "What security advisories has there been against the Go runtime?"
<pre><code>Like all software, the Go runtime has had security advisories and
vulnerabilities over time. Here are some notable security advisories related to
the Go runtime:

    In 2016, a vulnerability was found in the Go crypto/tls package that could
    allow an attacker to bypass certificate verification when making HTTPS
    connections. This vulnerability was assigned CVE-2016-2107 and was fixed in
    Go versions 1.6.4, 1.5.4, and 1.4.7.

    In 2018, a vulnerability was discovered in the Go net/http package that
    could allow an attacker to conduct HTTP request smuggling attacks. This
    vulnerability was assigned CVE-2019-6486 and was fixed in Go versions
    1.11.5 and 1.10.8.

    In 2019, a vulnerability was found in the Go crypto/x509 package that could
    allow an attacker to bypass certificate verification in certain
    circumstances. This vulnerability was assigned CVE-2019-6487 and was fixed
    in Go versions 1.12.1 and 1.11.6.

    In 2020, a vulnerability was discovered in the Go crypto/elliptic package
    that could allow an attacker to craft inputs that would cause the package
    to enter an infinite loop, leading to a denial of service. This
    vulnerability was assigned CVE-2020-28362 and was fixed in Go versions
    1.15.6 and 1.14.13.

It is important to note that these vulnerabilities have been fixed in the
affected versions of Go, and it is generally recommended to keep your Go
installation up to date with the latest security patches. Additionally, it is
a good practice to follow secure coding practices and to regularly review and
audit your code for security vulnerabilities.
</code></pre>
* ChatGPT's answer to "Which of these security advisories resulted in Debian package updates?"
<pre><code>Here are the Debian security advisories related to the Go runtime
vulnerabilities mentioned in my previous answer:

    CVE-2016-2107: This vulnerability was addressed in Debian in the "golang"
    package version 2:1.6.3-1+deb8u5 for Debian 8 (Jessie), and the "golang"
    package version 2:1.6.4-1+deb8u2 for Debian 9 (Stretch).

    CVE-2019-6486: This vulnerability was addressed in Debian in the
    "golang-1.11" package version 1.11.5-1+deb10u1 for Debian 10 (Buster), and
    the "golang-1.10" package version 1.10.8-3+deb9u1 for Debian 9 (Stretch).

    CVE-2019-6487: This vulnerability was addressed in Debian in the
    "golang-1.12" package version 1.12.1-2+deb10u1 for Debian 10 (Buster), and
    the "golang-1.11" package version 1.11.6-1+deb9u1 for Debian 9 (Stretch).

    CVE-2020-28362: This vulnerability was addressed in Debian in the
    "golang-1.15" package version 1.15.6-1+deb11u1 for Debian 11 (Bullseye),
    and the "golang-1.14" package version 1.14.13-3+deb10u1 for Debian 10 (Buster).

It's worth noting that these package versions are the latest versions as of my
knowledge cutoff date of September 2021. It's possible that newer package
versions have been released since then to address these vulnerabilities or
other security issues.
</code></pre>

# Checking Signature on Go Image; 2023-03-27
* docs.docker.com [docker trust inspect](https://docs.docker.com/engine/reference/commandline/trust_inspect/)
* I'd like to check the Go signature on the Go image I'm using.
<pre><code>$ docker trust inspect golang:1.20-bullseye
[
    {
        "Name": "golang:1.20-bullseye",
        "SignedTags": [
            {
                "SignedTag": "1.20-bullseye",
                "Digest": "3d372ad0a86b36516db073c4a4e214307ec1fa00ec937f5b22630e6856ef87d9",
                "Signers": [
                    "Repo Admin"
                ]
            }
        ],
        "Signers": [],
        "AdministrativeKeys": [
            {
                "Name": "Root",
                "Keys": [
                    {
                        "ID": "c6b86f21ae272f3ae27b8da8a5762df97f09d6d0604ab49dd1d9920c6e25b65b"
                    }
                ]
            },
            {
                "Name": "Repository",
                "Keys": [
                    {
                        "ID": "fb57d64910e2f7fa4456e938c547398305f26c15c76e9de89f76e4f32e1fd0bc"
                    }
                ]
            }
        ]
    }
]
</code></pre>
* The "pretty" version is:
<pre><code>$ docker trust inspect --pretty golang:1.20-bullseye

Signatures for golang:1.20-bullseye

SIGNED TAG      DIGEST                                                             SIGNERS
1.20-bullseye   3d372ad0a86b36516db073c4a4e214307ec1fa00ec937f5b22630e6856ef87d9   (Repo Admin)

Administrative keys for golang:1.20-bullseye

  Repository Key:	fb57d64910e2f7fa4456e938c547398305f26c15c76e9de89f76e4f32e1fd0bc
  Root Key:	c6b86f21ae272f3ae27b8da8a5762df97f09d6d0604ab49dd1d9920c6e25b65b
</code></pre>
* So it's signed by "Repo Admin". It seems this is probably the person that has access to the listed keys.
* What are administrative keys and what's the difference between the "repository key" and the "root key"?
* docs.docker.com [Content trust in Docker](https://docs.docker.com/engine/security/trust/)
  * [Docker Content Trust Keys](https://docs.docker.com/engine/security/trust/#docker-content-trust-keys):
    "Trust for an image tag is managed through the use of signing keys. A key
    set is created when an operation using DCT is first invoked. A key set
    consists of the following classes of keys: (1) an offline key that is the
    root of DCT for an image tag, (2) repository or tagging keys that sign
    tags, and (3) server-managed keys such as the timestamp key, which provides
    freshness security guarantees for your repository"
  * Gives examples of how to create they keys.
  * [Client Enforcement with Docker Content Trust](https://docs.docker.com/engine/security/trust/#client-enforcement-with-docker-content-trust)
    * "Content trust is disabled by default in the Docker Client. To enable it,
      set the `DOCKER_CONTENT_TRUST` environment variable to `1`. This prevents
      users from working with tagged images unless they contain a signature.
* Note once Docker Content Trust is enabled by setting `DOCKER_CONTENT_TRUST`
  to `1`, locally built images will fail to run unless signed.
<pre><code>$ docker run --rm -v go-dev-home:/home/sean -v /home/sean/dev/elias/docker/go-dev:/home/sean/.tmp-work-Erg0Of -w /home/sean/.tmp-work-Erg0Of -it --entrypoint /bin/bash stalexan/go-dev -c 'go version'
docker: you are not authorized to perform this operation: server returned 401.
See 'docker run --help'.
</code></pre>
* So now I'd like to sign my locally built `stalexan/go-dev:latest` image. How do I do this?
* docs.docker.com [Signing Images with Docker Content Trust](https://docs.docker.com/engine/security/trust/#signing-images-with-docker-content-trust)
* Generate a "delegation key pair", stored locally in `~/.docker/trust`
<pre><code>$ docker trust key generate sean-logan6-230327a
Generating key for sean-logan6-230327a...
Enter passphrase for new sean-logan6-230327a key with ID 8ac2f79: 
Repeat passphrase for new sean-logan6-230327a key with ID 8ac2f79: 
Successfully generated and loaded private key. Corresponding public key available: /home/sean/dev/elias/docker/go-dev/sean-logan6-230327a.pub
</code></pre>
* This created the file `sean-logan6-230327a.pub` in the current directory:
<pre><code>-----BEGIN PUBLIC KEY-----
role: sean-logan6-230327a

MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUpZ9ib/aofJ44HV3vhRv4b0Emhyo
d4zFM6ZP5R5Oa0Tx7XWsCln4NEr4CkHZk4rUeZzB6HaBQC7blR4pAaf/UQ==
-----END PUBLIC KEY-----
</code></pre>
* It put the private key in `~/.docker/trust/private/` in the file
  `8ac2f79a8c7def0b4882353a2a6f0c2a02db853a777a7611263a8f939dcdf54f.key`
* I want to sign the local image but it looks like I need to register my key with Docker Hub?
<pre><code>$ docker trust sign stalexan/go-dev:latest 
you are not authorized to perform this operation: server returned 401.
</code></pre>
* docs.docker.com [Delegations for content trust](https://docs.docker.com/engine/security/trust/trust_delegation/)
* This is more complicated than I need for now, though. I won't be uploading
  go-dev to Docker Hub and don't need to sign it.
* So I think the better approach is to not set `DOCKER_CONTENT_TRUST` globally. Just
  set it in the `build` script where the image is fetched.

# Exercism Inside go-dev Container; 2023-03-29
* exercism.org [Go Exercises](https://exercism.org/tracks/go/exercises)
* exercism.org [Working Locally](https://exercism.org/docs/using/solving-exercises/working-locally)
* github.com/exercism [cli/releases)](https://github.com/exercism/cli/releases/)
* Latest release: <https://github.com/exercism/cli/releases/download/v3.1.0/exercism-3.1.0-linux-x86_64.tar.gz>
* The tar file just has a README.md, a license, and the exercism binary.
* Copy binary to go-dev.
<pre><code>$ cd ~/Downloads/exercism
$ gdr cp exercism /home/sean/bin
$ gdr exercism

A command-line interface for the v3 redesign of Exercism.

Download exercises and submit your solutions.

Usage:
   [command]

Available Commands:
  configure    Configure the command-line client.
  download     Download an exercise.
  help         Help about any command
  open         Open an exercise on the website.
  prepare      Prepare does setup for Exercism and its tracks.
  submit       Submit your solution to an exercise.
  troubleshoot Troubleshoot does a diagnostic self-check.
  upgrade      Upgrade to the latest version of the CLI.
  version      Version outputs the version of CLI.
  workspace    Print out the path to your Exercism workspace.

Flags:
  -h, --help           help for this command
      --timeout int    override the default HTTP timeout (seconds)
      --unmask-token   will unmask the API during a request/response dump
  -v, --verbose        verbose output

Use " [command] --help" for more information about a command.
</code></pre>
* Configure:
<pre><code>$ gdr exercism configure --token xxx

You have configured the Exercism command-line client:

Config dir:                       /home/sean/.config/exercism
Token:         (-t, --token)      xxx
Workspace:     (-w, --workspace)  /home/sean/exercism
API Base URL:  (-a, --api)        https://api.exercism.io/v1


$ gdr cat /home/sean/.config/exercism/user.json
{
  "apibaseurl": "https://api.exercism.io/v1",
  "token": "xxx",
  "workspace": "/home/sean/exercism"
}
</code></pre>
* Download an exercism. I want the exercises to be downloaded to the host
  directory `/home/sean/dev/current/go/exercism/`. Inside the container they're
  downloaded to `/home/sean/exercism`. Mount the host directory
  `/home/sean/dev/current/go/exercism` to the container directory
  `/home/sean/exercism/go` inside the container.  
<pre><code>$ cd ~/dev/current/go/exercism
$ gdr exercism download --exercise=lasagna --track=go
</code></pre>
* Run the tests.
<pre><code>$ cd lasagna
$ gdr go test
</code></pre>
* Submit the solution. Here since this is an exercism operation we use the path
  inside the container.
<pre><code>$ gdr exercism submit /home/sean/exercism/go/lasagna/lasagna.go

    Your solution has been submitted successfully.
    View it at:

    https://exercism.org/tracks/go/exercises/lasagna
</code></pre>

# Reproducible Builds; 2023-04-10
* Are reproducible builds possible in Go?
* Wikipedia [Reproducible builds](https://en.wikipedia.org/wiki/Reproducible_builds)
* wiki.debian.org [Reproducible Builds](https://wiki.debian.org/ReproducibleBuilds)
* [Reproducible Builds](https://reproducible-builds.org/)

# Profiling With pprof; 2023-04-10
* Source and description: <https://github.com/google/pprof>
* Documentation: <https://github.com/google/pprof/blob/main/doc/README.md>
* Create a profile. Here `args.cpuProfile` is set to the name of the output file to create.
<pre><code>import (
    "os"
    "runtime/pprof"
)
...
// Start profiling.
if args.cpuProfile != "" {
    var file *os.File
    var err error
    if file, err = <b>os.Create(args.cpuProfile)</b>; err != nil {
        printFatalError(err.Error())
    }
    defer file.Close()
    if err = <b>pprof.StartCPUProfile(file)</b>; err != nil {
        printFatalError(err.Error())
    }
    defer <b>pprof.StopCPUProfile()</b>
}
</code></pre>
* Run app in profiling mode:
<pre><code>gomarkwiki <b>--cpuprofile=cpu.prof</b> ./tmp/iwiki-dev ./tmp/output
</code></pre>
* Generate text report:
<pre><code><b>go tool pprof -tree</b> /home/sean/go/bin/gomarkwiki cpu.prof  > tmp/perf.txt
</code></pre>
* Generate graphical report:
<pre><code><b>go tool pprof -svg -output=tmp/perf.svg</b> /home/sean/go/bin/gomarkwiki cpu.prof
</code></pre>

