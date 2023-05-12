# Asyncio and Subprocess Management; 2022-05
* Notes for asyncio and subprocess management are grouped together because both
  have support for subprocess management.

# Process Management; 2022-04 <span id=proc-manangement-220401 />
* docs.python.org [subprocess — Subprocess management](https://docs.python.org/3/library/subprocess.html)
* "The recommended approach to invoking subprocesses is to use the
  [run()](https://docs.python.org/3/library/subprocess.html#subprocess.run)
  function for all use cases it can handle. For more advanced use cases, the
  underlying [Popen](https://docs.python.org/3/library/subprocess.html#subprocess.Popen)
  interface can be used directly."
* [Popen Objects](https://docs.python.org/3/library/subprocess.html#popen-objects)
  * [wait()](https://docs.python.org/3/library/subprocess.html#subprocess.Popen.wait)
  * [send_signal()](https://docs.python.org/3/library/subprocess.html#subprocess.Popen.send_signal)

# asyncio; 2022-05
* docs.python.org [asyncio — Asynchronous I/O](https://docs.python.org/3/library/asyncio.html)
* docs.python.org [Coroutines and Tasks](https://docs.python.org/3/library/asyncio-task.html)
  * [Awaitables](https://docs.python.org/3/library/asyncio-task.html#awaitables):
    "There are three main types of __awaitable objects__: coroutines, Tasks, and Futures."
    * "__coroutines__ are awaitables and therefore can be awaited from other coroutines"
    * "__Tasks__ are used to schedule coroutines concurrently.  When a coroutine is
      wrapped into a Task with functions like
      [asyncio.create_task()](https://docs.python.org/3/library/asyncio-task.html#asyncio.create_task)
      the coroutine is automatically scheduled to run soon"
    * "A __Future__ is a special low-level awaitable object...Future objects in
      asyncio are needed to allow callback-based code to be used with
      async/await."
  * [Task Object](https://docs.python.org/3/library/asyncio-task.html#task-object)
    * [add_done_callback()](https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.add_done_callback)
* docs.python.org [Futures](https://docs.python.org/3/library/asyncio-future.html)
* docs.python.org [Synchronization Primitives](https://docs.python.org/3/library/asyncio-sync.html)
  * [Lock](https://docs.python.org/3/library/asyncio-sync.html#asyncio.Lock)
  * [Event](https://docs.python.org/3/library/asyncio-sync.html#asyncio.Event)
* docs.python.org [Subprocesses](https://docs.python.org/3/library/asyncio-subprocess.html):
  Similar to [Process Management](#proc-manangement-220401) but for async code.
  * [asyncio.create_subprocess_exec()](https://docs.python.org/3/library/asyncio-subprocess.html#asyncio.create_subprocess_exec): 
    Creates a subprocesses, represented by a 
    [Process](https://docs.python.org/3/library/asyncio-subprocess.html#asyncio.asyncio.subprocess.Process)
    instance.
* class asyncio.subprocess.[Process](https://docs.python.org/3/library/asyncio-subprocess.html#asyncio.asyncio.subprocess.Process)
  * [wait()](https://docs.python.org/3/library/asyncio-subprocess.html#asyncio.asyncio.subprocess.Process.wait)
  * [send_signa()](https://docs.python.org/3/library/asyncio-subprocess.html#asyncio.asyncio.subprocess.Process.send_signal)

# Misc Articles
* realpython.com [Async IO in Python: A Complete Walkthrough](https://realpython.com/async-io-python/)
