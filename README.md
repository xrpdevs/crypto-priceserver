# **priceserver v0.1**

Simple price scraper with HTTP server for use with prometheus

Currently working with Bitrue.com exchange but easily adaptable to others.

**Configuration:**

`make`

`make install`

This will build and install the application to the system, along with the included default config files.

You may then need to edit /etc/priceserver.yml - examples are provided inside the file.

Finally, add priceserver as a prometheus target, by adding the following to the bottom of /etc/prometheus/prometheus.yml

` - job_name: 'priceserver'`

`scrape_interval: 10s`

`scrape_timeout: 9s`

`    static_configs:`

`      - targets: ['127.0.0.1:7071']`

Target ip and port needs to be modified to match what you have in /etc/priceserver

(c) 2021 Jamie Prince / flareftso.com / xrpdevs.co.uk

GPLv2 license
