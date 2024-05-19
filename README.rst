*****
Chkok
*****

"chkok" checks if attributes of files and sockets match the provided conditions to ensure
system state is as expected. It can be used to monitor system health and resource availability.
Provides different running modes, useful for reporting to local and remote monitoring.
It's written in Go, has a small resource overhead with no runtime dependencies.

Main target platforms are modern GNU/Linux systems, but other operating systems may work as well.


Installation
------------

Download the latest release from the `releases` page.
Released artifacts can be verified with the checksum file containing sha256 hashes.

.. code-block:: shell

    sha256sum -c chkok-*SHA256SUMS


Or build and install from source (requires Go 1.22+):

.. code-block:: shell

    sudo make install
    # or to install in a custom location (e.g. for packaging)
    env DESTDIR=/tmp make install


Usage
-----

Run in CLI mode (default):


.. code-block:: shell

    chkok -conf examples/config.yaml


Run in HTTP mode, starting an HTTP server on the configured port:


.. code-block:: shell

    chkok -conf examples/config.yaml -verbose -mode http


The HTTP mode is useful for checking the results remotely, for example, from a monitoring system.
Currently there is no encryption supported, so it's recommended to use only in trusted networks,
or behind a local SSL terminating service.


Configuration
-------------

Configuration is done via a YAML file.

`runners` section configures how the checks should be run. The runner configurations
are merged with the `default` runner configuration.

.. code-block:: yaml

    runners:
        default:
            timeout: 5m
            # response_ok: "OK"
            # response_fail: "FAILED"
            # response_timeout: "TIMEOUT"
        cli: {}  # override default runner only for CLI mode
        http:  # override default runner only for HTTP mode
            listen_address: "127.0.0.1:51234"
            # request_read_timeout: 2s
            # response_write_timeout: 2s
            # timeout: 5s
            # max_header_bytes: 8192
            # max_concurrent_requests: 1  # 0 means no limit
            # request_required_headers:
            #   "X-Required-Header": "required-value"
            #   "X-Required-Header2": ""  # header existence is required, not value


The `check_suites` section defines the checks to be run. Each check suite
is a logical group of checks that should run sequentially.
Each check defines the expected properties of the resource (file or socket)
to be checked.

.. code-block:: yaml

    check_suites:
      etc:
        - type: dir
          path: /etc
          mode: 0755
          user: root
          group: root
        - type: file
          path: /etc/passwd
          min_size: 10
        - type: file
          path: /etc/group
          min_size: 5
          max_size: 10000
      default:
        - type: file
          path: /unwanted/file
          absent: true
        - type: dial
          network: tcp
          address: "localhost:22"
          timeout: 500ms


See the `examples` directory for sample configuration files.


Development
-----------

Make sure you have Go 1.22+ installed.
Most of the development and build tasks are automated with the `Makefile`.

To build the binary from source, run:

.. code-block:: shell

    make clean build


To run the tests and static checks, run:

.. code-block:: shell

    make test


License
-------

"chkok" is an open source project released under the terms of the `MIT license <https://opensource.org/licenses/MIT>`_.
It uses yaml.v3 library which is licensed under the MIT and Apache License 2.0 licenses.
See LICENSE file for more details.
