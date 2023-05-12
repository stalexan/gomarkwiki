# LXC
* Wikipedia [LXC](https://en.wikipedia.org/wiki/LXC)
* Website: [linuxcontainers.org](https://linuxcontainers.org/)
* wiki.debian.org [LXC](http://wiki.debian.org/LXC)
* wiki.archlinux.org [Linux Containers](https://wiki.archlinux.org/index.php/Linux_Containers)
* linuxcontainers.org [Getting Started](https://linuxcontainers.org/lxc/getting-started/)
* linuxcontainers.org [Documentation](https://linuxcontainers.org/lxc/documentation/)
* linuxcontainers.org [Security](https://linuxcontainers.org/lxc/security/)
* Stéphane Graber defines LXC containers as "zero-overhead Linux only
  virtualization technology that’s extremely flexible but requires you to share
  your kernel with the host", from
  [here](https://stgraber.org/2013/12/20/lxc-1-0-your-first-ubuntu-container/).
* help.ubuntu.com [LXC](https://help.ubuntu.com/lts/serverguide/lxc.html)

# cgroups
* Wikipedia [cgroups](https://en.wikipedia.org/wiki/Cgroups): "Control
  Groups". Is a Linux kernel feature to limit, account, and isolate 
  resource usage (CPU, memory, disk I/O, etc) of process groups.
* Jonathan Corbet [Process containers](https://lwn.net/Articles/236038/) (2007)
  "A 'container' is a group of processes which shares a set of parameters used
  by one or more subsystems."

# Version 2.0 
* stgraber.org [LXC 2.0 has been released!](https://stgraber.org/2016/04/06/lxc-2-0-has-been-released/) (2016-04)
* wiki.debian.org [LXC](http://wiki.debian.org/LXC):
  "LXC 1.0 is available in Jessie. __LXC 2.0 is available in Stretch__ and Jessie
  Backports. When looking for documentation, howtos and tutorials, please check
  which LXC version they apply to, as things might have changed...  Stretch
  ships with a new major release of LXC, which also includes a helper for
  easier networking setup called __lxc-net__. lxc-net allows you to set up a simple
  bridge with DHCP and NAT for your containers. LXC 2.0 also allows the use of
  unprivileged containers."
* lwn.net [Understanding the new control groups API](https://lwn.net/Articles/679786/)
* kernel.org [Control Group v2](https://www.kernel.org/doc/Documentation/cgroup-v2.txt) (2015-10):
  "This is the authoritative documentation on the design, interface and
  conventions of cgroup v2."

# LXC 2.0 Basics; 2017-07-11
* Package to install on Debian 9 Stretch: `lxc`
* [Linux LXC 2.0 - 2017](http://www.bogotobogo.com/Linux/linux_LXC_Linux_Container_Install_Run.php)
* linuxcontainers.org [Getting Started](https://linuxcontainers.org/lxc/getting-started/)
* Check configuration: `lxc-checkconfig`
* __View list of available images__, and manually pick one to create a container from:
<pre><code>$ lxc-create -t download -n foobar-container
Setting up the GPG keyring                    
Downloading the image index                   
                                              
---                                           
DIST    RELEASE ARCH    VARIANT BUILD         
---                                           
alpine  3.1     amd64   default 20170319_17:50
alpine  3.1     armhf   default 20161230_08:09
alpine  3.1     i386    default 20170319_17:50
alpine  3.2     amd64   default 20170504_18:43
...
debian  stretch amd64   default 20170711_07:43
debian  stretch arm64   default 20170710_22:42
debian  stretch armel   default 20170711_07:43
debian  stretch armhf   default 20170711_07:43
debian  stretch i386    default 20170711_07:43
debian  stretch powerpc default 20161104_22:42
debian  stretch ppc64el default 20170710_22:42
debian  stretch s390x   default 20170710_22:42
...
---

Distribution: debian
Release: stretch
Architecture: amd64

Downloading the image index
Downloading the rootfs
Downloading the metadata
The image cache is now ready
Unpacking the rootfs

---
You just created a Debian container (release=stretch, arch=amd64, variant=default)

To enable sshd, run: apt-get install openssh-server

For security reason, container images ship without user accounts and without
a root password.

Use lxc-attach or chroot directly into the rootfs to set a root password or
create user accounts.
</code></pre>
* Interesting. This looks alot like Docker then. Images are downloaded from somwhere else.
  LXC containers are not just a local configuration change.
* __Start the container__:
<pre><code>$ lxc-start -n foobar-container -d
</code></pre>
* __View list of containers__:
<pre><code>$ lxc-ls --fancy
NAME             STATE   AUTOSTART GROUPS IPV4 IPV6
foobar-container STOPPED 0         -      -    -

$ lxc-info foobar-container
Name:           foobar-container
State:          STOPPED
</code></pre>
* Start container. "If no configuration is defined, the default isolation is
  used. If no command is specified, the lxc-start will use the default
  `/sbin/init` command to run a system container.":
<pre><code>$ lxc-start -n foobar-container -d

$ lxc-info -n foobar-container 
Name:           foobar-container
State:          RUNNING
PID:            6660
CPU use:        0.25 seconds
BlkIO use:      0 bytes
Memory use:     13.14 MiB
KMem use:       2.65 MiB

$ ps -ef | grep 6660
root      6660  6656  0 13:08 ?        00:00:00 /sbin/init
root      6699  6660  0 13:08 ?        00:00:00 /lib/systemd/systemd-journald
root      6753  6660  0 13:08 pts/0    00:00:00 /sbin/agetty --noclear tty1 linux
root      6754  6660  0 13:08 pts/3    00:00:00 /sbin/agetty --noclear tty4 linux
root      6755  6660  0 13:08 pts/2    00:00:00 /sbin/agetty --noclear tty3 linux
root      6756  6660  0 13:08 pts/10   00:00:00 /sbin/agetty --noclear --keep-baud console 115200,38400,9600 vt220
root      6757  6660  0 13:08 pts/1    00:00:00 /sbin/agetty --noclear tty2 linux
</code></pre>
* __Attach a shell__ to the container:
<pre><code>root@fraser2:~# lxc-attach -n foobar-container
root@foobar-container:~#
</code></pre>
* __Stop container__:
<pre><code>lxc-stop -n foobar-container
</code></pre>
* __Delete container__:
<pre><code>$ lxc-destroy -n foobar-container
Destroyed container foobar-container
</code></pre>

# stgraber.org [LXC 1.0: Your first Ubuntu container [1/10]](https://stgraber.org/2013/12/20/lxc-1-0-your-first-ubuntu-container/) (2013)
* From stgraber.org [LXC 1.0: Blog post series](https://stgraber.org/2013/12/20/lxc-1-0-blog-post-series/) (2013)
* __Create a container__ called "foobar2-container using the "debian" template and the same version
  of Debian and architecture as the host, as root:
<pre><code>$ lxc-create -t debian -n foobar2-container

debootstrap is /usr/sbin/debootstrap
Checking cache download in /var/cache/lxc/debian/rootfs-stretch-amd64 ...
Downloading debian minimal ...
I: Retrieving InRelease
I: Retrieving Release
I: Retrieving Release.gpg
I: Checking Release signature
I: Valid Release signature (key id 067E3C456BAE240ACEE88F6FEF0F382A1A7B6500)
I: Retrieving Packages
I: Validating Packages
I: Resolving dependencies of required packages...
I: Resolving dependencies of base packages...
I: Found additional required dependencies: libaudit-common libaudit1 libbz2-1.0
libcap-ng0 libdb5.3 libdebconfclient0 libgcrypt20 libgpg-error0 liblz4-1
libncursesw5 libsemanage-common libsemanage1 libsystemd0 libudev1 libustr-1.0-1
I: Found additional base dependencies: adduser debian-archive-keyring dmsetup
gpgv iproute2 libapparmor1 libapt-pkg5.0 libbsd0 libc-l10n libcap2
libcryptsetup4 libdevmapper1.02.1 libdns-export162 libedit2 libelf1
libgssapi-krb5-2 libidn11 libip4tc0 libisc-export160 libk5crypto3 libkeyutils1
libkmod2 libkrb5-3 libkrb5support0 libmnl0 libncurses5 libprocps6 libseccomp2
libssl1.0.2 libstdc++6 libwrap0 openssh-client openssh-sftp-server procps
systemd systemd-sysv ucf
I: Checking component main on http://httpredir.debian.org/debian...
I: Checking component main on http://httpredir.debian.org/debian...
I: Retrieving libacl1 2.2.52-3+b1
I: Validating libacl1 2.2.52-3+b1
I: Retrieving adduser 3.115
I: Validating adduser 3.115
I: Retrieving libapparmor1 2.11.0-3
I: Validating libapparmor1 2.11.0-3
...
I: Installing core packages...
I: Unpacking required packages...
I: Unpacking libacl1:amd64...
I: Unpacking libattr1:amd64...
I: Unpacking libaudit-common...
I: Unpacking libaudit1:amd64...
I: Unpacking base-files...
I: Unpacking base-passwd...
I: Unpacking bash...
I: Unpacking libbz2-1.0:amd64...
...
I: Unpacking zlib1g:amd64...
I: Configuring required packages...
I: Configuring gcc-6-base:amd64...
I: Configuring lsb-base...
I: Configuring sensible-utils...
I: Configuring ncurses-base...

...
I: Configuring openssh-server...
I: Configuring libc-bin...
I: Configuring systemd...
I: Base system installed successfully.
Download complete.
Copying rootfs to /var/lib/lxc/foobar2-container/rootfs...Generating locales (this might take a while)...
  en_US.UTF-8... done
  en_US.UTF-8... done
Generation complete.
update-rc.d: error: cannot find a LSB script for checkroot.sh
update-rc.d: error: cannot find a LSB script for umountfs
update-rc.d: error: cannot find a LSB script for hwclockfirst.sh
Creating SSH2 RSA key; this may take some time ...
2048 SHA256:X4ysHP2vmABEa9s7trL/dzULVrvs3mYrNI+6FsYXY6E root@fraser2 (RSA)
Creating SSH2 ECDSA key; this may take some time ...
256 SHA256:5BQoV1E9H83vWJhEVwPhY2OMPZfGU2SQID/9UG+Skys root@fraser2 (ECDSA)
Creating SSH2 ED25519 key; this may take some time ...
256 SHA256:qUp1ZgqVbpVSAfUWukHcXXimSXA2OQ2qP6k+OLayPd4 root@fraser2 (ED25519)
invoke-rc.d: could not determine current runlevel
invoke-rc.d: policy-rc.d denied execution of start.

Current default time zone: 'Etc/UTC'
Local time is now:      Tue Jul 11 18:38:29 UTC 2017.
Universal Time is now:  Tue Jul 11 18:38:29 UTC 2017.
</code></pre>
* __Start the container__ (in the background), as root:
<pre><code>$ lxc-start -n foobar2-container -d
</code></pre>
* __Access the container's console__ (ctrl-a + q to detach):
<pre><code>$ lxc-console -n foobar2-container

Connected to tty 1
Type &lt;Ctrl+a q&gt; to exit the console, &lt;Ctrl+a Ctrl+a&gt; to enter Ctrl+a itself
Debian GNU/Linux 9 foobar2-container tty1
foobar2-container login:
</code></pre>
* __However, no user accounts are created by default.__ Earlier output was: "For
  security reason, container images ship without user accounts and without
  a root password. Use lxc-attach or chroot directly into the rootfs to set
  a root password or create user accounts."
* __Spawn bash directly in the container__ (bypassing the console login):
<pre><code>root@fraser2:# lxc-attach -n foobar2-container
root@foobar2-container:~#
</code></pre>
* __SSH into the container__  However, by default, containers don't have users or IP addresses:
<pre><code># Determine IP address.
$ lxc-info -n 

$ ssh &lt;user&gt;@&lt;ip from lxc-info&gt;
</code></pre>
* __Stop container from within an attached shell__, as root:
<pre><code>poweroff
</code></pre>
* __Stop container cleanly from outside__, as root:
<pre><code>lxc-stop -n foobar2-container
</code></pre>
* __Kill container from outside__, as root:
<pre><code>lxc-stop -n foobar2-container -k
</code></pre>

# Templates; 2017-07
* stgraber.org [LXC 1.0: Your second container [2/10]](https://stgraber.org/2013/12/21/lxc-1-0-your-second-container/) (2013)
* __Templates__ are just executables or scripts that produce a working rootfs in the path 
  that’s passed to them.
* Location of templates is `/usr/share/lxc/templates`, in Debian 9 stretch.
* The template for debian containers is the bash script `lxc-debian`. It uses `debootstrap` 
  and `apt-get` to to create the rootfs.
* What a template does is specific to how a given distro is bootstraped.
* Most templates use a local cache, so the initial bootstrap of a container for
  a given arch will be slow. Any subsequent one will just be a local copy from
  the cache and will be much faster.

# Autostarting Containers; 2017-07
* stgraber.org [LXC 1.0: Your second container [2/10]](https://stgraber.org/2013/12/21/lxc-1-0-your-second-container/) (2013)
* [LXC Autostart Container at boot, choose order and delay](https://coderwall.com/p/ysog_q/lxc-autostart-container-at-boot-choose-order-and-delay)
* __Containers can be configured to auto-start__ using the autostart related config
  variables in the container's config file.
* A container's config file is: `/var/lib/lxc/CONTAINER-NAME/config`
* The autostart variables are
  * `lxc.start.auto` = 0 (disabled) or 1 (enabled)
  * `lxc.start.delay` = 0 (delay in second to wait after starting the container)
  * `lxc.start.order` = 0 (priority of the container, higher value means starts earlier)
  * `lxc.group` = group1,group2,group3,… (groups the container is a member of)
* As an example say these are the settings for container p1:
<pre><code>lxc.start.auto = 1
lxc.group = ubuntu
</code></pre>
* And these are the settings for container p2:
<pre><code>lxc.start.auto = 1
lxc.start.delay = 5
lxc.start.order = 100
</code></pre>
* Container p2 will start at boot time because "auto = 1" and it is not in a group; 
  containers in a group do not start at boot.
* The command `lxc-autostart` can be used to start and stop containers configured for
  autostart.
* __To start all containers__ with "auto = 1". Forhh
<pre><code>lxc-autostart -a
</code></pre>
* For the above settings, p2 would start first because it has "order = 100". There
  will be a 5 second delay due to "delay = 5" and then p1 will be started.
* __To restart all containers__ in the "ubuntu" group:
<pre><code>lxc-autostart -r -g ubuntu
</code></pre>
* Add the `-L` argument to see what would happen with out actually doing anything.

# Freezing Containers; 2017-07
* Containers can be frozen. The processes in the container are stopped, but continue 
  to consume memory.
* __To freeze a container__:
<pre><code>lxc-freeze -n CONTAINER-NAME
</code></pre>

# Sharing Data With a Container; 2017-07
* stgraber.org [LXC 1.0: Advanced container usage [3/10]](https://stgraber.org/2013/12/21/lxc-1-0-advanced-container-usage/) (2013)
* __The container's root can be accessed at__ `/var/lib/lxc/CONTAINER-NAME/rootfs/`
* This is the static file system, though. __To see the filesystem at runtime__:
<pre><code>ls -lh /proc/$(sudo lxc-info -n test-001 -p -H)/root/run/
</code></pre>
* __To mount a host directory in the container__, edit `/var/lib/lxc/CONTAINER-NAME/fstab`
  to have the following for example, and restart the container:
<pre><code>/var/cache/lxc var/cache/lxc none bind,create=dir
</code></pre>
* The first path is the host directory to mount and the second is where to mount it 
  in the container. The second path has no beginning `/` to make it relative to the 
  container's root.
* The options used here are: mount it as a bind-mount (“`none`” fstype and “`bind`” option) 
  and create any directory that’s missing in the container (“`create=dir`”). To limit the 
  container to read access add the `ro` option.

# Copying a Container; 2017-07
* stgraber.org [LXC 1.0: Container storage [5/10]](https://stgraber.org/2013/12/27/lxc-1-0-container-storage/) (2013)
* Note: `lxc-copy` has replaced `lxc-clone`
* __To copy a container__, using p1 to create p4:
<pre><code>lxc-copy -n p1 -N p4
</code></pre>
* __To use an overlay file system__ for the clone:
<pre><code>lxc-copy -n p1 -N p1-test -B overlayfs -s
</code></pre>
* Wikipedia [OverlayFS](https://en.wikipedia.org/wiki/OverlayFS)

# Container Updates; 2017-07 <span id=container-updates.2017-07 />
* How are containers updated, when package updates are released?
* cyberciti.biz [How to update Debian or Ubuntu Linux containers (lxc) VM](https://www.cyberciti.biz/faq/how-to-update-debian-or-ubuntu-linux-containers-lxc/)
* Update has to been done individually within each container.
* `lxc-attach` can be used __to update the container from the host__; e.g.
<pre><code>lxc-attach -n CONTAINER-NAME -- apt-get update
lxc-attach -n CONTAINER-NAME -- apt-get -y upgrade
</code></pre>
* [Containers: Just Because Everyone Else is Doing Them Wrong, Doesn't Mean You Have To](https://www.hastexo.com/blogs/florian/2016/02/21/containers-just-because-everyone-else/) (2016): Recommends using __one base container that has
  important packages and then overlayfs containers based on that container__. Make
  the root filesystem readonly to the overlayfs containers. Package updates then
  only have to be made to the base container; e.g. to update libc. All other containers
  get the update by just restarting.

# Container Snapshots; 2017-07
* stgraber.org [LXC 1.0: Container storage [5/10]](https://stgraber.org/2013/12/27/lxc-1-0-container-storage/) (2013)
* __Create a container snapshot__:
<pre><code>echo "before installing apache2" &gt; snap-comment
sudo lxc-snapshot -n p1-lvm -c snap-comment
</code></pre>
* __List snapshots__ for a container:
<pre><code>lxc-snapshot -n p1-lvm -L -C
</code></pre>
* __Revert a container to a previous snapshot__:
<pre><code>lxc-snapshot -n p1-lvm -r snap0
</code></pre>
* __Restore a snapshot as its own container__:
<pre><code>lxc-snapshot -n p1-lvm -r snap0 p1-lvm-snap0
</code></pre>

# Security Features; 2017-07
* stgraber.org [LXC 1.0: Security features [6/10]]() (2013)
* "unless you are using __unprivileged containers__, you shouldn’t give root
  access to a container to someone whom you’d mind having root access to your
  host."
* "A little while back we added __Apparmor profiles__ support to LXC...The
  Apparmor support is rather simple, there’s one configuration option
  `lxc.aa_profile` which sets what apparmor profile to use for the container."
* "__SELinux support__ is very similar to Apparmor’s. An SELinux context can be
  set using `lxc.se_context`."
* "__Seccomp__ is a fairly recent kernel mechanism which allows for filtering
  of system calls.  As a user you can write a seccomp policy file and set it
  using `lxc.seccomp`"
* "And last but not least, __what’s probably the only way of making a container
  actually safe__. LXC now has support for __user namespaces__...LXC is no longer
  running as root so even if an attacker manages to escape the container, he’d
  find himself having the privileges of a regular user on the host...I’ll cover
  how to actually set that up and use those unprivileged containers in the next
  post."

# Unprivileged containers   Unprivilileged; 2017-07
* stgraber.org [LXC 1.0: Unprivileged containers [7/10]](https://stgraber.org/2014/01/17/lxc-1-0-unprivileged-containers/) (2013)
* linuxcontainers.org [Creating unprivileged containers as a user](https://linuxcontainers.org/lxc/getting-started/#creating-unprivileged-containers-as-a-user)
* Suggests downloading pre-built containers because it's hard to configure a container
  correctly so that it functions with limited privileges.
* Seems a part of this is likely that there are tweaks that have to be made to the Linux
  kernel and/or each distro to get unprivileged containers to work. It's more than just
  a script, but patches as well.
* Downloading pre-built images is useful for regular containers to, that are pre-built.
* To __download a pre-built container__ use `-t download`:
<pre><code>lxc-create -t download -n p1 -- -d ubuntu -r trusty -a amd64
</code></pre>
* This starts to feel like Docker again, with images downloads. However, in this
  case it seems the images might not be that different from regular package builds.
* Debian packages, last I heard, were built by individual developers and uploaded to
  Debian servers. 
* cyberciti.biz [How to create unprivileged LXC container on Ubuntu Linux 14.04 LTS](https://www.cyberciti.biz/faq/how-to-create-unprivileged-linux-containers-on-ubuntu-linux/) (2016-02)

# Ubuntu Versus Debian; 2017-07 <span id=ubuntu-vs-debian.2017-07 />
* It seems it might be better to use LXC from an Ubuntu host. 
* LXC is developed by Ubuntu. 
* Also, LXC on Ubuntu comes with more preconfigured; e.g. the network bridge.
* More is locked down. LXC on Ubuntu comes with AppArmor profiles preconfigured, etc.
* It would be good to have the latest LXC for quick bug fixes and release updates.

