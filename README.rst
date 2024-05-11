*****
Chkok
*****

"chkok" checks if system resources are OK, and provides a report to demonstrate
system health and resource availablity.
It's written in Go, has a small resource overhead with no runtime dependencies.

Usage
-----

Run in CLI mode (default):

```
chkok -conf examples/config.yaml
```

Run in HTTP mode, starting a server on configured port:

```
chkok -conf examples/config.yaml -verbose -mode http
```

Configuration
-------------

Configuration is done via a YAML file. See the `examples` directory for sample configuration files.


License
-------

"chkok" is an open source project released under the terms of the `MIT license <https://opensource.org/licenses/MIT>`_.
It uses yaml.v3 library which is licensed under the MIT and Apache License 2.0 licenses.
See LICENSE file for more details.