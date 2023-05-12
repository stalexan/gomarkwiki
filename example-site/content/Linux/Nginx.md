# Nginx
* Wikipedia [Nginx](https://en.wikipedia.org/wiki/Nginx)
* Website: <https://www.nginx.com/>
* Or there's this site too which looks more bare bones and may be for just Nginx and not Nginx Plus: <https://nginx.com/>

# Installing; 2022-08
* [Installing NGINX Open Source](https://docs.nginx.com/nginx/admin-guide/installing-nginx/installing-nginx-open-source/)

# Deploying With Docker; 2022-08-03
* docs.nginx.com [Deploying NGINX and NGINX Plus on Docker](https://docs.nginx.com/nginx/admin-guide/installing-nginx/installing-nginx-docker/)
* The [offical image](https://hub.docker.com/_/nginx) is based on `debian:bullseye-slim`.
* Run a simple image:
<pre><code>$ docker run --name nginx-container -p 22080:80 --detach --rm nginx
</code></pre>
* Additional packages to install to be able to experiment, view files, processes, etc:
  * `procps`: For `ps`
  * `vim`: For `view`

# Web Server Configuration; 2022-08
* docs.nginx.com [Configuring NGINX and NGINX Plus as a Web Server](https://docs.nginx.com/nginx/admin-guide/web-server/web-server/)
* docs.nginx.com [Creating NGINX Plus and NGINX Configuration Files](https://docs.nginx.com/nginx/admin-guide/basic-functionality/managing-configuration-files/)
* Default config file is `/etc/nginx/nginx.conf`
* The default location for HTML is `/usr/share/nginx/html`.
* docs.nginx.com [Serving Static Content](https://docs.nginx.com/nginx/admin-guide/web-server/serving-static-content/)
* Seems the best way to configure Nginx and make HTML files available is to copy a custom `/etc/nginx/nginx.conf` to
  the image, to not affect any of the other default settings in `/etc/nginx`, and then use a volume for the HTML, mounting
  `/home/sean/nginx/html` to `/usr/share/nginx/html`.

# Default Config Files; 2022-08-04
* `/etc/nginx/nginx.conf`:
<pre><code>user  nginx;
worker_processes  auto;

error_log  /var/log/nginx/error.log notice;
pid        /var/run/nginx.pid;

events {
    worker_connections  1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    #gzip  on;

    include /etc/nginx/conf.d/*.conf;
}
</code></pre>
* `/etc/nginx/conf/default.conf`:
<pre><code>server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    #access_log  /var/log/nginx/host.access.log  main;

    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
    }

    #error_page  404              /404.html;

    # redirect server error pages to the static page /50x.html
    #
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }

    # proxy the PHP scripts to Apache listening on 127.0.0.1:80
    #
    #location ~ \.php$ {
    #    proxy_pass   http://127.0.0.1;
    #}

    # pass the PHP scripts to FastCGI server listening on 127.0.0.1:9000
    #
    #location ~ \.php$ {
    #    root           html;
    #    fastcgi_pass   127.0.0.1:9000;
    #    fastcgi_index  index.php;
    #    fastcgi_param  SCRIPT_FILENAME  /scripts$fastcgi_script_name;
    #    include        fastcgi_params;
    #}

    # deny access to .htaccess files, if Apache's document root
    # concurs with nginx's one
    #
    #location ~ /\.ht {
    #    deny  all;
    #}
}
</code></pre>

# SSL
* docs.nginx.com [NGINX SSL Termination](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/)
* Configuration after testing with [SSL Labs](https://www.ssllabs.com/):
<pre><code>server {
    listen                      443 ssl;
    server_name                 www.alexan.org;
    ssl_certificate             /etc/letsencrypt/live/www.alexan.org/fullchain.pem;
    ssl_certificate_key         /etc/letsencrypt/live/www.alexan.org/privkey.pem;
    ssl_protocols               TLSv1.2;
    ssl_prefer_server_ciphers   on;
    ssl_ciphers                 "EECDH+ECDSA+AESGCM EECDH+aRSA+AESGCM EECDH+ECDSA+SHA384 EECDH+ECDSA+SHA256 EECDH+aRSA+SHA384 EECDH+aRSA+SHA256 EECDH+aRSA+RC4 EECDH EDH+aRSA RC4 !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS";
    ...
}
</code></pre>

# Basic HTTP Authentication
* docs.nginx.com [Restricting Access with HTTP Basic Authentication](https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-http-basic-authentication/)
* [Module ngx_http_auth_basic_module](https://nginx.org/en/docs/http/ngx_http_auth_basic_module.html)
* Apache's `htpasswd` can be used to create a password. Install:
<pre><code>apt-get install apache2-utils
</code></pre>
* Create passwords. The `-c` options creates the password file and only needs to be used once:
<pre><code>$ htpasswd -c /etc/apache2/.htpasswd user1
$ htpasswd /etc/apache2/.htpasswd user2
</code></pre>

# Reverse Proxy for weight-log App
* docs.nginx.com [NGINX Reverse Proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/)

# Don't Cache PDFs; 2023-04-23
* Tell browser not to cache PDFs, so user isn't shown a stale copy, in `nginx.conf`:
```
# Have browser always request latest version of pdfs.
location ~ \.pdf$ {
    try_files $uri $uri/ =404;
    add_header Cache-Control "max-age=0, no-cache, no-store, must-revalidate";
}

location / {
    try_files $uri $uri/ =404;
}
```
* Only one location block applies to each request. The more specific location block is used, which
  in this case is the pdf block, for pdf files.
