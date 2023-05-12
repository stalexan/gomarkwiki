# Let's Encrypt
* <https://letsencrypt.org/>
* Wikipedia [Let's Encrypt](https://en.wikipedia.org/wiki/Let%27s_Encrypt)
* tracker.debian.org [python-letsencrypt](https://tracker.debian.org/pkg/python-letsencrypt)
* [Documentation](https://letsencrypt.readthedocs.org/en/latest/intro.html)

# Ready for use on Debian Stable? (2016-02)
* Can Let's Encrypt be used on Debian Stable (jessie)?
* letsencrypt.org [How It Works](https://letsencrypt.org/howitworks/)
* Looks like it should work.

# Renew Certificate; 2020-11-11
* Unblock port 80 in `/etc/iptables-init`
* Is it also necessary to enable the HTTP site 000-default in Apache? It seems
  like it may be, and that certbot needs this. Create a `/var/www/html` directory and then:
<pre><code>a2ensite 000-default
</code></pre>
* Stop Apache:
<pre><code>systemctl stop apache2.service
</code></pre>
* Renew with:
<pre><code>certbot renew
</code></pre>
* Then reverse previous steps by blocking port 80 in `/etc/iptables-init`, and:
<pre><code>a2dissite 000-default
systemctl start apache2.service
</code></pre>
* Also, __reblock port 80__:
<pre><code>vi /etc/iptables-init
/etc/iptables-init
iptables -L -v
</code></pre>

# Renewal After Switch to Nginx; Fri 2022-08-12
* I'm now running Nginx in a Docker container on elias, instead of Apache.
* It looks like I need some part of Apache at least, though. I get this error
  when trying to renew:
<pre><code>$ certbot renew
...
Cert is due for renewal, auto-renewing...
Could not choose appropriate plugin: The requested apache plugin does not appear to be installed
Failed to renew certificate www.alexan.org with error: The requested apache plugin does not appear to be installed
</code></pre>
* letsencrypt.org [Getting Started](https://letsencrypt.org/getting-started/):
  "We recommend that most people with shell access use the Certbot ACME
  client."
* eff.org [Certbot](https://certbot.eff.org/):
  [certbot instructions: Nginx on Debian 10](https://certbot.eff.org/instructions?ws=nginx&os=debianbuster)
* I think ideally I'd run the renewal from inside the Nginx container. That's the web server
  I'm running on elias now.
* Although, the certificates are on the host. I could reinstall Apache just for cert renewal.
* It looks like there's a standalone web server I should be able to use, though:
<pre><code>$ certbot plugins
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
&ast; standalone
Description: Spin up a temporary webserver
Interfaces: IAuthenticator, IPlugin
Entry point: standalone = certbot.&UnderBar;internal.plugins.standalone:Authenticator

&ast; webroot
Description: Place files in webroot directory
Interfaces: IAuthenticator, IPlugin
Entry point: webroot = certbot.&UnderBar;internal.plugins.webroot:Authenticator
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
</code></pre>
* This worked, as root:
<pre><code>certbot --standalone renew
</code></pre>
* Bring Nginx down and then back up first. Complete set of steps, as root:
<pre><code>elias-nginx-down
certbot --standalone renew
elias-nginx-up
</code></pre>
