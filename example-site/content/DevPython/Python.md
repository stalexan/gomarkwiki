# Python
* Wikipedia [Python](https://secure.wikimedia.org/wikipedia/en/wiki/Python_(programming_language))
* Home: [python.org](https://www.python.org/)

# Style Guides
* python.org [[Style Guide for Python Code|http://www.python.org/dev/peps/pep-0008/]] 
* python.org [[Docstring Conventions|http://www.python.org/dev/peps/pep-0257/]] 

# Documenting Python Code; Fri 2022-01-14
* realpython.com [Documenting Python Code: A Complete Guide](https://realpython.com/documenting-python-code/)
* python.org [PEP 257 -- Docstring Conventions](https://www.python.org/dev/peps/pep-0257/)
* [NumPy/SciPy docstrings](https://numpydoc.readthedocs.io/en/latest/format.html)
* [reStructuredText](https://docutils.sourceforge.io/rst.html): 
  [Quick reStructuredText](https://docutils.sourceforge.io/docs/user/rst/quickref.html)
* sphinx-doc.org [Example NumPy Style Python Docstrings](https://www.sphinx-doc.org/en/1.3.1/ext/example_numpy.html)
* numpydoc.readthedocs.io [numpydoc example.py](https://numpydoc.readthedocs.io/en/latest/example.html)

# Python Linter; Fri 2022-01-14
* I'd like to run `netstat-monitor` through a linter.
* Visual Studio Code recognizes several, [here](https://code.visualstudio.com/docs/python/linting#_specific-linters).
* Pylint seems popular. Maybe start with it.
* [Pylint](https://pylint.org/)
* Checks are based on [PEP 8 -- Style Guide for Python Code](https://www.python.org/dev/peps/pep-0008/)
* [Pylint User Manual](https://pylint.pycqa.org/en/latest/)
* Install:
<pre><code>apt-get install pylint
</code></pre>
* Actually, it's better to install pylint to your virtual environment if using one, so that way
  the import checker doesn't report false errors.
<pre><code>pip install pylint
</code></pre>
* Although, I'm still getting an import error, for this line:
<pre><code>import netaddr
</code></pre>
* The error is:
<pre><code>netstat.py:45:0: E0401: Unable to import 'netaddr' (import-error)
</code></pre>
* This is the one package dependency that has to be installed. It's installed.
* Thu 2022-01-20: I've corrected most errors but am still getting a "R0801:
  Similar lines in 2 files". It turns out this is a Pylint bug. See 
  [The duplicate-code (R0801) can't be disabled #214](https://github.com/PyCQA/pylint/issues/214).
  The "duplicate code" is import statemets in two files, and has to be "duplicate".

# Visual Studio Code; Thu 2022-03-10
* code.visualstudio.com [Python in Visual Studio Code](https://code.visualstudio.com/docs/languages/python)
* code.visualstudio.com [Visual Studio Code on Linux](https://code.visualstudio.com/docs/setup/linux)
* Get Microsoft package signing key:
<pre><code>wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
</code></pre>
* Place in `/etc/apt/trusted.gpg.d`
* Add Microsoft repo to sources, as root:
<pre><code>sh -c 'echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" > /etc/apt/sources.list.d/vscode.list'
</code></pre>
* Install VS Code:
<pre><code>apt-get install code
</code></pre>
* Install Python: [Manage Button] -> Extensions -> Python (from Microsoft)
* Install vim key bindings: [Manage Button] -> Extensions -> Vim (from vscodevim)

# Building From Source; Mon 2022-03-28
* See notes here: [thudaka]({{WIKI2}}/SysConfig/thudaka.html)

# Python Build Ends Up Working; Sun 2022-01-09
* I was able to get the Python build to work, following different instructions.
* See [Build Python, Take 2]({{WIKI2}}/SysConfig/thudaka.html#build-python-210109).
* This is on thudaka which has no Python installs from Debian itself. Stick with this
  for now, and get netstat-monitor working.
* Keep the questions about how to ensure packages come from pip versus the distro, and
  experiment with this later. Create an Ubuntu container to test netstat-monitor there too,
  and see how all this works there.

# Subscribe to Mailing List for Security Updates; Tue 2022-01-11
* I'll now need to watch for security updates, to recompile.
* python.org [Mailing Lists](https://www.python.org/community/lists/)
* Mailing list is: [Python-announce-list](https://mail.python.org/mailman3/lists/python-announce-list.python.org/)

# Cython; 2022-05
* Wikipedia [Cython](https://en.wikipedia.org/wiki/Cython): "is a programming
  language that aims to be a superset of the Python programming language,
  designed to give C-like performance with code that is written mostly in
  Python with optional additional C-inspired syntax. ¶ Cython is a compiled
  language that is typically used to generate CPython extension modules.
  Annotated Python-like code is compiled to C or C++ then automatically wrapped
  in interface code, producing extension modules that can be loaded and used by
  regular Python code using the import statement, but with significantly less
  computational overhead at run time. Cython also facilitates wrapping
  independent C or C++ code into python-importable modules." 

# CPython; 2022-05
* Wikipedia [CPython](https://en.wikipedia.org/wiki/CPython): "__Not to be
  confused with Cython.__...is the reference implementation of the Python
  programming language. Written in C and Python"
* Source code: <https://github.com/python/cpython>
  * Languages: 66% Python, 32% C
* Issue tracker: github.com/python/cpython [Issues](https://github.com/python/cpython/issues)
* realpython.com [Your Guide to the CPython Source Code](https://realpython.com/cpython-source-code-guide/)
* realpython.com [CPython Internals](https://realpython.com/products/cpython-internals-book/) (book)
* realpython.com [C for Python Programmers](https://realpython.com/c-for-python-programmers/)
