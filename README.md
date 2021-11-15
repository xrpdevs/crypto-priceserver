# **priceserver v0.1**

Simple price scraper with HTTP server for use with prometheus

Currently supporting SGB, more flexibility to be added in the near future

**Configuration:**

`make`

`make install`

You then need to edit /etc/priceserver.conf - examples are provided inside the file.

Finally, add priceserver as a prometheus target, by adding the following to the bottom of /etc/prometheus/prometheus.yml

` - job_name: 'priceserver'`

`scrape_interval: 10s`

`scrape_timeout: 9s`

`    static_configs:`

`      - targets: ['192.168.7.101:7071']`

Target needs to be modified to match what you have in /etc/priceserver

(c) 2021 Jamie Prince / flareftso.com / xrpdevs

GPLv2 license
