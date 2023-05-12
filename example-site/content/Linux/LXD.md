# LXD; 2016-05-09 
* LXD wraps LXC, and makes it easier to use. LXC commands are hyphenated (e.g. `lxc-create`) while
  LXD commands use a space (e.g. `lxc list`).
* Official site: <https://linuxcontainers.org/lxd/>
* Source: <https://github.com/lxc/lxd>
* [Mailing lists](https://lists.linuxcontainers.org/listinfo): 
  [lxc-users](https://lists.linuxcontainers.org/pipermail/lxc-users/)
* ubuntu.com [LXD](http://www.ubuntu.com/cloud/lxd)
* ubuntu.com [Getting started with LXD](https://insights.ubuntu.com/2015/04/28/getting-started-with-lxd-the-container-lightervisor/)
* Package: `lxd`
* Logout and back in after installing for membership to `lxd` group, or run `newgrp lxd`.
* linuxcontainers.org LXD [Getting started - command line](https://linuxcontainers.org/lxd/getting-started-cli/)
* insights.ubuntu.com [The LXD 2.0 Story (Prologue)](https://insights.ubuntu.com/2016/03/14/the-lxd-2-0-story-prologue/) (2016-04)

# insights.ubuntu.com [LXD 2.0: Introduction to LXD](https://insights.ubuntu.com/2016/03/14/lxd-2-0-introduction-to-lxd/); 2016-03
* __Images__ are configured using __profiles__, to create a running __container__.
* LXD ships with a __default profile__ called `default`. It just configures an `eth0` network device for the container.
* __Snapshots__ are identical to containers but are immutable.

# [LXD 2.0: Installing and configuring LXD [2/12]](https://insights.ubuntu.com/2016/03/16/lxd-2-0-installing-and-configuring-lxd-212/); 2016-03
* "LXD supports a number of storage backends. It’s best to know what backend
  you want to use prior to starting to use LXD as we do not support moving
  existing containers or images between backends...Our recommendation is
  __ZFS__ as it supports all the features LXD needs to offer the fastest and
  most reliable container experience."
* "To use ZFS with LXD, you first need ZFS on your system. If using Ubuntu 16.04, 
  simply install it with: `sudo apt install zfsutils-linux`"
* "Every few hours (6 by default), LXD also goes looking for a newer version of
  the image and updates its local copy."

# Creating containers
* [LXD 2.0: Your first LXD container [3/12]](https://insights.ubuntu.com/2016/03/22/lxd-2-0-your-first-lxd-container/) (2016-03)
* "If all you want is the best supported release of Ubuntu, all you have to do is:
  `lxc launch ubuntu:`"
* "Note however that the meaning of this will change as new Ubuntu LTS releases
  are released. So for scripting use, you should stick to mentioning the actual
  release you want"
* "To get the latest, tested, stable image of Ubuntu 14.04 LTS, you can simply run:"
<pre><code>lxc launch ubuntu:14.04
</code></pre>
* "In this mode, a random container name will be picked.
  If you prefer to specify your own name, you may instead do:"
<pre><code>lxc launch ubuntu:14.04 c1
</code></pre>
* "Should you want a specific (non-primary) architecture, say a 32bit Intel image, 
  you can do:"
<pre><code>lxc launch ubuntu:14.04/i386 c2
</code></pre>
* Get more details about a container:
<pre><code>$ lxc info xenial-64
Name: xenial-64
Architecture: x86_64
Created: 2016/05/09 18:46 UTC
Status: Running
Type: persistent
Profiles: default
Pid: 8546
Ips:
  eth0:	inet	10.127.46.227	vethY3S80A
  eth0:	inet6	fd7b:41c4:65ca:51d4:216:3eff:feeb:d109	vethY3S80A
  eth0:	inet6	fe80::216:3eff:feeb:d109	vethY3S80A
  lo:	inet	127.0.0.1
  lo:	inet6	::1
Resources:
  Processes: 9
  Memory usage:
    Memory (current): 9.78MB
    Memory (peak): 10.98MB
  Network usage:
    eth0:
      Bytes received: 2.54kB
      Bytes sent: 1.44kB
      Packets received: 23
      Packets sent: 12
    lo:
      Bytes received: 0 bytes
      Bytes sent: 0 bytes
      Packets received: 0
      Packets sent: 0
</code></pre>
* "If you want to just create a container or a batch of container but not also
  start them immediately, you can just replace `lxc launch` by `lxc init`"
* List all images on the remote `ubuntu`:
<pre><code>$ lxc image list ubuntu:
</code></pre>
* Add a remote called `images`:
<pre><code>$ lxc remote add images images.linuxcontainers.org
Generating a client certificate. This may take a minute...
</code></pre>

# Starting, stoping, and deleting containers
* To start a container:
<pre><code>lxc start [container]
</code></pre>
* To stop a container:
<pre><code>lxc stop [container]
</code></pre>
* To force stop a container:
<pre><code>lxc stop [container] --force
</code></pre>
* To delete a container:
<pre><code>lxc delete [container]
</code></pre>

# Profiles
* List available profiles:
<pre><code>lxc profile list
</code></pre>
* See the content of a given profile:
<pre><code>lxc profile show [profile]
</code></pre>
* Edit a profile:
<pre><code>lxc profile edit [profile]
</code></pre>
* To change the list of profiles which apply to a given container:
<pre><code>lxc profile apply [container] [profile1],[profile2],[profile3],...
</code></pre>

# Editing a container
* For things that are unique to a container and so doesn't make sense to put into 
  a profile, you can just set it directly against the container.
<pre><code>lxc config edit [container]
</code></pre>
* Instead of opening the whole thing in a text editor, you can also modify individual 
  keys with:
<pre><code>lxc config set [container] [key] [value]
</code></pre>
* Or add devices, for example:
<pre><code>lxc config device add my-container kvm unix-char path=/dev/kvm
</code></pre>
* Show read the container local configuration:
<pre><code>lxc config show [container]
</code></pre>
* For the expanded configuration (including all the profile keys):
<pre><code>lxc config show --expanded [container]
</code></pre>
* "unless indicated in the documentation, all configuration keys and device
  entries are __applied to affected containers live__. This means that you can
  add and remove devices or alter the security profile of running containers
  without ever having to restart them."
* Edit a file in the container:
<pre><code>lxc file edit [container]/[path]
</code></pre>

# Cloning and renaming containers
* Clone a container. Resulting container is identical except has no snapshots and
  volatiles keys are reset (e.g. MAC address):
<pre><code>lxc copy [source container] [destination container]
</code></pre>
* To rename a container:
<pre><code>lxc move [old name] [new name]
</code></pre>

# Snapshots
* Create a snapshot:
<pre><code>lxc snapshot [container] [snapshot name]
</code></pre>
* List snapshot:
<pre><code>lxc info [container]
</code></pre>
* Restore a snapshot:
<pre><code>lxc restore [container] [snapshot name]
</code></pre>
* Rename a snapshot:
<pre><code>lxc move [container]/[snapshot name] [container]/[new snapshot name]
</code></pre>
* Create a container from a snapshot. New container is identical to another
  container's snapshot except volatile info is reset and MAC address is reset:
<pre><code>lxc copy [source container]/[snapshot name] [destination container]
</code></pre>
* Delete a snapshot:
<pre><code>lxc delete [container]/[snapshot name]
</code></pre>

# [LXD 2.0: Resource control [4/12]](https://insights.ubuntu.com/2016/03/30/lxd-2-0-resource-control-412/) (2016-03)
* Limit RAM:
<pre><code>lxc config set my-container limits.memory 256MB
</code></pre>
* Turn swap off:
<pre><code>lxc config set my-container limits.memory.swap false
</code></pre>
* Turn off hard memory limits:
<pre><code>lxc config set my-container limits.memory.enforce soft
</code></pre>
* Limit to 2 CPUs:
<pre><code>lxc config set my-container limits.cpu 2
</code></pre>
* Limit to second and fourth cores:
<pre><code>lxc config set my-container limits.cpu 1,3
</code></pre>
* Limit to 10% of CPU:
<pre><code>lxc config set my-container limits.cpu.allowance 10%
</code></pre>
* Set disk limit to 20G:
<pre><code>lxc config device set my-container root size 20GB
</code></pre>
* Limit network bandwidth to 100 Mbps in and out:
<pre><code>lxc profile device set default eth0 limits.ingress 100Mbit
lxc profile device set default eth0 limits.egress 100Mbit
</code></pre>
* Set network priority:
<pre><code>lxc config set my-container limits.network.priority 5
</code></pre>

# [LXD 2.0: Image management [5/12]](https://insights.ubuntu.com/2016/04/01/lxd-2-0-image-management-512) (2016-04)
* List images with grep on alias or fingerprint:
<pre><code>lxc image list amd64
</code></pre>
* Limit images with property filter:
<pre><code>lxc image list os=ubuntu
</code></pre>
* Show image info:
<pre><code>lxc image info ubuntu
</code></pre>
* Edit an image:
<pre><code>lxc image edit [alias or fingerprint]
</code></pre>
* Create an image tarball:
<pre><code>lxc image export old-ubuntu .
</code></pre>
* Article also has instructions on __hot to create your own images__, from
  scratch or from an existing container or snapshot.

# [LXD 2.0: Remote hosts and container migration [6/12]](https://insights.ubuntu.com/2016/04/13/lxd-2-0-remote-hosts-and-container-migration-612/) (2016-04)
* Remotes can be both image sources and also where containers are run.
* Create a remote to run images on:
<pre><code>$ lxc remote add foo foo.mydomain.com
2607:f2c0:f00f:2770:216:3eff:fee1:bd67
Certificate fingerprint: fdb06d909b77a5311d7437cabb6c203374462b907f3923cefc91dd5fce8d7b60
ok (y/n)? y
Admin password for foo: 
Client certificate stored at server: foo
</code></pre>
* Run the image `14.04` from remote `ubuntu` on remote `foo` and call the container `c1`:
<pre><code>lxc launch ubuntu:14.04 foo:c1
</code></pre>
* List the containers running on the remote `foo`:
<pre><code>lxc list foo:
</code></pre>
* Create a snapshot of a container and run it on another machine:
<pre><code>lxc snapshot foo:c1 current
lxc copy foo:c1/current c3
</code></pre>
* Move a container to another machine:
<pre><code>lxc stop foo:c1
lxc move foo:c1 c1
</code></pre>

# Accessing a container:
* Get a shell inside the container:
<pre><code>$ lxc exec xenial-64 /bin/bash

root@xenial-64:~# id
uid=0(root) gid=0(root) groups=0(root)

root@xenial-64:~# uname -a
Linux xenial-64 4.4.0-22-generic #39-Ubuntu SMP Thu May 5 16:53:32 UTC 2016 x86_64 x86_64 x86_64 GNU/Linux

root@xenial-64:~# lsb_release -a
No LSB modules are available.
Distributor ID: Ubuntu
Description:    Ubuntu 16.04 LTS
Release:        16.04
Codename:       xenial
</code></pre>
* Copy file from container (file: `/tmp/foo.txt`)
<pre><code>lxc file pull xenial-64/tmp/foo.txt ~/tmp
</code></pre>
* Copy file to container:
<pre><code>lxc file push ~/tmp/tmp.txt xenial-64/tmp/foo.txt
</code></pre>

# [LXD 2.0: Docker in LXD [7/12]](https://insights.ubuntu.com/2016/04/13/stephane-graber-lxd-2-0-docker-in-lxd-712/) (2016-04)
* Multiple docker containers can be run inside a given LXD container.

# [LXD 2.0: LXD in LXD [8/12]](https://insights.ubuntu.com/2016/04/15/lxd-2-0-lxd-in-lxd-812/) (2016-04)
* LXD containers can be run inside an LXD container

# [LXD networking: lxdbr0 explained](https://insights.ubuntu.com/2016/04/07/lxd-networking-lxdbr0-explained/) (2016-04)
* Explains how to configure basic networking.

# State of LXD; Thu 2022-01-06
* wiki.debian.org [LXD](https://wiki.debian.org/LXD): "LXD...is not currently
  packaged for Debian...You may be interested in LXC instead."
* stgraber.org [LXD on Debian (using snapd)](https://stgraber.org/2017/01/18/lxd-on-debian/) (2017)
* Wikipedia [Snap (package manager)](https://en.wikipedia.org/wiki/Snap_(package_manager)):
  * "Snap is a software packaging and deployment system developed by Canonical
    for operating systems that use the Linux kernel. The packages, called
    snaps, and the tool for using them, snapd, work across a range of Linux
    distributions and allow upstream software developers to distribute their
    applications directly to users. Snaps are self-contained applications
    running in a sandbox with mediated access to the host system. Snap was
    originally released for cloud applications but was later ported to work for
    Internet of Things devices and desktop applications too."
  * "Red Hat employee Adam Williamson, while acknowledging his own bias, has
    criticized Snap for keeping the server side closed-source, not having
    a mechanism for using third party servers, and having to sign a contributor
    license agreement to contribute to its development."
    * This is referring to the Snap store, where Snap packages are installed from.
* arstechnica.com [Adios apt and yum? Ubuntu’s snap apps are coming to distros everywhere](https://arstechnica.com/information-technology/2016/06/goodbye-apt-and-yum-ubuntus-snap-apps-are-coming-to-distros-everywhere/) (2016)
* Seems worth giving this a try. The idea's interesting. An snap package is
  basically everything needed for an app packaged into a container, that has
  limited access to the host system.

# LXD on Debian; 2022-01-06
* stgraber.org [LXD on Debian (using snapd)](https://stgraber.org/2017/01/18/lxd-on-debian/) (2017)
* The requirements from the article are:
  * "A Debian “testing” (stretch) system"
  * "The stock Debian kernel without apparmor support"
  * "If you want to use ZFS with LXD, then the “contrib” repository must be
    enabled and the “zfsutils-linux” package installed on the system"
* The article was written 5 years ago. Is there anything more recent on this?
  Does Debian 11 (Buster) meet the requirements?
* snapcraft.io [Install lxd on Debian](https://snapcraft.io/install/lxd/debian) (2021)
* This requires letting Snap run as root.
* Snap is from Ubuntu.
* Install and enable snap:
<pre><code>$ apt-get install snapd
The following additional packages will be installed:
  squashfs-tools

$ snap install core
Warning: /snap/bin was not found in your $PATH. If you've not restarted your session since you installed snapd, try doing that.
core 16-2.52.1 from Canonical installed
</code></pre>
* I restarted my root session and see:
<pre><code>echo $PATH
/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin
</code></pre>
* Install LXD:
<pre><code>$ snap install lxd
lxd 4.21 from Canonical✓ installed
</code></pre>
* linuxcontainers.org [LXD­— Getting started](https://linuxcontainers.org/lxd/getting-started-cli/)
* Configure LXD, as root:
<pre><code>$ which lxd
/snap/bin/lxd

$ lxd init
WARNING: cgroup v2 is not fully supported yet, proceeding with partial confinement
Would you like to use LXD clustering? (yes/no) [default=no]:
</code></pre>
* What's this message?
* linuxquestions.org [cgroup v2 error](https://www.linuxquestions.org/questions/linux-containers-122/cgroup-v2-error-4175671573/)
* This may be a problem even with LXC, and not just LXD. Looks like the Linux kernel on Debian 11 (Buster) doesn't
  have all the cgroup support it needs to fully "confine" containers.
* The bigger concern might be that later LXC has problems because not everything it needs is supported, and doesn't work.
* github.com/lxc/lxd/issues [A Centos 7 container running on Debian 11 is unable to mount cgroup #9228](https://github.com/lxc/lxd/issues/9228)
* It looks like there's a grub option I could set to fix this.
* Try pressing ahead for now.
* Configure LXD, as root:
<pre><code>$ lxd init
WARNING: cgroup v2 is not fully supported yet, proceeding with partial confinement
Would you like to use LXD clustering? (yes/no) [default=no]: 
Do you want to configure a new storage pool? (yes/no) [default=yes]: 
Name of the new storage pool [default=default]: 
Name of the storage backend to use (dir, lvm, ceph, btrfs) [default=btrfs]: 
Create a new BTRFS pool? (yes/no) [default=yes]: 
Would you like to use an existing empty block device (e.g. a disk or partition)? (yes/no) [default=no]: 
Size in GB of the new loop device (1GB minimum) [default=30GB]: 
Would you like to connect to a MAAS server? (yes/no) [default=no]: 
Would you like to create a new local network bridge? (yes/no) [default=yes]: 
What should the new bridge be called? [default=lxdbr0]: 
What IPv4 address should be used? (CIDR subnet notation, “auto” or “none”) [default=auto]: 
What IPv6 address should be used? (CIDR subnet notation, “auto” or “none”) [default=auto]: 
Would you like the LXD server to be available over the network? (yes/no) [default=no]: 
Would you like stale cached images to be updated automatically? (yes/no) [default=yes] 
Would you like a YAML "lxd init" preseed to be printed? (yes/no) [default=no]: 
</code></pre>
* Launch a container, as root:
<pre><code>$ lxc launch images:ubuntu/20.04 foobar-220106a
WARNING: cgroup v2 is not fully supported yet, proceeding with partial confinement
Creating foobar-220106a
Starting foobar-220106a  

$ lxc list
WARNING: cgroup v2 is not fully supported yet, proceeding with partial confinement
+----------------+---------+----------------------+-----------------------------------------------+-----------+-----------+
|      NAME      |  STATE  |         IPV4         |                     IPV6                      |   TYPE    | SNAPSHOTS |
+----------------+---------+----------------------+-----------------------------------------------+-----------+-----------+
| foobar-220106a | RUNNING | 10.241.221.38 (eth0) | fd42:eab8:c639:ae34:216:3eff:fe92:edd1 (eth0) | CONTAINER | 0         |
+----------------+---------+----------------------+-----------------------------------------------+-----------+-----------+
</code></pre>
* Get shell access:
<pre><code>$ lxc exec foobar-220106a -- /bin/bash
WARNING: cgroup v2 is not fully supported yet, proceeding with partial confinement

$ root@foobar-220106a:~# whoami
root

$ root@foobar-220106a:~# pwd
/root

$ root@foobar-220106a:~# ps -ef
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 19:52 ?        00:00:00 /sbin/init
root          65       1  0 19:52 ?        00:00:00 /lib/systemd/systemd-journald
root         102       1  0 19:52 ?        00:00:00 /lib/systemd/systemd-udevd
root         108       1  0 19:52 ?        00:00:00 /usr/sbin/cron -f
message+     109       1  0 19:52 ?        00:00:00 /usr/bin/dbus-daemon --system --address=systemd: --nofork --nopidfile --systemd-activation --syslo
root         112       1  0 19:52 ?        00:00:00 /usr/bin/python3 /usr/bin/networkd-dispatcher --run-startup-triggers
syslog       113       1  0 19:52 ?        00:00:00 /usr/sbin/rsyslogd -n -iNONE
root         114       1  0 19:52 ?        00:00:00 /lib/systemd/systemd-logind
systemd+     115       1  0 19:52 ?        00:00:00 /lib/systemd/systemd-networkd
systemd+     121       1  0 19:52 ?        00:00:00 /lib/systemd/systemd-resolved
root         125       1  0 19:52 pts/0    00:00:00 /sbin/agetty -o -p -- \u --noclear --keep-baud console 115200,38400,9600 linux
root         146       0  0 19:57 pts/1    00:00:00 /bin/bash
root         153     146  0 19:57 pts/1    00:00:00 ps -ef
</code></pre>
* This is Ubuntu, and Python 3 is installed by default.
<pre><code>$ root@foobar-220106a:~# which python3
/usr/bin/python3

$ root@foobar-220106a:~# python3 --version
Python 3.8.10
</code></pre>
* Stop container:
<pre><code>$ lxc stop foobar-220106a 
WARNING: cgroup v2 is not fully supported yet, proceeding with partial confinement

$ lxc list
WARNING: cgroup v2 is not fully supported yet, proceeding with partial confinement
+----------------+---------+------+------+-----------+-----------+
|      NAME      |  STATE  | IPV4 | IPV6 |   TYPE    | SNAPSHOTS |
+----------------+---------+------+------+-----------+-----------+
| foobar-220106a | STOPPED |      |      | CONTAINER | 0         |
+----------------+---------+------+------+-----------+-----------+
</code></pre>

# LXD group; Mon 2022-01-10
* Anyone in the `lxd` group can interact with LXD. 
  See [Security and access control](https://linuxcontainers.org/lxd/getting-started-cli/#security-and-access-control).
* Add `sean` to `lxd`, and then I don't need to be root to get shell access to a container, etc.
* Restart `lxd` daemon with:
<pre><code>systemctl restart snap.lxd.daemon
</code></pre>
* I'm still unable to start a container, though, even after logging out and back in again:
<pre><code>$ /snap/bin/lxc start thudaka
Error: Get "http://unix.socket/1.0": dial unix /var/snap/lxd/common/lxd/unix.socket: connect: permission denied
</code></pre>
* I've rebooted and can now run `lxc`.
* Add snap to path, in ~/.bashrc:
<pre><code>PATH="/snap/bin:$PATH"
</code></pre>
