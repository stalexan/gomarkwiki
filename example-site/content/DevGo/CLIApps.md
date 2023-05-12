# Building CLI Apps in Go
* go.dev [Command-line Interfaces (CLIs)](https://go.dev/solutions/clis):
  * "__Cobra__ is both a library for creating powerful modern CLI applications and
    a program to generate applications and CLI applications in Go. Cobra powers
    most of the popular Go applications including CoreOS, Delve, Docker,
    Dropbox, Git Lfs, Hugo, Kubernetes, and many more. With integrated command
    help, autocomplete and documentation “[it] makes documenting each command
    really simple”"
  * "__Viper__ is a complete configuration solution for Go applications,
    designed to work within an app to handle configuration needs and formats.
    Cobra and Viper are designed to work together."

# Cobra
* pkg.go.dev [cobra](https://pkg.go.dev/github.com/spf13/cobra)
* Code: github.com/spf13 [cobra](https://github.com/spf13/cobra)
* Website: [cobra.dev](https://cobra.dev/)
* github.com/spf13/cobra-cli [README.md](https://github.com/spf13/cobra-cli/blob/main/README.md)
* github.com/spf13/cobra [User Guide](https://github.com/spf13/cobra/blob/main/user_guide.md)

# Viper
* pkg.go.dev [viper](https://pkg.go.dev/github.com/spf13/viper)
* Code: github.com/spf13 [viper](https://github.com/spf13/viper)
* Is a configuration manager that works well with Cobra.
* Looks at environment variables as well as config file; e.g. the vaule of the
  environment variable `DATAFILE` will be used for
  `viper.GetString("datafile")`
* Set application specific prefix for environment variables. For example with this the
  environment variable would now need to be `TRI_DATAFILE`:
<pre><code>viper.SetEnvPrefix("tri")
</code></pre>
* Or, to set a config variable in a config file:
<pre><code>echo "datafile: /home/sean/.tridos" > /home/sean/.tri.yaml
</code></pre>

# spf13.com [Building an Awesome CLI App in Go – OSCON 2017](https://spf13.com/presentation/building-an-awesome-cli-app-in-go-oscon/)
* Install Cobra module:
<pre><code>go get -u github.com/spf13/cobra@latest
</code></pre>
* Install Cobra CLI:
<pre><code>go install github.com/spf13/cobra-cli@latest
</code></pre>
* Create Go module:
<pre><code>$ mkdir ~/dev/current/go/cobra-tri-tutorial
$ cd !$
$ gdr go mod init stalexan/tri
</code></pre>
* Initialize application:
<pre><code>gdr 'cobra-cli init --author "Sean Alexandre sean@alexan.org"'
</code></pre>
* Build and run:
<pre><code>gdr go install
gdr tri
</code></pre>
* Add the add command, and run it:
<pre><code>gdr cobra-cli add add
gdr go install
gdr tri add
</code></pre>
* Add the list command:
<pre><code>gdr cobra-cli add list
</code></pre>
