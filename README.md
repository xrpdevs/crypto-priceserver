# **priceserver v0.3**

Simple price scraper with HTTP server/exporter for use with Prometheus

Currently working with Bitrue.com exchange but easily adaptable to others.

**Configuration:**

`make`

`make install`

This will build and install the application to the system, along with the included default config files.

You may then need to edit /etc/priceserver.yml - examples are provided inside the file.

Finally, add priceserver as a prometheus target, by adding the the contents of "add_to_prometheus.yml" to the end of /etc/prometheus/prometheus.yml

Target IP address and port need to be modified to match what you have in /etc/priceserver.yml - usually the defaults will work just fine as long as you're running this exporter and prometheus on the same machine.

(c) 2021 Jamie Prince / flareftso.com / xrpdevs.co.uk

GPLv2 license
