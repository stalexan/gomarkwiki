# Docker
* Wikipedia [Docker](https://en.wikipedia.org/wiki/Docker_\(software))
* Website: [Docker](https://www.docker.com/)
* [Docker Hub](https://hub.docker.com/)

# Installing
* [Install Docker Engine on Debian](https://docs.docker.com/engine/install/debian/)
  * See [Debian Install on logan5](#install-220415)
* docs.docker.com [Install Docker Compose](https://docs.docker.com/compose/install/)
  * See [install for Docker connect tutorial](#docker-compose-220530)

# Documentation, Docker:
* docs.docker.com [Dockerfile reference](https://docs.docker.com/engine/reference/builder/)
* docs.docker.com [docker build](https://docs.docker.com/engine/reference/commandline/build/)
* docs.docker.com [docker run](https://docs.docker.com/engine/reference/commandline/run/)

# Documentation, Compose:
* docs.docker.com [Compose specification](https://docs.docker.com/compose/compose-file/)
  * Format for `ports` is `[host]:[container]`; e.g. `3001:3000` surfaces port
    `3000` in container as port `3001` on host.
* docs.docker.com [docker compose](https://docs.docker.com/engine/reference/commandline/compose/)
* docker.com/get-started [Use Docker Compose](https://docs.docker.com/get-started/08_using_compose/)
  * Place `command` entries in compose file to specify what command will run within container.
  * The `environment` keyword can be used to specify environment variables.
  * Logs can be monitored across all containers with `docker compose --file docker-compose-dev.yml logs -f`

# Docker Commands: Images
* Pull an image. By default the image tagged `latest` is pulled:
<pre><code>docker pull busybox
</code></pre>
* Delete an image:
<pre><code>docker rmi [image-id]
</code></pre>
* List images:
<pre><code>docker images
</code></pre>
* Delete all images
<pre><code>docker rmi -f $(docker images -aq)
</code></pre>
* Delete all images, containers, and volumes not used by running containers:
<pre><code>docker system prune
</code></pre>

# Docker Commands: Containers
* Run a container, binding port 8080 of the container to port 80 of host (host port comes first). This creates a container:
<pre><code>docker run -p 127.0.0.1:80:8080/tcp ubuntu bash
</code></pre>
* Stop a container
<pre><code>docker stop nginx-container
</code></pre>
* Restart a stopped container. This assumes `--rm` wasn't used when the container was started.
<pre><code>docker container start nginx-container
</code></pre>
* Restart a running container:
<pre><code>docker restart nginx-container
</code></pre>
* Delete all containers that have exited:
<pre><code>docker container prune
</code></pre>
* Delete all containers, including volumes:
<pre><code>docker rm -vf $(docker ps -aq)
</code></pre>
* List running containers:
<pre><code>docker ps
</code></pre>
* List all containers:
<pre><code>docker ps -a
</code></pre>
* See all details for a container:
<pre><code>docker inspect [container-name]
</code></pre>
* See port mappings for a container:
<pre><code>docker port [container-name]
</code></pre>
* Stop a detached container:
<pre><code>docker stop [container-name]
</code></pre>
* Kill a detached container, with no wait time for services to be brought down gracefully:
<pre><code>docker kill [container-name]
</code></pre>
* Run a command in a running container:
<pre><code>docker exec -it hello-world-node-container ping database
</code></pre>
* Copy a file to a container:
<pre><code>docker cp backup.sql pg_container:/
</code></pre>
* Start interactive bash shell in container:
<pre><code>docker exec -it [container-name] bash
</code></pre>
* View logs for a container:
<pre><code>docker logs [container-name]
</code></pre>

# Docker Commands: Volumes
* Create a volume:
<pre><code>docker volume create go-bin
</code></pre>
* Use the volume in a call to `docker run`:
<pre><code>docker run -it --rm -v my-go-bin:/go/bin golang
</code></pre>
* List all volumes:
<pre><code>docker volume ls
</code></pre>
* Delete volume:
<pre><code>docker volume rm [volume]
</code></pre>
* Delete all volumes not used by any containers:
<pre><code>docker volume prune
</code></pre>

# Docker Commands: Networks
* List all networks:
<pre><code>docker network ls
</code></pre>

# Docker Compose Commands
* Show version:
<pre><code>docker compose version
</code></pre>
* Bring containers up in detached mode, which will cause them to be removed with `docker compose down` later:
<pre><code>docker compose up --detach
</code></pre>
* Bring containers down, and remove them:
<pre><code>docker compose [--file compose.yml] down
</code></pre>
* Run just one container, in detached mode, and delete it on exit:
<pre><code>docker compose run --rm --detach [service-name]
</code></pre>
* Rebuild images:
<pre><code>docker compose build --no-cache --pull
</code></pre>
* List all volumes:
<pre><code>docker volume ls
</code></pre>
* Display where a volume is stored on host, using a Go template:
<pre><code>docker volume inspect --format '{{ .Mountpoint }}' [volume-name]
</code></pre>
* Delete a volume:
<pre><code>docker volume rm 
</code></pre>

# Docker Compose with Anchors and Overrides; 2022-10
* [YAML Anchors, Aliases, and Overrides](https://www.linode.com/docs/guides/yaml-anchors-aliases-overrides-extensions/)

# Security
* github.com/veggiemonk [awesome-docker](https://github.com/veggiemonk/awesome-docker):
  [Security](https://github.com/veggiemonk/awesome-docker#security-1)
* github.com/AonCyberLabs [Docker Secure Deployment Guidelines](https://github.com/AonCyberLabs/Docker-Secure-Deployment-Guidelines) (2022)
  * "do not run untrusted applications with root privileges within containers"

# Intro to Docker; TriLUG; 2014-03-13
* Vincent Batts; vbatts@redhat.com
* Developer concern: all servers look the same.
* Ops conern: all containers look the same.
* Terms
    * __Image__: snapshot that you're starting from; are read-only.
    * __Layer__: an image has multiple layers
        * Layers stack on top of each other to create new images.
        * Each commit creates a layer.
        * Base image has no parent.
    * __Container__: a running image; usually just one process; commit creates new image;
      isolated from host system
* Tar up file system to create first layer.
* Based on:
    * cgroups
    * Namespaces: process have their own PIDs in a container, but can be seen with;
      their other PID outside the container; same with network, filesystem
    * AuFS/DeviceMapper/Btrfs: CoW snapshoting of filesystem images.
* v1 not out yet; lots of change.
* `--privileged` flag can be used to limit what a container can do to host; lots
  of work still happening with this, though.
* Images don't have a kernel of their own.
* Can vnc into a docker container; is a work in progress.
* Networking: NAT'd through.
* Dockerfile: scripts to run images.
* OpenShift is looking to integrate this. 
* OpenStack is as well.
* OS support: any Linux 64-bits.
* Update model is still being worked out. So if a layer up the chain gets updated,
  the layers that descend from that may or may not work.
* Is written in Go Lang.

# Docker Security; 2014
* [Are Docker containers really secure?](https://opensource.com/business/14/7/docker-security-selinux) (2014-07):
  Treat privileged processes in a container the same as a privileged process outside the container; 
  e.g. don't run misc containers. Docker containers don't protect against malware.
* [Bringing new security features to Docker](https://opensource.com/business/14/9/security-for-docker) (2014-09):
  What's being done to keep a container from breaking out into host system.
* serverfault.com [How to handle security updates within Docker containers?](https://serverfault.com/questions/611082/how-to-handle-security-updates-within-docker-containers) (2014-07): Require updating the base image, and then rebuilding (reinstalling, etc) the application image. To help simplify this, don't store state in an image.
* [Lets review.. Docker (again)](http://iops.io/blog/docker-hype/) (2015-01)
* 2017-07-10: How's Docker security now, though?
* docker.com [Docker security](https://docs.docker.com/engine/security/security/)
* It looks like Docker security is a top priority now.
* From Wikipedia [Docker](https://en.wikipedia.org/wiki/Docker_\(software)):
  "In April 2014, it was revealed that the __C.I.A.__'s investment arm In-Q-Tel was
  a large investor in Docker. __Docker has yet to respond__ to any questions
  about the nature of the investment and if any adulteration to the final
  product was requested by the spy agency and if so, what requests were met."
* 2017-07-10 <span id=security-concerns.2017-07-10 />: One concern I still have, though, 
  after reading the [Get Started](#get-started.2017-07) 
  guide and [Samples](#samples.2017-07) is that __the model that's really encouraged or
  promoted is of just using whatever public images people have published. Trust is wide
  open and seems the mode is ripe for easy backdoor insertion.__
* That said, there's the new [Enterprise Edition](#enterprise-edition.2017-07). One
  of the main things it provides is "certified" images. So it seems they're moving to 
  income generation model where the service they provide, at a cost, is making
  more secure/trusted images available.

# Enterprise Edition; 2017-07 <span id=enterprise-edition.2017-07 />
* There's now an "Enterprise Edition". Is it no longer open source?
* Hacker News discussion [here](https://news.ycombinator.com/item?id=13774136) (2017)
* [Docker Enterprise Edition: Is it ready for the enterprise … and is the enterprise ready for it?](http://techgenix.com/docker-enterprise-edition/) (2017-05)
* Looks like this is more like what Red Hat has done with Fedora. The empahsis on the
  paid model, at least for now, is on certified images and support.

# Learning Docker; 2017-07 <span id=learning.2017-07 />
* docker.com [Get Started](https://docs.docker.com/get-started/) <span id=get-started.2017-07 />:
  Much of the emphasis is on spinning up many instances of an image, related instances, and doing
  this on multiple machines. __One of the main uses cases is large scale application deployment.__
* docker.com [User Guide](https://docs.docker.com/engine/userguide/)
* docker.com [Overview of Docker Compose](https://docs.docker.com/compose/overview/)
* docker.com [Samples](https://docs.docker.com/samples/) <span id=samples.2017-07 />:
  __Seemingly sketchy list of all different kinds of images to use;__; see 
* docker.com [Admin Guide](https://docs.docker.com/engine/admin/)

# Misc Articles; 2014
* Joey Hess [what does docker.io run -it debian sh run? ](http://joeyh.name/blog/entry/docker_run_debian/) (2014-06-19): 
  Build your own Debian Docker images, and guidelines for.
* Hacker News [Docker container breakout?](https://news.ycombinator.com/item?id=7909622) (2014-06-18)
* [shocker.c](http://stealth.openwall.net/xSports/shocker.c): 
  "Demonstrates that any given docker image someone is asking you to run in
  your docker setup can access ANY file on your host, e.g. dumping hosts
  /etc/shadow or other sensitive info, compromising security of the host and
  any other docker VM's on it."

# Docker for Debian; 2018-01-13
* docs.docker.com [Get Docker CE for Debian](https://docs.docker.com/engine/installation/linux/docker-ce/debian/):
  "CE" is the "Community Edition". "Docker EE is not supported on Debian."
* docs.docker.com [Install using the repository](https://docs.docker.com/engine/installation/linux/docker-ce/debian/#install-using-the-repository)
* blog.docker.com [Online Meetup Recap: Docker Community Edition (CE) and Enterprise Edition (EE)](https://blog.docker.com/2017/03/docker-online-meetup-recap-docker-enterprise-edition-ee-community-edition-ce/) (2017-03):
  "Docker EE, supported by Docker Inc., is available on certified operating
  systems and cloud providers and runs certified Containers and Plugins from
  Docker Store. The Docker open source products are now Docker CE and we have
  adopted a new lifecycle and time-based versioning scheme for both Docker EE
  and CE."
* blog.docker.com [Announcing Docker Enterprise Edition](https://blog.docker.com/2017/03/docker-enterprise-edition/) (2017-03)
* docker.com [Docker for Debian](https://www.docker.com/docker-debian)
* docs.docker.com [Get Started, Part 1: Orientation and setup](https://docs.docker.com/get-started/)

# Docker Security 2022
* docs.docker.com [Docker security](https://docs.docker.com/engine/security/)
  * "Docker containers are very similar to LXC containers, and they have
    similar security features. When you start a container with docker run,
    behind the scenes Docker creates a set of namespaces and control groups
    for the container."
  * The Docker daemon runs as root and care should be taken to [secure it](https://docs.docker.com/engine/security/#docker-daemon-attack-surface):
    * Only trusted users should be allowed to control your Docker daemon. 
    * Secure the REST endpoints with either HTTPS or ssh over TLS.
* owasp.org [Docker Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
  * Not just the server but containers need to be kept up to date; e.g. the Dirty COW kernel exploit allowed access from container kernel to host kerne.
  * Review this article again when I'm more familiar wit Docker terms.

# Debian Install on logan5 (dev machine); Fri 2022-04-15 <span id=install-220415 />
* [Install Docker Engine on Debian](https://docs.docker.com/engine/install/debian/):
  Installs packages from docker.com.
* Install dependencies:
<pre><code>apt-get install ca-certificates curl gnupg lsb-release
</code></pre>
* Store Docker’s official GPG key to local file `/usr/share/keyrings/docker-archive-keyring.gpg`:
<pre><code>curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
</code></pre>
* Define apt source:
<pre><code>echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
</code></pre>
* This creates the file `/etc/apt/sources.list.d/docker.list` with:
<pre><code>deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian bullseye stable
</code></pre>
* Install Docker Engine:
<pre><code>apt-get update
apt-get install docker-ce docker-ce-cli containerd.io
</code></pre>
* Verify that install worked:
<pre><code>$ sudo docker run hello-world

Unable to find image 'hello-world:latest' locally
latest: Pulling from library/hello-world
2db29710123e: Pull complete 
Digest: sha256:10d7d58d5ebd2a652f4d93fdd86da8f265f5318c6a73cc5b6a9798ff6d2b2e67
Status: Downloaded newer image for hello-world:latest

Hello from Docker!
This message shows that your installation appears to be working correctly.

To generate this message, Docker took the following steps:
 1. The Docker client contacted the Docker daemon.
 2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    (amd64)
 3. The Docker daemon created a new container from that image which runs the
    executable that produces the output you are currently reading.
 4. The Docker daemon streamed that output to the Docker client, which sent it
    to your terminal.

To try something more ambitious, you can run an Ubuntu container with:
 $ docker run -it ubuntu bash

Share images, automate workflows, and more with a free Docker ID:
 https://hub.docker.com/

For more examples and ideas, visit:
 https://docs.docker.com/get-started/
</code></pre>
* docs.docker.com [Post-installation steps for Linux](https://docs.docker.com/engine/install/linux-postinstall/)
* docs.docker.com [Develop with Docker](https://docs.docker.com/develop/)

# Docker Tutorial: Image and Container Basics; Fri 2022-04-15 <span id=tutorial-220415 />
* This looks very well written, and interesting:
  docker-curriculum.com [A Docker Tutorial for Beginners](https://docker-curriculum.com/)
* The writer (Prakhar Srivastav) is a developer at Google. He looks interesting: prakhar.me [About](https://prakhar.me/about/)
* Pull busybox image, as root:
<pre><code>$ docker pull busybox
Using default tag: latest
latest: Pulling from library/busybox
50e8d59317eb: Pull complete 
Digest: sha256:d2b53584f580310186df7a2055ce3ff83cc0df6caacf1e3489bff8cf5d0af5d8
Status: Downloaded newer image for busybox:latest
docker.io/library/busybox:latest
</code></pre>
* View available images, as root:
<pre><code>$ docker images
REPOSITORY    TAG       IMAGE ID       CREATED        SIZE
busybox       latest    1a80408de790   39 hours ago   1.24MB
ubuntu        latest    825d55fb6340   9 days ago     72.8MB
hello-world   latest    feb5d9fea6a5   6 months ago   13.3kB
</code></pre>
* Run the busybox image. The images runs and exits when the command passed to it is done:
<pre><code>$ hello from busybox
</code></pre>
* See what's running (nothing in this case):
<pre><code>$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
</code></pre>
* See what's running together with a history of what's run:
<pre><code>$ docker ps -a
CONTAINER ID   IMAGE         COMMAND                  CREATED              STATUS                          PORTS     NAMES
c076e9d4e77c   busybox       "echo 'hello from bu…"   About a minute ago   Exited (0) About a minute ago             exciting_heyrovsky
b5e6318b962c   ubuntu        "bash"                   8 minutes ago        Exited (0) 7 minutes ago                  pedantic_mahavira
f053aa02cfae   hello-world   "/hello"                 9 minutes ago        Exited (0) 9 minutes ago                  quirky_blackwell
</code></pre>
* Run the busybox image interactively. Delete some things.
<pre><code>$ docker run -it busybox sh
/ # pwd
/
/ # which ls
/bin/ls
/ # ls
bin   dev   etc   home  proc  root  sys   tmp   usr   var
/ # rm -rf bin
/ # ls
sh: ls: not found
/ # exit
</code></pre>
* Rerun the images and see that everything's back:
<pre><code>$ docker run -it busybox sh
/ # ls
bin   dev   etc   home  proc  root  sys   tmp   usr   var
/ # 
</code></pre>

* It's a good idea to delete containers. They exit but remain on disk. Note,
  this is not deleting the images, just the containers that were created based
  on the images:

<pre><code># List images.
$ docker images
REPOSITORY    TAG       IMAGE ID       CREATED        SIZE
busybox       latest    1a80408de790   41 hours ago   1.24MB
ubuntu        latest    825d55fb6340   9 days ago     72.8MB
hello-world   latest    feb5d9fea6a5   6 months ago   13.3kB

# List containers.
$ docker ps -a
CONTAINER ID   IMAGE         COMMAND                  CREATED       STATUS                     PORTS     NAMES
39cbe241ab52   busybox       "sh"                     2 hours ago   Exited (0) 2 minutes ago             frosty_kepler
b3673c3c91a6   busybox       "sh"                     2 hours ago   Exited (127) 2 hours ago             modest_euler
04ac31e71f07   busybox       "bash"                   2 hours ago   Created                              confident_varahamihira
c076e9d4e77c   busybox       "echo 'hello from bu…"   2 hours ago   Exited (0) 2 hours ago               exciting_heyrovsky
b5e6318b962c   ubuntu        "bash"                   2 hours ago   Exited (0) 2 hours ago               pedantic_mahavira
f053aa02cfae   hello-world   "/hello"                 2 hours ago   Exited (0) 2 hours ago               quirky_blackwell

# Delete a specific container
$ docker rm c076e9d4e77c
c076e9d4e77c

# Delete containers whose status is exited.
$ docker rm $(docker ps -a -q -f status=exited)
39cbe241ab52
b3673c3c91a6
b5e6318b962c
f053aa02cfae

# Same effect but more succint.
docker container prune
</code></pre>
* Also, `docker run` has a `--rm` flag that will automatically remove a container when it exits.

# Docker Tutorial: Run a Static Web App; Fri 2022-04-15
* Continued from the same tutorial, here: [Web Apps with Docker](https://docker-curriculum.com/#webapps-with-docker)
* Run a static web site in a Docker container:
<pre><code>$ docker run --rm -it prakhar1989/static-site
Unable to find image 'prakhar1989/static-site:latest' locally
latest: Pulling from prakhar1989/static-site
d4bce7fd68df: Pull complete 
a3ed95caeb02: Pull complete 
573113c4751a: Pull complete 
31917632be33: Pull complete 
77e66f18af1c: Pull complete 
df3f108f3ade: Pull complete 
d7a279eb19f5: Pull complete 
e798589c05c5: Pull complete 
78eeaf458ae0: Pull complete 
Digest: sha256:bb6907c8db9ac4c6cadb25162a979e286575cd8b27727c08c7fbaf30988534db
Status: Downloaded newer image for prakhar1989/static-site:latest
Nginx is running...
</code></pre>
* This runs the site, but we don't know what port it's running on. Also, the container blocks any input to the terminal. Ctrl-C to exit.
* This time run the container in detached mode (`-d`) to not block the
  terminal, publish the ports that are being used (`-P`) and give the container
  the name `static-site`:
<pre><code># Start the container
$ docker run -d -P --name static-site prakhar1989/static-sitete
3535ed143bd27360b5ea9e7b672ea2caf9ce66e1e3d040da7aeb43eeef92f250

# See what the port mappings are:
$ docker port static-site
443/tcp -> 0.0.0.0:49153
443/tcp -> :::49153
80/tcp -> 0.0.0.0:49154
80/tcp -> :::49154
</code></pre>
* Browse to the site locally: <http://localhost:49154/>
* Ports can be specified explicitly:
<pre><code>$ docker run -p 8888:80 prakhar1989/static-site
</code></pre>
* Stop a detached container, using either the container name or ID:
<pre><code>docker stop static-site
</code></pre>

# Docker Tutorial: Run a Flask Web App; Fri 2022-04-15
* Continued from the same tutorial, [here](https://docker-curriculum.com/#docker-images).
* Wikipedia [Flask](https://en.wikipedia.org/wiki/Flask_(web_framework)):
  "is a micro web framework written in Python. It is classified as
  a microframework because it does not require particular tools or
  libraries. It has no database abstraction layer, form validation, or any
  other components where pre-existing third-party libraries provide common
  functions. However, Flask supports extensions that can add application
  features as if they were implemented in Flask itself. Extensions exist for
  object-relational mappers, form validation, upload handling, various open
  authentication technologies and several common framework related tools."
* There are two types of images, __base__ and __child__. Base images have no parent, and are usally
  OS based (e.g. ubuntu, debian, busybox, etc). Child images are built on base images to add functionality.
* Images can also be categorized based on whether they are __official__ or
  __user images__. Official images come from Docker and usually have one-word
  names such as `python`, `ubuntu`, `debian`, `hello-world`, `busybox`, etc.
  User images build on base images and are typically named with the format
  `user/image-name`.
* Clone Prakhar's repo that has the files needed to build an image:
<pre><code>$ mkdir -p ~/dev/github-prakhar1989
$ git clone https://github.com/prakhar1989/docker-curriculum.git
</code></pre>
* cd to the directory that has the files that will be used to build the image. Do this
  as root since the docker daemon will be used to build the image:
<pre><code>$ cd /home/sean/dev/github-prakhar1989/docker-curriculum/flask-app
</code></pre>
* The `Dockerfile` there defines the steps that will be used to build the image:
<pre><code>FROM python:3.8

# set a directory for the app
WORKDIR /usr/src/app

# copy all the files to the container
COPY . .

# install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# tell the port number the container should expose
EXPOSE 5000

# run the command
CMD ["python", "./app.py"]
</code></pre>
* The `FROM` line specifies the base image that will be used. In this case it's 
  a basic Debian server where Python 3.8 was built from source. See: 
  * hub.docker.com [python](https://hub.docker.com/_/python/)
  * There are tags for lots of different versions of Python built on different
    base images. I see a 3.10 Python built on Bullseye. The link for it takes
    you do the Docker file used to build the image: github.com
    docker-library/python/3.10/bullseye/[Dockerfile](https://github.com/docker-library/python/blob/0b9aee903589af7182db9dfc8cb1f5203332a92f/3.10/bullseye/Dockerfile).
    The Dockerfile has steps to get and build the source. It verifies the GPG signature too.
* The `WORKDIR` line specifies where in the image the subsequent steps will be run. 
  * He's using `/usr/src/app`
  * Wikipedia [Filesystem Hierarchy Standard](https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard)
  * `/usr/src` is for "Source code (e.g., the kernel source code with its header files)."
  * For another app later in the tutorial, he instead creates a directory under `/opt` which is for "Add-on application software packages."
* The `COPY` line copies all files to the image. 
* `RUN` installs dependencies.
* `EXPOSE` says what port to expose. The Flask app is running on port 5000, and so port 5000 is exposed.
* `CMD` tells the container what to run when it's started.
* Build the image:
<pre><code>$ docker build -t stalexan/catnip .
Sending build context to Docker daemon  8.704kB
Step 1/6 : FROM python:3.8     
3.8: Pulling from library/python          
dbba69284b27: Pull complete        
9baf437a1bad: Pull complete 
6ade5c59e324: Pull complete 
b19a994f6d4c: Pull complete 
8fc2294f89de: Pull complete 
79232cc264be: Pull complete 
146537824410: Pull complete 
918b0e031fce: Pull complete 
0138e904e943: Pull complete 
Digest: sha256:e812ab84b97f37b020304e968ed5b16ef6604145437cc3267417979cf0e2ff1e
Status: Downloaded newer image for python:3.8
 ---> cf0ca5be5f0b
Step 2/6 : WORKDIR /usr/src/app
 ---> Running in 4c3fbe00c1b1
Removing intermediate container 4c3fbe00c1b1
 ---> 5146f9d881f4
Step 3/6 : COPY . .
 ---> a2e09e9ce224
Step 4/6 : RUN pip install --no-cache-dir -r requirements.txt
 ---> Running in 4663641cc520
Collecting Flask==2.0.2
  Downloading Flask-2.0.2-py3-none-any.whl (95 kB)
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 95.2/95.2 KB 1.2 MB/s eta 0:00:00
Collecting Jinja2>=3.0
  Downloading Jinja2-3.1.1-py3-none-any.whl (132 kB) 
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 132.6/132.6 KB 2.7 MB/s eta 0:00:00
Collecting click>=7.1.2
  Downloading click-8.1.2-py3-none-any.whl (96 kB)
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 96.6/96.6 KB 13.0 MB/s eta 0:00:00
Collecting Werkzeug>=2.0
  Downloading Werkzeug-2.1.1-py3-none-any.whl (224 kB)
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 224.7/224.7 KB 13.4 MB/s eta 0:00:00
Collecting itsdangerous>=2.0
  Downloading itsdangerous-2.1.2-py3-none-any.whl (15 kB)
Collecting MarkupSafe>=2.0
  Downloading MarkupSafe-2.1.1-cp38-cp38-manylinux_2_17_x86_64.manylinux2014_x86_64.whl (25 kB)
Installing collected packages: Werkzeug, MarkupSafe, itsdangerous, click, Jinja2, Flask
Successfully installed Flask-2.0.2 Jinja2-3.1.1 MarkupSafe-2.1.1 Werkzeug-2.1.1 click-8.1.2 itsdangerous-2.1.2
WARNING: Running pip as the 'root' user can result in broken permissions and conflicting behaviour with the system package manager. It is recommended to use a virtual environment instead: https://pip.pypa.io/war
nings/venv
Removing intermediate container 4663641cc520
 ---> 64a2792786c0
Step 5/6 : EXPOSE 5000
 ---> Running in b3f699abab6d
Removing intermediate container b3f699abab6d
 ---> 65f6cf72a813
Step 6/6 : CMD ["python", "./app.py"]
 ---> Running in f3e197ae43ed
Removing intermediate container f3e197ae43ed
 ---> abf6284cda9e
Successfully built abf6284cda9e
Successfully tagged stalexan/catnip:latest
</code></pre>
* Run the container:
<pre><code>root@logan5:~:139> docker run -p 8888:5000 stalexan/catnip
 * Serving Flask app 'app' (lazy loading)
 * Environment: production
   WARNING: This is a development server. Do not use it in a production deployment.
   Use a production WSGI server instead.
 * Debug mode: off
 * Running on all addresses (0.0.0.0)
   WARNING: This is a development server. Do not use it in a production deployment.
 * Running on http://127.0.0.1:5000
 * Running on http://172.17.0.2:5000 (Press CTRL+C to quit)
</code></pre>
* Browse to site: <http://172.17.0.2:5000/>

# Docker Tutorial: Docker on AWS; Fri 2022-04-15
* Same tutorial continued [here](https://docker-curriculum.com/#docker-on-aws).
* Publish the image to Docker Hub. Optionally this could be some other registry, including a self-hosted registry.
  This is done so AWS can find the image.
<pre><code>$ docker login
Login with your Docker ID to push and pull images from Docker Hub. If you don't have a Docker ID, head over to https://hub.docker.com to create one.
Username: stalexan
Password: 
WARNING! Your password will be stored unencrypted in /root/.docker/config.json.
Configure a credential helper to remove this warning. See
https://docs.docker.com/engine/reference/commandline/login/#credentials-store
Login Succeeded

$ docker push stalexan/catnip
Using default tag: latest
The push refers to repository [docker.io/stalexan/catnip]
9a050616d508: Pushed 
574305f019b7: Pushed 
0d97ffb8d840: Pushed 
2ee62ae2f903: Mounted from library/python 
793849be4c50: Mounted from library/python 
db91079cf39e: Mounted from library/python 
74fa5149f3c8: Mounted from library/python 
c5579a205adc: Mounted from library/python 
7a7698da17f2: Mounted from library/python 
d59769727d80: Mounted from library/python 
348622fdcc61: Mounted from library/python 
4ac8bc2cd0be: Mounted from library/python 
latest: digest: sha256:85136765973c5a7dadb3013189f16f660f1fcd10af1eacbf5050664dd7313683 size: 2844
</code></pre>
* The message is my unencrypted password will be stored to `/root/.docker/config.json` but instead I see
  a authorization token or hash?
* It's just base64 encoded. See [Docker Login the Right Way](https://luiscachog.io/docker-login-the-right-way/)
* The above article describes how to use `docker-credential-secret service`, a utility provided by Docker.
* [How to set up secure credential storage for Docker](https://www.techrepublic.com/article/how-to-setup-secure-credential-storage-for-docker/):
  Describes how to use another tool provided by Docker: `docker-credential-pass`
* Look into using a [Credentials store](https://docs.docker.com/engine/reference/commandline/login/#credentials-store)
* For now, just log out after pushing the image:
<pre><code>docker logout
</code></pre>
* Anyone can now run the image with:
<pre><code>docker run -p 8888:5000 stalexan/catnip
</code></pre>
* Wikipedia [AWS Elastic Beanstalk](https://en.wikipedia.org/wiki/AWS_Elastic_Beanstalk)
* Browse to "Elastic Beanstalk" in the AWS console.
* Click "Create new application"
* Application name: catnip-tutorial-foobar123
* Platform: Docker 
* Select "Upload your code" and "Local file". Browse to `Dockerrun.aws.json` in
  `~/dev/github-prakhar1989/docker-curriculum/flask-app`
* Click "Create application"
* I get "Creating Catniptutorialfoobar123-env This will take a few minutes..." and some logging messages.       
* Took 5 minutes to come up.
* URL: <http://catniptutorialfoobar123-env.eba-yi2223fg.us-west-1.elasticbeanstalk.com/>    
* "Terminate environment" to cleanup, and not be charged resources for keeping the application around.
* I got a message about it was going to take a few minutes to cleanup. I still see the environment and application.
* This time I deleted the application and now everything's gone.

# Tutorial for Developing a Web App Inside a Docker Container Using VS Code; Fri 2022-04-15 <span id=vs-code-220415 />
* Continued from here: [Angular Needs Node.js Too](../Dev/weight-log.html#angular-220415)
* Also, see previous tutorial on Docker in general: [Docker Tutorial: Image and Container Basics](#tutorial-220415)
* docs.microsoft.com [Tutorial: Create and share a Docker app with Visual Studio Code](https://docs.microsoft.com/en-us/visualstudio/docker/tutorials/docker-tutorial)
* code.visualstudio.com [Developing inside a Container](https://code.visualstudio.com/docs/remote/containers)
* code.visualstudio.com [Remote development in Containers](https://code.visualstudio.com/docs/remote/containers-tutorial)
* code.visualstudio.com [Docker in Visual Studio Code](https://code.visualstudio.com/docs/containers/overview)
* Install the [Remote Development](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.vscode-remote-extensionpack) extension pack on the host (logan5).
* Tue 2022-04-26: Do this from within VS Code: `ctrl-p` and then `ext install ms-vscode-remote.vscode-remote-extensionpack`.
* Install the "Remote - Containers" extension: `vscode:extension/ms-vscode-remote.remote-containers`
* Click on the green Remote Status bar icon in the lower left corner of VS Code and click 
  "Remote-Containers: Try a Development Container Sample...". Select the Node sample from the list.
* So that worked. Now try creating my own version of this, debugging code in the python-dev container I've created.
* code.visualstudio.com [Create a development container](https://code.visualstudio.com/docs/remote/create-dev-container)
* Open `/home/sean/dev/elias/docker/python-dev` folder in VS Code.
* `F1` and then "Remote-Containers: Add Development Container Configuration Files...". Select "From Dockerfile...".
* This creates `.devcontainer\devcontainer.json` and prompts with "Folder contains a Dev Container configuration file. Reopen folder to develop in a container".
* Click "Reopen in Container"
* The container build fails on the Dockerfile step `Step 11/25 : COPY files-to-copy/bashig-python-dev.tar.gz /home/sean/tmp`.
* This is because there's no build context that has the file it's trying to copy.
* The `build` script creates the context, but VS Code is only running the Dockerfile.
* I think maybe to start it will be best to skip creating the context. 
* Create a new Dockerfile and image that is called `python-vsdev` (instead of `python-dev`).
* The context might not be as important since most of my work will now be in VS Code on the host versus on the command line in the container.
* That worked, and I'm now running VS Code on logan5 with a container built from `elias\docker\python-vsdev\Dockerfile`.
* The log for the container build is interesting.
* The docker build is done with:
<pre><code>docker build -f /home/sean/dev/elias/docker/python-vsdev/Dockerfile \
    -t vsc-python-vsdev-878cb8db87710f59b7bf0b3e220b0df2 \
    /home/sean/dev/elias/docker/python-vsdev
</code></pre>
* So it looks like if I want to create a context, I'd create a "pre-build"
  script that copied things from `elias\docker\python-vsdev\Dockerfile` to
  a temp folder, along with "files to copy", and then I'd open that temp folder
  in VS code and have the container build happen there.
* VS code then takes the image created from my Dockerfile and uses it to create another image.
* A container is then started with:
<pre><code>docker run --sig-proxy=false -a STDOUT -a STDERR --mount type=bind,source=/home/sean/dev/elias/docker/python-vsdev,target=/workspaces/python-vsdev \
    --mount type=volume,src=vscode,dst=/vs    code -l vsch.local.folder=/home/sean/dev/elias/docker/python-vsdev \
    -l vsch.quality=stable --entrypoint /bin/sh vsc-python-vsdev-878cb8db87710f59b7bf0b3e220b0df2-uid -c echo Container started
</code></pre>
* I think it's best not to do much customization at all, of the image. VS Code runs it in its own particular way.
* The container volume the development files go in the same directory as the Dockerfile. In the container it's `/workspaces/python-vsdev `.
* This all works. When I create new file `foo.py` the directory shown in VS
  Code for the file is `/workspaces/python-vsdev`. I see the file in
  `elias/docker/python-vsdev`.
* So really the best approach seems to be:
  * Create the Dockerfile inside the development folder (e.g. `~/dev/current/stock-ticker`)
  * Don't customize the Dockerfile much. The container commandline won't be used much, just for basic
    things from within VS Code (via "Terminal: New Terminal").
* __Steps then__:
  * Create a basic Dockerfile in a development folder. The only customization needed is to install packages.
  * In VS Code `F1` and then "Remote-Containers: Add Development Container Configuration Files...". Select "From Dockerfile...".
  * Click "Reopen folder to develop in a container"."
* 2022-04-26: The user inside containers created this way is root. Look into
  creating a non-root user, and running as that user instead.

# docs.microsoft.com [Tutorial: Create and share a Docker app with Visual Studio Code](https://docs.microsoft.com/en-us/visualstudio/docker/tutorials/docker-tutorial); 2022-04-16
* Install "Docker" VS Code extension.
* Start container:
<pre><code>docker run -d -p 80:80 docker/getting-started
</code></pre>
* Docker in VS Code shows "Failed to connect. Is Docker running?"
* One of the listed prereqs is "Docker Desktop" but it only runs on Windows and Mac.
* I think the problem I'm seeing now is probably because the Docker daemon is only
  available to root, and I'm running VS Code as sean.
* code.visualstudio.com [Docker in Visual Studio Code](https://code.visualstudio.com/docs/containers/overview)
* docs.docker.com [Manage Docker as a non-root user](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user)
* "Warning: The docker group grants privileges equivalent to the root user. For
  details on how this impacts security in your system, see [Docker Daemon Attack
  Surface](https://docs.docker.com/engine/security/#docker-daemon-attack-surface)."
* So once I've configured `sean` to manage Docker, it's essentially as if
  `sean` is now root.
* Add `sean` to `docker` group:
<pre><code>usermod -aG docker sean
</code></pre>
* Configure Docker to start on boot:
<pre><code>systemctl enable docker.service
systemctl enable containerd.service
</code></pre>
* The tutorial changes `.js` files that are used in the image. It rebuilds the image
  each time `.js` files are changed. Is there a way to modify files in a running container?
* And, is there a way to run commands in the container. My goal for now is to work through
  the code from the book "JavaScript: The Definitive Guide" in a container using Node.js.
* code.visualstudio.com [Developing inside a Container](https://code.visualstudio.com/docs/remote/containers)

# Node.js Docker Container for Learning; 2022-04-17
* I want to run Node.js in a container to learn Node.js and JavaScript.
* Requirements:
  * Ability to save files.
  * Command line access for running JavaScript.
  * Development using VS Code on host.
* Download image:
<pre><code>$ docker pull node
Using default tag: latest
latest: Pulling from library/node
...
Digest: sha256:3f8047ded7bb8e217a879e2d7aabe23d40ed7f975939a384a0f111cc041ea2ed
Status: Downloaded newer image for node:latest
docker.io/library/node:latest

$ docker images
REPOSITORY   TAG       IMAGE ID       CREATED      SIZE
node         latest    8778d77035e2   4 days ago   991MB
</code></pre>
* Run container interactively:
<pre><code>$ docker run -it node bash
root@64261d02eea6:/# pwd
/
root@64261d02eea6:/# whoami
root
root@64261d02eea6:/# ls
bin  boot  dev  etc  home  lib  lib64  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
root@64261d02eea6:/# ls home
node
root@64261d02eea6:/# exit
exit
</code></pre>
* digitalocean.com [How To Share Data Between the Docker Container and the Host](https://www.digitalocean.com/community/tutorials/how-to-share-data-between-the-docker-container-and-the-host)
* Run container interactively as `node` user with shared files:
<pre><code>docker run --rm --user node -v /home/sean/dev/current/js/flanagan-book:/home/node/book -it node bash
</code></pre>
* Now files edited on host in `/home/sean/dev/current/js/flanagan-book` can be seen in container in `/home/node/book`.

# Docker Build and Checking Signatures; Mon 2022-04-18
* The official Rust Docker image doesn't check the signed package that it
  downloads. It seems it should, although the keys would probably have to be
  downloaded over HTTPS too. Or how are GPG signature checks done in Docker
  builds? Downloading the keys over HTTPS for a check ins't great, but it's
  another layer of "defense in depth". If the keys are later checked and found
  to be wrong, that raises a red flag that can be followed up on.
* One possibility might be that you provide the public key to the Docker build.
  Docker still downloads the keys, but can check key hashes.
* See if there are any issues reported in Github for the official Rust Docker build.
* The official Python image checks keys. It places the public key hash in the Dockerbuild
  file, and then uses it to check the private key it downloads. This seems like a great
  compromise. It's not perfect, but having the public key listed plainly within the Dockerfile
  is a great layer in defense in depth.
* The official Debian images Dockerbuild file is just 3 lines long and just uses a compressed
  built image of Debian that's stored on hub.docker.com. There are no key checks.

# Docker Container For Rust Development; Wed 2022-04-20 <span id=rust-container-220420 />
* Continued from [here](../DevRust/drinks-tally.html#docker-container-220418).
* I want to create a Docker container for Rust development.
* Requirements:
  * Container runs in interactive mode.
  * tmux installed, and bash configured with all settings from bashig.
  * vim
  * `rusty` user added and by default container is accessed as that user.
  * Development files mounted from bendor4:./dev-rust-mnt to /home/rusty/dev-rust-mnt
  * Container name: rust-dev
  * Base image: `rust` official image
  * target dir for build output is in container and not on bendor4.
* Build with:
<pre><code>$ docker build -t stalexan/rust-dev .

$ docker images
REPOSITORY          TAG          IMAGE ID       CREATED         SIZE
stalexan/rust-dev   latest       42b86045e9ce   3 seconds ago   1.3GB
rust                1-bullseye   5593c6ce4c4e   12 days ago     1.3GB
rust                latest       5593c6ce4c4e   12 days ago     1.3GB
hello-world         latest       feb5d9fea6a5   6 months ago    13.3kB

</code></pre>
* The `-t` flag names the image `stalexan/rust-dev`. Optionally I could have tagged
  it by appending `:some-tag`.
* Run with:
<pre><code>docker run --rm --it stalexan/rust-dev bash
</code></pre>
* When an new images is created the previous becomes unnamed:
<pre><code>REPOSITORY          TAG          IMAGE ID       CREATED          SIZE
stalexan/rust-dev   latest       9177ffcd1b3c   13 seconds ago   1.36GB
&lt;none&gt;              &lt;none&gt;       42b86045e9ce   15 minutes ago   1.3GB
</code></pre>
* To remove these unnamed images:
<pre><code>docker image prune
</code></pre>
* Files that are copied to the container (e.g.
  `~/dev/current/bashig/files/bash/bashrc-for-etc`) need to be in the "build context". This
  is the directory from which `docker build` is invoked. Everything in this directory is copied to the Docker daemon, so it shouldn't have much. 
  Only files in the context will be available to the build. 
* What's the best approach for this? I don't want `~/dev/current/rust-dev` to have temporary files that I copy there just for the build.
* Have a separate build context directory. Create it from a `build` bash script in the same directory as the Dockerfile.
* Files are in `~/dev/current/rust-dev`.
* `build` has:
<pre><code>#!/usr/bin/env bash

# Creates context for docker build and builds.
# Then to run container:
#     docker run --rm -it -v /home/sean/rust-dev-vol:/home/rusty/rust-dev-vol stalexan/rust-dev bash

# Options
OWNER=stalexan
IMAGE=rust-dev
CONTEXT_BASE="/root/tmp"

# Show trace debugging statements.
set -x

# Exit on error
set -e

# Where are we?
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Executables
CP=/usr/bin/cp
DOCKER=/usr/bin/docker
MKDIR=/usr/bin/mkdir
RM=/usr/bin/rm

# Create build context
CONTEXT="${CONTEXT_BASE}/${IMAGE}-context/"
$RM -rf "$CONTEXT"
$MKDIR -p "$CONTEXT"
cd "$CONTEXT"

# Copy files to context that will then be copied to image.
FILES_TO_COPY="$CONTEXT/files-to-copy"
$MKDIR -p "$FILES_TO_COPY"
$CP /home/sean/dev/current/bashig/package-for-host-results/bashig-rust-dev.tar.gz "$FILES_TO_COPY"

# Build
$DOCKER build -t ${OWNER}/${IMAGE} -f ${SCRIPT_DIR}/Dockerfile .
</code></pre>
* `Dockerfile` has:
<pre><code>FROM rust:1-bullseye

# Install tools
USER root
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends apt-utils \
        locales \
        vim \
        tmux \
        bash-completion git;

# Create user rusty
USER root
RUN set -eux; \
    groupadd rusty; \
    useradd --gid rusty --create-home --shell /bin/bash rusty; \
    mkdir /home/rusty/tmp; \
    chmod go-rxw /home/rusty; \
    chown -R rusty:rusty /home/rusty;

# Configure git
USER rusty
RUN set -eux; \
    git config --global init.defaultBranch main; \
    git config --global user.name "Sean Alexandre"; \
    git config --global user.email "sean@alexan.org";

# Configure bash, vim, tmux, and users
USER rusty
COPY files-to-copy/bashig-rust-dev.tar.gz /home/rusty/tmp
WORKDIR /home/rusty/tmp
RUN set -eux; \
    tar xvzf bashig-rust-dev.tar.gz
USER root
WORKDIR /home/rusty/tmp/bashig-rust-dev
RUN set -eux; \
    ./bashig;

# Run container as rusty
USER rusty
WORKDIR /home/rusty
</code></pre>
* Wed 2022-04-20: `cargo build` works when I first log in but not after I start tmux.
* The path before starting tmux is:
<pre><code>/usr/local/cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
</code></pre>
* After starting tmux it's:
<pre><code>/usr/local/bin:/usr/bin:/bin:/usr/local/games:/usr/games
</code></pre>
* It looks like the tmux version comes from `/etc/profile`, which has:
<pre><code># /etc/profile: system-wide .profile file for the Bourne shell (sh(1))
# and Bourne compatible shells (bash(1), ksh(1), ash(1), ...).

if [ "$(id -u)" -eq 0 ]; then
  PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
else
  PATH="/usr/local/bin:/usr/bin:/bin:/usr/local/games:/usr/games"
fi
export PATH
</code></pre>
* The [Rust Dockerfile](https://github.com/rust-lang/docker-rust/blob/59ffcf40eed2aca0160790f0ae2f0288868e0291/1.60.0/bullseye/Dockerfile) has:
<pre><code>ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH \
    RUST_VERSION=1.60.0
</code></pre>
* So this seems to be just a conflict between how Docker containers are expected to run
  and how I'm running dev-rust. Normally they're standalone and run a service. I'm
  running it interactively.
* It seems the solution is to just add a `PATH=/usr/local/cargo/bin:$PATH` to my user environment, probably
  in `.bashrc_custom`:
<pre><code># The official Rust image puts /usr/local/cargo/bin in PATH, but this gets
# overwritten when starting our own shell with tmux, etc. Add it back.
if &lsqb;&lsqb; :$PATH: != &ast;:"/usr/local/cargo/bin":&ast; &rsqb;&rsqb; ; then
    PATH=/usr/local/cargo/bin:$PATH
fi
</code></pre>
* That worked.
* Wed 2022-04-20: All crates have to be downloaded every time I start the container.
  It's it possible to restart a container, versus creating a new one each time?
  That seems counter to how Docker containers are supposed to work, though. It seems
  it would be better to create another volume for build output.
* stackoverflow.com [Where are modules installed by Cargo stored in a Rust project?](https://stackoverflow.com/questions/49844681/where-are-modules-installed-by-cargo-stored-in-a-rust-project)
* By default crates are stored in `~/.cargo/registry`
* stackoverflow.com [How can the location of Cargo's configuration directory be overridden?](https://stackoverflow.com/questions/38050995/how-can-the-location-of-cargos-configuration-directory-be-overridden)
* `CARGO_HOME` can be set to an alterate location.
* Seems best place for this would be `rust-dev-vol/.cargo`. Then don't back this up since they're just
  temp files.
* Although, this assumes I'm only ever running one instance of the image at
  a time. That's the plan, but I'd need to be careful not to accidently run
  two. I think for deveopment purposes this is fine.
  * Actually, I think it would support multiple containers since that would be as if 
    a user were doing multiple cargo builds from one user account, which it seems must be supported.
* Wed 2022-04-20: Build target is currently `drinks-tally/target` and so everything  
  gets saved to a directory that's backed up. How to put this in a temp dir?
* The temp dir needs to be in `rust-dev-vol` since the container needs access to it. 
* The Cargo Book [Configuration](https://doc.rust-lang.org/cargo/reference/config.html)
* So it looks like I want a `drinks-tally/.cargo/config.toml` that has: 
<pre><code>[build]
target-dir = "../../tmp/.rust-targets/drinks-tally"
</code></pre>
* That worked.
* Thu 2022-04-21: Accented characters are displayed as an undercore.
* stackoverflow.com [How to set the locale inside a Debian/Ubuntu Docker container?](https://stackoverflow.com/questions/28405902/how-to-set-the-locale-inside-a-debian-ubuntu-docker-container)
* hub.docker.com Debian [Locales](https://hub.docker.com/_/debian?tab=description&name=stable-20220316)

# Docker Container For Node Development; Fri 2022-04-22 <span id=node-container-220422 />
* I'm going to create a Docker Container for Node.js development.
* Move rust-dev code to elias where it can be cloned to logan5, and then do development on logan5.
* Create bare repo on beorn@elias:
<pre><code>$ mkdir -p ~/repos/dev/docker/rust-dev.git
$ cd !$
$ git init --bare 
</code></pre>
* Clone repo to sean@bendor4:
<pre><code>cd ~/dev/elias/docker
git clone ssh://beorn@elias:/home/beorn/repos/dev/docker/rust-dev.git
</code></pre>

# Docker Container For Python Development; Mon 2022-04-25 <span id=docker-for-python-220425 />
* I want to get a Docker container running for Python development too.
* I'll be using lots of recent packages and this will be better for security.
* Also, I can use Python 3.10 without having to build it on logan5. 
* More thoughts here: Dev/stock-ticker [Docker Container for Development](../Dev/stock-ticker.html#docker-container-220425).
* This will probably be more similar to the Node container than the Rust container, since the Python
  container comes with a user predefined similar to the Node container.
* Create bare repo on beorn@elias:
<pre><code>$ mkdir -p ~/repos/dev/docker/python-dev.git
$ cd !$
$ git init --bare 
</code></pre>
* Clone repo to sean@logan5:
<pre><code>cd ~/dev/elias/docker
git clone ssh://beorn@elias:/home/beorn/repos/dev/docker/python-dev.git --config core.sshCommand="ssh -i ~/.ssh/logan5.sean.id_ed25519.2022-03-10"
</code></pre>
* Actually the python Docker image does not create a python user.
* Create one, called sean.

# Supplying Password to Dockerfile; 2022-05-27
* I'm creating a Postgres database in a Dockerfile and need to supply a password. What's the best way to do this?
* stackoverflow.com [Docker and securing passwords](https://stackoverflow.com/questions/22651647/docker-and-securing-passwords)
* Create an environment file (e.g. `Dockerfile.env`) that has passwords. Add
  `Dockerfile.env` to `.gitignore`.  Pass `Docker.env` to `docker build` with
  the `--env-file` option.

# linode.com [How to Connect Docker Containers](https://www.linode.com/docs/guides/docker-container-communication/) (2021-08) <span id=connect-containers-210801 />
* Creates 2 Docker containers: one that runs a Postgres database with a simple
  "hello world" table, and another that runs a Node app to query the table.
* See code here: `dev/current/docker-connect`
* __Part 1__ of article: Have Node container talk to Postgres on host.
* __To see the internal IP address of the docker host__ (`172.17.0.1`):
<pre><code>$ ifconfig docker0

docker0: flags=4099<UP,BROADCAST,MULTICAST>  mtu 1500
        <b>inet 172.17.0.1</b>  netmask 255.255.0.0  broadcast 172.17.255.255
        inet6 fe80::42:1cff:fed5:9fbc  prefixlen 64  scopeid 0x20<link>
</code></pre>
* "Since 172.17.0.1 is the IP of the Docker host, all of the containers on the
  host will have an IP address in the range 172.17.0.0/16."
* "The `--add-host` option defines a database host, which points to the IP
  address of the Docker host. Declaring the database host at runtime, rather
  than hard-coding the IP address in the app, helps keep the container
  reusable."
* "Run the node image again. This time, instead of `--add-host`, use the
  `--link` option to connect the container to pg_container"
* __To see IP address of a container__ (`172.17.0.2`):
<pre><code>$ docker exec -it hello-world-node-container <b>ip a</b>

1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
17: eth0@if18: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    <b>inet 172.17.0.2/16</b> brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
</code></pre>
* __There are 2 config changes needed to allow Node container to talk to Postgress on host__.
* __The first__ is `/etc/postgresql/13/main/postgresql.conf` needs to listen for connections on
  more than just `localhost`:
<pre><code>listen_addresses = '&ast;'
</code></pre>
* __The second__ allows connections from `docker0` through firewall in `/etc/iptables-init`:
<pre><code># Allow input from Docker containers
$IPT4 -A INPUT -i docker0 -j ACCEPT 
$IPT6 -A INPUT -i docker0 -j ACCEPT 
</code></pre>
* __Part 2__ of article: Have Node container talk to Postgres in another container.
* To run Postgres container:
<pre><code>docker run --rm -d -e POSTGRES_PASSWORD=xxx -v `pwd`:/backup/ --name pg_container postgres
</code></pre>
* Run the `backup.sql` script inside the container:
<pre><code>$ docker exec -it pg_container bash
$ cd /backup
$ psql -U postgres -f backup.sql postgres
$ exit
</code></pre>
* I had to manually reset the password for postgres database user. My guess is this
  has something to do with the password hash made on the host (and inside backup.sql)
  being different from how passwords are hashed in the container (based on Alpine Linux.)
* Only one config change was needed to have the Node container start talking to
  Postgres in the Postgres container instead of the host.  __Change
  `--add-host=database:172.17.0.1` for the `docker run` command to
  `--link=pg_container:database`.__
* __Part 3__ of article: Use __Docker Compose__. <span id=docker-compose-220530 />
* docs.docker.com [Install Docker Compose](https://docs.docker.com/compose/install/)
* Install Docker Compose:
<pre><code>apt-get install docker-compose-plugin
</code></pre>
* Check version:
<pre><code>$ docker compose version
Docker Compose version v2.5.0
</code></pre>
* Bring containers up:
<pre><code>$ docker compose up --detach
</code></pre>
* To bring the containers own, and remove them (equivalent of having run with `docker run --rm`):
<pre><code>$ docker compose down
</code></pre>
* linode.com [How to Use Docker Compose](https://www.linode.com/docs/guides/how-to-use-docker-compose/) (2021-08):
  * "In Docker a service is the name for a “Container in production”."
  * "Caution: The example `docker-compose.yml` above uses the `environment`
    directive to store MySQL user passwords directly in the YAML file to be
    imported into the container as environment variables. This is not
    recommended for sensitive information in production environments. Instead,
    sensitive information can be stored in a separate `.env` file (which is not
    checked into version control or made public) and accessed from within
    `docker-compose.yml` by using the `env_file` directive."
* docs.docker.com [Compose specification](https://docs.docker.com/compose/compose-file/)
* docs.docker.com [docker compose](https://docs.docker.com/engine/reference/commandline/compose/)

# VS Code and devcontainer.json; 2022-06 
* To map a port, doing what in docker-compose.yml is the equivalent of
  `-ports`. The first port is on the host, and the second in the container:
<pre><code>"appPort": [ "3005:3000" ]
</code></pre>

# Keeping Images Up To Date; Sat 2022-07-02
* It turns out docking just a __`docker build` does not pull the latest version of
  base images. Base images have to be manually deleted and then pulled, with `docker rmi`
  and `docker pull`.__
* Find or write a script that checks for out of date images.
* Also, find a script that checks any `requirements.txt` for Python package updates needed, 
  and something similar for Node updates.
* What's a good strategy for this? Deleting a base image requires that all dependent images
  be deleted too. So, for example, if I delete the `node` image for say `try-react18` then
  I have to delete the images for all other projects.
* I see there's a `docker scan` command that will look for vulnerabilities. However, it
  requires that your image be built on Docker Hub.
* serverfault.com [How to keep applications in own Docker images up to date?](https://serverfault.com/questions/861322/how-to-keep-applications-in-own-docker-images-up-to-date) (2017)
* It looks like __`docker build` will uses the latest images with the options `--no-cache --pull`__.
* The official documentation doesn't have much on these options, but I do see this in [docker build](https://docs.docker.com/engine/reference/commandline/build/):
  * `--no-cache`: "Do not use cache when building the image"
  * `--pull`: "Always attempt to pull a newer version of the image"
* It seems that just `--pull`.
* stackoverflow.com [Does Docker build --no-cache actually download and refresh the base image?](https://stackoverflow.com/questions/52664744/does-docker-build-no-cache-actually-download-and-refresh-the-base-image)
* So `--no-cache` will not use the locally cached versions of layers that have been built locally. However,
  it will not fetch new versions downloaded with the `FROM` command. For that, use `--pull`.
* __So to get the latest of everything, use both `--no-cache` and `--pull`.__
  * `--pull` will ensure the latest base images are pulled.
  * `--no-cache` will ensure that packages installed within an image are the latest.

# docs.docker.com [Best practices for writing Dockerfiles](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#decouple-applications)
* Ideally __each container should run just one application__. 
  * See docs.docker.com [Decouple applications](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#decouple-applications)
  * [The Twelve-Factor App](https://12factor.net/): [VI. Processes](https://12factor.net/processes)
* __Create a build context (separate dir) for `docker build`__ since everything in
  the directory where the build is done will be copied to the build daemon and
  image, resulting in a longer build process and larger image than necessary.
* __Use a `.dockerignore` file__ to exclude files from being sent to the docker daemon and image.
* Order commands so that those that change more frequently come later. This allows 
  cached layers from previous builds to be reused.
* "Whenever possible, use current official images as the basis for your images.
  __We recommend the Alpine image__ as it is tightly controlled and small in size
  (currently under 6 MB), while still being a full Linux distribution."
* "Always combine `RUN apt-get update` with `apt-get install` in the same RUN statement";
  see [here](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#run).
* `ENTRYPOINT` can be used in combination with `CMD` to allow users to specify options to 
  pass to the application that's run in the container. `ENTRYPOINT` is set to just the name of the
  application to run. `CMD` is set to any default options. If the user later runs the image
  and gives command line parameters, the `CMD` parameters are ingored and the custom parameters
  are used instead. See [here](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#entrypoint).
* Use the `USER` command if the process that's run can be run as a non-root user; see [here](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user).
  "__Avoid installing or using `sudo`__ as it has unpredictable TTY and
  signal-forwarding behavior that can cause problems. If you absolutely need
  functionality similar to `sudo`, such as initializing the daemon as `root` but
  running it as non-`root`, consider using “__gosu__”.
* github.com/tianon [gosu](https://github.com/tianon/gosu): "The core use case
  for gosu is to step down from root to a non-privileged user during container
  startup (specifically in the ENTRYPOINT, usually)."
  * Has great simple steps for checking the package's GPG signature.
  * "`su-exec` is a very minimal re-write of gosu in C, making for a much smaller binary, and is available in the main Alpine package repository."

# lwn.net [Docker and the OCI container ecosystem](https://lwn.net/Articles/902049/) (2022-07-26)
* "Failure to limit access to Docker's socket can be a significant security
  hazard. By default dockerd runs as root. Anyone who is able to connect to the
  Docker socket has complete access to the API. Since the API allows things
  like running a container as a specific UID and binding arbitrary filesystem
  locations, it is trivial for someone with access to the socket to become root
  on the host. __Support for running in rootless mode was added in 2019 and
  stabilized in 2020__, but is still not the default mode of operation."
* docs.docker.com [Run the Docker daemon as a non-root user (Rootless mode)](https://docs.docker.com/engine/security/rootless/)

# Mount Versus Volume; 2022-08
* [Docker Tip #33: Should You Use the Volume or Mount Flag?](https://nickjanetakis.com/blog/docker-tip-33-should-you-use-the-volume-or-mount-flag)
* It sounds like the general recommendation is to use `--mount` versus `--volume` since it's more explicit and
  doesn't automatically create a directory on the host if it doesn't exist.

# Auto Restart; 2022-08-06
* docs.docker.com [Start containers automatically](https://docs.docker.com/config/containers/start-containers-automatically/)
  "Docker provides restart policies to control whether your containers start
  automatically when they exit, or when Docker restarts. Restart policies
  ensure that linked containers are started in the correct order. Docker
  recommends that you use restart policies, and avoid using process managers to
  start containers."
* Option to add to `docker run`: `--restart always`
* This is incompatible with the `--rm` option, though.
* Docker Compose has a `restart` option. See [here](https://docs.docker.com/compose/compose-file/#restart).

# Secrets <span id=secrets />
* diogomonica.com [Why you shouldn't use ENV variables for secret data](https://blog.diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/) (2017)
* Use `docker secret` for keys and passwords.
* Create a secret:
<pre><code>openssl rand -base64 32 | docker secret create secure-secret -
</code></pre>
* Use a secret by passing to a container:
<pre><code>docker service create --secret="secure-secret" redis:alpine
</code></pre>
* Then in the container the secret is available in the file `/run/secrets/secure-secret`
* docker.com [Manage sensitive data with Docker secrets](https://docs.docker.com/engine/swarm/secrets/)
* docker.com [Compose file: secrets](https://docs.docker.com/compose/compose-file/compose-file-v3/#secrets)
* Attempting to list secrets results in an error:
<pre><code>$ docker secret ls
Error response from daemon: This node is not a swarm manager. Use "docker swarm
init" or "docker swarm join" to connect this node to swarm and try again.
</code></pre>
* docs.docker.com [Swarm mode overview](https://docs.docker.com/engine/swarm/)
  "Current versions of Docker include swarm mode for natively managing
  a cluster of Docker Engines called a swarm. Use the Docker CLI to create
  a swarm, deploy application services to a swarm, and manage swarm behavior.
  ¶ Docker Swarm mode is built into the Docker Engine. Do not confuse Docker
  Swarm mode with Docker Classic Swarm which is no longer actively developed."
* Hmmm. Or, just store secrets in a file that can only be read by Docker group or root, and mount
  that file in the containers that need it.
* I'm not sure that "docker secrets" would buy much more, looking more closely at it.
  Anyone with root or docker group permissions can display secrets, as if they were
  in a file readable by only root or the docker group.
* So do follow the advice to not use environment variables, but don't use "docker secret" to 
  avoid the overhead of "swarm mode" which I don't need.

# Logging
* docs.docker.com [docker logs](https://docs.docker.com/engine/reference/commandline/logs/)
* View logs for a container:
<pre><code>docker logs [container-name]
</code></pre>
* Although, once the container is gone the logs are gone.
* stackoverflow.com [How to save log files from docker container?](https://stackoverflow.com/questions/42772159/how-to-save-log-files-from-docker-container#42772751)
* docs.docker.com [Configure logging drivers](https://docs.docker.com/config/containers/logging/configure/)
* docs.docker.com [JSON File logging driver](https://docs.docker.com/config/containers/logging/json-file/)

# Docker Content Trust (DCT)
* docs.docker.com [Content trust in Docker](https://docs.docker.com/engine/security/trust/)
* Enable image signature checking:
```
 export DOCKER_CONTENT_TRUST=1
```
