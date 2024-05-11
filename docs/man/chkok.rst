=====
chkok
=====

---------------------------------
Checks if system resources are OK
---------------------------------

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
chkok checks if system resources are OK, and provides a report to demonstrate system
health and resource availablity.

OPTIONS
=======

.. code-block::

  -conf string
        path to configuration file in YAML format (default "/etc/chkok.yaml")
  -mode string
        running mode: cli,http (default "cli")
  -verbose
        more output, include logs


FILES
=====

**\/etc\/chkok.yaml**
    The default configuration file, if available should contain valid configuration in YAML format.


REPORTING BUGS
==============
Bugs can be reported with https://github.com/farzadghanei/chkok/issues
