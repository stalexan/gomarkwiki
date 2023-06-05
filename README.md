# Introduction

Gomarkwiki is a command-line program that converts Markdown to HTML, providing
a fast and straightforward method for generating static websites. It serves as a 
useful tool for maintaining personal wikis and note-taking. I was using 
[ikiwiki](https://ikiwiki.info/), but wanted a faster alternative that supports 
more modern syntax, such as [CommonMark](https://en.wikipedia.org/wiki/Markdown#Standardization) and
[GFM](https://github.github.com/gfm) tables. Developed in Go with 
the [Goldmark](https://github.com/yuin/goldmark) parser, Gomarkwiki ensures exceptional
speed, complete support for CommonMark 0.30,
and support for GFM extensions such as tables.

Here's a example site generated with Gomarkwiki:
[Example](https://www.alexan.org/gomarkwiki-example/Gomarkwiki%20Example.html).
This site is a few pages from personal wikis I keep. The pages aren't meant for
public consumption, but give a good idea of what Gomarkwiki can do.
The Markdown used to create this is in the
[example-site](https://github.com/stalexan/gomarkwiki/tree/main/example-site)
directory.

# Usage

```
NAME
       gomarkwiki - Generates HTML from Markdown.

SYNOPSIS
       gomarkwiki [options] source_dir dest_dir
       gomarkwiki [options] -wikis wikis_file

DESCRIPTION
       gomarkwiki generates HTML from Markdown. Each Markdown file found in
       source_dir/content results in a corresponding HTML page generated in
       dest_dir. 

       The source_dir is processed recursively. The directory structure found
       in source_dir is mirrored in dest_dir.

       Markdown files are identified by the file extensions .md, .mdwn, and
       .markdown.

       Other files found in source_dir, that are not Markdown, are copied to
       dest_dir.

       Multiple source_dir and dest_dir pairs can be specified using the -wikis
       option, giving it the path to a CSV file that has one wiki defined per
       line, formatted as source_dir,dest_dir.

       A default CSS style sheet called source.css is placed in dest_dir.
       Styles can be overridden by creating a local.css file in
       source_dir/content. The local.css file will be copied to dest_dir, and
       override any default styles found in styles.css.

       A favicon can be placed in source_dir/content to give HTML pages
       a default icon. The filename should be favicon.ico.

       String substitions can be made using string pairs in the file
       source_dir/substition-strings.csv. Each line of substition-strings.csv
       is a comma separated pair of strings, with the string placeholder first,
       followed by a comma, and then the string substition. Then anywhere
       {{placeholder}} is found in a Markdown file, the corresponding HTML file
       have the substition instead. For example if substition-strings.csv has
       the line "FOOBAR,www.foobar.com" then any instance of {{FOOBAR}} in
       a Markdown file will result in www.foobar.com in the HTML file.

OPTIONS
       -clean
              Delete any files in dest_dir that do not have a corresponding
              file in source_dir. By default no files are deleted from dest_dir.

       -help
              Show help and exit.

       -regen
              Regenerate all HTML regardless of timestamps. By default an HTML
              file is only regenerated when the timestamp on its Markdown file
              is newer that the timestamp on the HTML file.

       -verbose
              Print all status messages.

       -version
              Print version information and exit.

       -watch
              Remain running and watch for changes to regenerate files on the fly.

       -wikis wikis_file
              Generate wikis specified in CSV file, with one wiki defined per
              line formatted as source_dir,dest_dir.
```

# Examples

To generate the HTML for the example site found in the
[example-site](https://github.com/stalexan/gomarkwiki/tree/main/example-site)
directory, assuming this repo has been cloned to ~/gomarkwiki and the HTML
should be saved to ~/wikis-html/example-site:

```
gomarkwiki ~/gomarkwiki/example-site ~/wikis-html/example-site
```

Or to generate multiple wikis, say we have the CSV file `/etc/gomarkwiki/wikis.csv` with:

```
/path/to/src/wiki1,/path/to/dest/wiki1
/path/to/src/wiki2,/path/to/dest/wiki2
/path/to/src/wiki3,/path/to/dest/wiki3
```

We can generate all three wikis in one pass, clean any files from the dest
directories that don't have corresponding files in their source directory, and
remain running to watch for changes and regenerate dest files with:

```
gomarkwiki -clean -watch -wikis /etc/gomarkwiki/wikis.csv
```

# Installation

Gomarkwiki can be installed by either building from source, or downloading
a prebuilt binary.

## From Source

Gomarkwiki is written in the Go programming language and you need at least Go
version 1.20. To build gomarkwiki from source, execute the following steps:

```
$ git clone https://github.com/stalexan/gomarkwiki
$ cd gomarkwiki
$ make build
```

This builds gomarkwiki and places the executable in the `./build` directory.

## Binaries

You can download the latest stable release versions of gomarkwiki from the
[gomarkwiki releases page](https://github.com/stalexan/gomarkwiki/releases/latest).
These builds are considered stable and releases are made regularly in
a controlled manner.

There’s both pre-compiled binaries for different platforms as well as the
source code available for download.  Just download and run the one matching
your system.

If you like, you can verify the integrity of your downloads by testing the
SHA-256 checksums listed in SHA256SUMS, and verifying the integrity of the file
SHA256SUMS with the PGP signature in SHA256SUMS.asc. The PGP signature was
created using the key ([0x26565B27732B7C75](https://www.alexan.org/SeanAlexandre.asc)):

```
pub   rsa3072 2023-04-29 [SC] [expires: 2025-04-28]
      A6951D3EEB4DDAF71434364E26565B27732B7C75
uid           Sean Alexandre <sean@alexan.org>
sub   rsa3072 2023-04-29 [S] [expires: 2024-04-28]
      AAAB32D28EB8110409B4B33CD856897AA7E38BD1
```

# Reproducible Builds

The binaries released with each gomarkwiki version are
[reproducible](https://reproducible-builds.org/), which means you can reproduce
a byte identical version from the source code for that release.

Reproducible builds can be done with either Docker using the release build
scripts found in the [release-builder](https://github.com/stalexan/gomarkwiki/tree/main/example-site)
directory, or without Docker by manually doing what the scripts do.

In either case, the first step is to determine which version of Go was used to
build gomarkwiki, and which version of gomarkwiki to build. This can be done
with the `--version` option:

```
$ gomarkwiki --version
gomarkwiki v0.2.0 compiled with go1.20.5 on linux/amd64
```

## With Docker

To do a reproducible build with Docker, we first create a Docker image by
running `build-image.sh` in the
[release-builder](https://github.com/stalexan/gomarkwiki/tree/main/release-builder)
directory, giving it the version of Go that will be used to do the build. Here
we create an image that will have Go version 1.20.5:

```
./build-image.sh 1.20.5
```

Then to create the release build run `build-release.sh`, in the same directory,
giving it the location of source to build, where to place the binaries that are
created, and which commit of gomarkwiki to build. Here we build gomarkwiki
version `v0.2.0` using the source that's in `~/gomarkwiki`, and place the
binaries that are created in `~/tmp/build-output`:

```
./build-release.sh ~/gomarkwiki ~/tmp/build-output v0.2.0
```

## Without Docker

A reproducible build can also be done manually, without Docker. Here we perform
the same steps as done in the build scripts, but for just one executable. 

First install the version of Go that was used to build Gomarkwiki. For example,
to install Go 1.20.5:

```
$ GO_URL="https://dl.google.com/go/go1.20.5.linux-amd64.tar.gz"
$ cd /tmp
$ wget -O go.tar.gz.asc "${GO_URL}.asc"
$ wget -O go.tar.gz "$GO_URL" --progress=dot:giga
$ gpg --verify go.tar.gz.asc go.tar.gz
$ tar -C /usr/local -xzf go.tar.gz; \
```

If needed, the key signing key can be installed with:

```
gpg --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796'
```

More on this signing key can be found here: google.com [Linux Software Repositories](https://www.google.com/linuxrepositories/).

Add Go to `PATH` and set `GOPATH`:

```
$ export PATH=$PATH:/usr/local/go/bin
$ export GOPATH=$HOME/go
```

Extract the source to build. Here we extract to `/tmp/build`, but any directory will do:

```
$ mkdir /tmp/build
$ cd /tmp/build
$ curl -L https://github.com/stalexan/gomarkwiki/releases/download/v0.2.0/gomarkwiki-v0.2.0.tar.gz | tar xz --strip-components=1
```

Build Gomarkwiki:

```
$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X 'main.version=v0.2.0'" -o gomarkwiki_v0.2.0_linux_amd64 ./cmd/main.go
```

Create the zipped version:

```
$ bzip2 gomarkwiki_v0.2.0_linux_amd64
```

Compare the SHA256 sums. Here we've downloaded the release's SHA256SUMS file to
the current directory:

```
$ sha256sum --ignore-missing --check SHA256SUMS
gomarkwiki_v0.2.0_linux_amd64.bz2: OK
```

The SHA256 sums are the same, and we've done a reproducible build.

# License

Gomarkwiki is licensed under the [MIT License](https://spdx.org/licenses/MIT.html).
You can find the complete text in
[LICENSE.txt](https://github.com/stalexan/gomarkwiki/blob/main/LICENSE.txt).
