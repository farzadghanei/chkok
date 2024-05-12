=====
chkok
=====

-------------------------------
Check if file attributes are OK
-------------------------------

:Author: Farzad Ghanei
:Date:   2024-05-11
:Copyright:  Copyright (c) 2024 Farzad Ghanei. chkok is an open source project released under the terms of the MIT license.
:Version: 0.3.0
:Manual section: 1
:Manual group: General Command Manuals


SYNOPSIS
========
    chkok [OPTIONS]


DESCRIPTION
===========
chkok checks if attributes of files and sockets match the provided conditions to ensure
system state is as expected. It can be used to monitor system health and resource availability.
Provides different running modes, useful for reporting to local and remote monitoring.

OPTIONS
=======

.. code-block::

  -conf string
        path to configuration file in YAML format (default "/etc/chkok.yaml")
  -mode string
        running mode: cli,http (default "cli")
  -verbose
        more output, include logs


CONFIGURATION
=============

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


FILES
=====

**\/etc\/chkok.yaml**
    The default configuration file, if available should contain valid configuration in YAML format.


REPORTING BUGS
==============
Bugs can be reported with https://github.com/farzadghanei/chkok/issues
