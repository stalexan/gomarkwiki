# Python Lanaguage; 2022-03
* docs.python.org [The Python Language Reference](https://docs.python.org/3/reference/index.html)

# Strings
* docs.python.org [Text Sequence Type — str](https://docs.python.org/3/library/stdtypes.html#text-sequence-type-str)
* docs.python.org [Format String Syntax](https://docs.python.org/3/library/string.html#format-string-syntax)
* [A Guide to the Newer Python String Format Techniques](https://realpython.com/python-formatted-output/)
* [The <format_spec\> Component](https://realpython.com/python-formatted-output/#the-format_spec-component)
* Given this:
<pre><code>amt: float = 1234567.890123

print(f"Amount take 1: {amt}")

# 20 characters side, comma separator, zero places after decimal.
print(f"Amount take 2: ${amt:20,.0f}")
</code></pre>
* Output is:
<pre><code>Amount take 1: 1234567.890123
Amount take 2: $           1,234,568
</code></pre>
* towardsdatascience.com [Python f-strings Are More Powerful Than You Might Think](https://towardsdatascience.com/python-f-strings-are-more-powerful-than-you-might-think-8271d3efbd7d) (2022-04)
* Formatting floats as ints with leading 0s:
<pre><code># Print minutes and seconds as "01:02"
elapsed_time = datetime.datetime.now() - start_time
minutes, seconds = divmod(elapsed_time.total_seconds(), 60)
print(f"{minutes:02.0f}:{seconds:02.0f}")
</code></pre>

# Enumerations
* [enum — Support for enumerations](https://docs.python.org/3/library/enum.html)
* [Enum HOWTO](https://docs.python.org/3/howto/enum.html#enum-howto)

# Iterating
* [enumerate() method](https://docs.python.org/3/library/functions.html#enumerate): To create tuples with a count.
<pre><code>&gt;&gt;&gt; seasons = ['Spring', 'Summer', 'Fall', 'Winter']
&gt;&gt;&gt; list(<b>enumerate(seasons)</b>)
[(0, 'Spring'), (1, 'Summer'), (2, 'Fall'), (3, 'Winter')]
</code></pre>
* [zip() method](https://docs.python.org/3/library/functions.html#zip): To combine iterables as tuples.
<pre><code>&gt;&gt;&gt; for item in <b>zip</b>([1, 2, 3], ['sugar', 'spice', 'everything nice']):
...    print(item)
...
(1, 'sugar')
(2, 'spice')
(3, 'everything nice')
</code></pre>

# Lists
* docs.python.org [Built-in Types: Sequence Types: list](https://docs.python.org/3.8/library/stdtypes.html#list)
* Length:
<pre><code>len(list)
</code></pre>
* docs.python.org [List Comprehensions](https://docs.python.org/3/tutorial/datastructures.html#list-comprehensions)
<pre><code># Both are equivalent
squares = list(map(lambda x: x&ast;&ast;2, range(10)))
squares = [x&ast;&ast;2 for x in range(10)]
</code></pre>
* `if` can be used to filter, and comprehensions can be nested:
<pre><code>[(x, y) for x in [1,2,3] for y in [3,1,4] if x != y]
</code></pre>
* Convert list to a [set](https://docs.python.org/3.8/library/stdtypes.html#set-types-set-frozenset) to 
  remove duplicates. Then convert it back to a list if it needs to be used as a list:
<pre><code>mylist = ['nowplaying', 'PBS', 'PBS', 'nowplaying', 'job', 'debate', 'thenandnow']
myset = set(mylist)
mynewlist = list(myset)
</code></pre>

# Dictionaries
* [Dictionaries in Python](https://realpython.com/python-dicts/)
<pre><code>teams = {
    'Colorado' : 'Rockies',
    'Boston'   : 'Red Sox',
    'Minnesota': 'Twins',
}
print(f"teams: {teams}") # teams: {'Colorado': 'Rockies', 'Boston': 'Red Sox', 'Minnesota': 'Twins'}
print(f"teams.<b>keys()</b>: {teams.keys()}") # teams.keys(): dict_keys(['Colorado', 'Boston', 'Minnesota'])
print(f"teams.<b>values()</b>: {teams.values()}") # teams.values(): dict_values(['Rockies', 'Red Sox', 'Twins'])
print(f"teams.<b>items()</b>: {teams.items()}") # teams.items(): dict_items([('Colorado', 'Rockies'), ('Boston', 'Red Sox'), ('Minnesota', 'Twins')])
</code></pre>
* realpython.com [Python args and kwargs: Demystified](https://realpython.com/python-kwargs-and-args/)
* python-reference.readthedocs.io [\*\* Dictionary Unpacking](https://python-reference.readthedocs.io/en/latest/docs/operators/dict_unpack.html)
* Dictionary unpacking can be used for function calls.
<pre><code>def add(oper1: int = 0, oper2: int = 0) -> int:
    """ Try dictionary unpacking for function calls """
    return oper1 + oper2

params_dict = {'oper1': 2, 'oper2': 3}
add(&ast;&ast;params_dict) # 5
</code></pre>
* Dictionary unpacking can also be used to concatenate dictionaries.
<pre><code>my_first_dict = {"A": 1, "B": 2}
my_second_dict = {"C": 3, "D": 4}
my_merged_dict = {&ast;&ast;my_first_dict, &ast;&ast;my_second_dict}

print(my_merged_dict) # {'A': 1, 'B': 2, 'C': 3, 'D': 4}
</code></pre>


# Variable Scope
* `if` statements do not define a new scope. For example, this works, and `x` is seen:
<pre><code>def myfunc():
    foo = True
    if foo:
        x = 200
    print(x)
</code></pre>

# Static Method Versus Class Method
* [[What is the difference between @staticmethod and @classmethod in Python?|http://stackoverflow.com/questions/136097/what-is-the-difference-between-staticmethod-and-classmethod-in-python]] 
* Static methods are similar to other languages, and are just methods that use the class name as part of their namespace.
  Nothing extra is passed to them.
* Class methods are similar to regular methods, except the class itself is passed to the method instead
  of an instance of the class.
* Class methods can be useful in cases where normally you'd want more than one
  contructor. Python only allows one definition of `__init()` per class. See
  stackoverflow.com [Is it not possible to define multiple constructors in Python?](https://stackoverflow.com/questions/2164258/is-it-not-possible-to-define-multiple-constructors-in-python). Example:
<pre><code>class C(object):

    def __init__(self, fd):
        # Assume fd is a file-like object.
        self.fd = fd

    @classmethod
    def fromfilename(cls, name):
        return cls(open(name, 'rb'))

# Now you can do:
c = C(fd)
# or:
c = C.fromfilename('a filename')
</code></pre>

# Change reference passed to function; 2013-10-18
* Is it possible to pass a reference to an object to a function and have the reference
  itself changed in the function, like is possible in C# by using the ref keyword on 
  a function parameter?
* stackoverflow.com [Python: How do I pass a variable by reference?](http://stackoverflow.com/questions/986006/python-how-do-i-pass-a-variable-by-reference): TL;DR No, all parameters are pass by value.

# Private methods
* Prefix method name with two underscores to make it private. See
  [3.9. Private functions](http://www.faqs.org/docs/diveintopython/fileinfo_private.html)

# Python Modules and Packages; 2015-08-11 <span id=packages.2015-08-11 />
* A module's name is its file name without the extension; e.g. `foo.py` is the module `foo`.
* Other file types of code files are possible; e.g. a byte code file `foo.pyc`, 
  a compiled extension module `foo.so`, etc.
* `import foo` does 3 things, for the _first import only_:
    * Find the file `foo.py`, using the module search path.
    * Compile `foo.py` to byte code.
    * Run the module.
* `reload foo` will redo these 3 things.
* The __module search path__ is made up of:
    * The __home directory__.
        * For a program, this is the directory that contains the top-level script.
        * For an interactive session, this is the current working directory.
    * `PYTHONPATH` environment variable, if set.
    * The __standard library directories__.
    * The content of any `.pth` files. The directory for these files is site/install
      specific.
* `sys.path` is a list of all the directories that make up the module search path.
  It can be configured at runtime to futher customize the module search path.
* A module __package__ is a special type of module that references a directory
  that can contain other modules (source files and/or other directories.)
  Usually it's just referred to as a package, although internally both
  a regular source code module and directory module are both represented by the
  `module` class. (Also, `import foo` could refer to either file `foo.py` or
  directory `foo/`.) Each package directory must have a `__init__.py` file, which
  is run when the module/package is loaded.
* Example directory structure:
<pre><code>    dir0\
        dir1\
            __init__.py
            dir2\
            __init__.py
            mod.py
</code></pre>
* If `dir0` is in the module search path, then `import dir1.dir2.mod` will run/load:
    * `dir0\dir1\__init__.py`
    * `dir0\dir1\dir2\__init__.py`
    * `dir0\dir1\dir2\mod.py`
* __Relative imports__ search current package. Examples assuming code located in module `A.B.C`:
<pre><code>from . import D     # Imports A.B.D
from .. import E    # Imports A.E 
from .D import X    # Imports A.B.D.X
from ..E import X   # Imports A.E.X
</code></pre>

# Type Hinting; 2022-01
* mypy.readthedocs.io [Type hints cheat sheet](https://mypy.readthedocs.io/en/stable/cheat_sheet_py3.html)
* docs.python.org [typing — Support for type hints](https://docs.python.org/3/library/typing.html)
* python.org [PEP 483 -- The Theory of Type Hints](https://www.python.org/dev/peps/pep-0483/)
* python.org [PEP 484 -- Type Hints](https://www.python.org/dev/peps/pep-0484/)
* python.org [PEP 526 -- Syntax for Variable Annotations](https://www.python.org/dev/peps/pep-0526/)
* lwn.net [Type hinting for Python](https://lwn.net/Articles/627418/)
* [mypy](http://mypy-lang.org/index.html)
* mypy.readthedocs.io [mypy documentation](https://mypy.readthedocs.io/en/latest/index.html)
* mypy.readthedocs.io [Type hints cheat sheet](https://mypy.readthedocs.io/en/stable/cheat_sheet_py3.html)
* stackoverflow.com [Does Python evaluate type hinting of a forward reference?](https://stackoverflow.com/questions/55320236/does-python-evaluate-type-hinting-of-a-forward-reference) (2019): 
  Type hint notations can cause forward reference problems, where the Python
  interpreter doesn't see a type because it hasn't been fully defined yet (e.g.
  returning an instance of the type from a method of the type). One
  workaround for now is to use quotes around the return type; e.g. 
  `->"SocketInfo"` instead of `-> SocketInfo`.
  * Thu 2022-01-27: Although, this appears to cause problems with callers from
    other modules, when those modules use type hinting as well. The test module
    for both of the `-> "SocketInfo"` methods in the SocketInfo class get odd signature
    mismatch errors. The errors say the calling parameters aren't correct, but they are.
* The content of a method is not checked until a return type for the method is defined. 
* Install on Debian:
<pre><code>apt-get install mypy
</code></pre>
* To install all available type packages for current environment:
<pre><code>mypy --install-types
</code></pre>

# Pylance; 2022-04-02
* A form of type hinting is provided by Pylance in VS Code. This seems espeically useful for
  Jupyter notebooks which don't have mypy support.
* visualstudio.com [Pylance](https://marketplace.visualstudio.com/items?itemName=ms-python.vscode-pylance):

  "is an extension that works alongside Python in Visual Studio Code...Pylance
  is powered by Pyright, Microsoft's static type checking tool. Using Pyright,
  Pylance has the ability to supercharge your Python IntelliSense experience
  with rich type information"

# Conditional Expressions; 2022-05-03
* docs.python.org [Conditional expressions](https://docs.python.org/3/reference/expressions.html#conditional-expressions)
<pre><code>x if C else y
</code></pre>

