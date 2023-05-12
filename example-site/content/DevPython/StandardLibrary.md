# Python Standard Library; 2022-03
* [The Python Standard Library](https://docs.python.org/3/library/index.html)
* docs.python.org [Exception hierarchy](https://docs.python.org/3/library/exceptions.html#exception-hierarchy)

# File and Stream I/O; 2022-10
* docs.python.org [File and Directory Access](https://docs.python.org/3/library/filesys.html)
* docs.python.org [Glossary: file object](https://docs.python.org/3/glossary.html#term-file-object):
  File objects are created with `open()`. Interfaces for the different types of file objects are defined in the `io` module.
* docs.python.org [Built-in Functions: open()](https://docs.python.org/3/library/functions.html#open)
* Use [os.open()](https://docs.python.org/3/library/os.html#os.open) instead of
  `open()` to get an `fd` instead of a `TextIO`. Use `try/finally` instead of
  `with`, though, since `os.open()` just returns an int.

<pre><code>fd = os.open(PATH, os.O_CREAT | os.O_WRONLY)
try:
    os.write(fd, "foo")
finally:
    os.close(fd)
</code></pre>
* docs.python.org [io: TextIOWrapper](https://docs.python.org/3/library/io.html#io.TextIOWrapper): Returned by `open()` for text files.
* docs.python.org [io — Core tools for working with streams](https://docs.python.org/3/library/io.html#module-io)
* docs.python.org [7.2. Reading and Writing Files](https://docs.python.org/3/tutorial/inputoutput.html#tut-files)

# Dates; 2022-04
* docs.python.org [datetime — Basic date and time types](https://docs.python.org/3/library/datetime.html)
* [strftime() and strptime() Behavior](https://docs.python.org/3/library/datetime.html#strftime-strptime-behavior): 
  Formatting dates as strings.
* docs.python.org [Format Codes](https://docs.python.org/3/library/datetime.html#strftime-and-strptime-format-codes)

# Regular Expressions; 2022-04
* docs.python.org [Regular Expression HOWTO](https://docs.python.org/3/howto/regex.html)

# Sorting; 2022-05
* docs.python.org [Sorting HOW TO](https://docs.python.org/3/howto/sorting.html)

# Logging; 2022-05
* docs.python.org [Logging HOWTO](https://docs.python.org/3/howto/logging.html)
* docs.python.org [logging — Logging facility for Python](https://docs.python.org/3/library/logging.html)
* docs.python.org [Logging Cookbook](https://docs.python.org/3/howto/logging-cookbook.html)

# Configuration File Parser; 2022-05
* docs.python.org [configparser — Configuration file parser](https://docs.python.org/3/library/configparser.html)

# Data Classes; 2022-05
* docs.python.org [dataclasses — Data Classes](https://docs.python.org/3/library/dataclasses.html)

# Argument Parser; 2022-05
* docs.python.org [argparse — Parser for command-line options, arguments and sub-commands](https://docs.python.org/3/library/argparse.html)
* docs.python.org [Argparse Tutorial](https://docs.python.org/3/howto/argparse.html)

