Description
===========

`docker-env` is a deployment tool that can be used for creating
_Docker environments_ with the help of [docker-machine](https://docs.docker.com/machine/).
You can specify your _environment_ as multiple machine definitions
in a YAML file, and then `docker-env` will create all your Docker
hosts for you, installing Docker on them.

Example scenario
----------------

For example, you could define a OpenStack production environment with
this YAML file:

```YAML
# docker-env.yml
driver:
  openstack:
    flavor-name:        tiny
    image-name:         Ubuntu 14.04 LTS
    floatingip-pool:    myfloatingips
machines:
  master:
    instances:          1
    swarm:
      master:           true
      discovery:        token://12345
  worker-$(#):
    instances:          4
    swarm:
      discovery:        token://12345
```

Then you could create the environment with:

```
$ docker-env create
```

... and five machines would be created (with Docker installed) in your OpenStack cluster.
Then you could configure your Docker client for talking to the Docker Swarm
master installed at `master` with:

```
$ eval "$(docker-machine env --swarm master)"
```

You could also define your application with a `docker-compose.yml` file, so
running `docker-compose up` would install and run the application in your
_OpenStack_ cluster, managed by Swarm. So your code repository would
typically contain the following YAMLs:

```
$ ls *.yml
docker-compose.yml    docker-env.yml    docker-env-production.yml    docker-env-qa.yml
```

Status
------

This is a work in progress and it cannot be considered production-ready yet.
I have used it only with VirtualBox, so all the other drivers can just
not work at all...

Installation
============

The easiest way to install `docker-env` is by running it with Docker in
your local machine. It will be automaticaly pulled from the
[Docker Hub build](https://hub.docker.com/r/inercia/docker-env/) with:

```
docker run --rm inercia/docker-env --help
```

Otherwise, you can checkout this repository and `make deps prog/docker-env`
and obtain the `docker-env` binary in `prog/docker-env`


Configuration files
===================

The global sections are:

* `vars`
* `auth`
* `engine`
* `driver`
* `swarm`
* `machines`

And machines definitions can have the following subsections:

* `auth`
* `engine`
* `driver`
* `swarm`

When machines do not specify one of these sections, they will
be automatically copied from the global sections. For example:

```YAML
# docker-env.yml
driver:
  openstack:
    flavor-name:        tiny
    image-name:         Ubuntu 14.04 LTS
    floatingip-pool:    myfloatingips
machines:
  master:
    instances:          1
    swarm:
      master:           true
```

The `master` machine will use the `openstack` driver defined in
the global section, as if it were defined with a subsection.


Variables
---------

You can define variables in a global `vars` section. For example:

```YAML
# docker-env.yml
vars:
  NUM_DATABASES:    1
  SWARM_DISCOVERY:  token://12345
...
```

Some rules apply on variables:

* variables are replaced _after_ all configration files have been merged.
* variables can be used in any _string_ and _string list_ fields, as well
as the number instances for a machine.
* the special variable `$(#)` is replaced in a machine definition by the
number of instance. For example:
```YAML
vars:
  NUM_DATABASES:    3
  SWARM_DISCOVERY:  token://12345
machines:
	database-$(#):
	  instances:    $(NUM_DATABASES)
...
```
* you can define (or replace) variables at the command line with the `-X`
flag. For example, you could use a fresh discovery token with:
```
$ docker-env create -X SWARM_DISCOVERY=token://$(swarm create)
```


Files modularity
----------------

A machines definition can split in several files. This is specially
useful for having some common definition in a `docker-env.yml`
file and then use different files for diferent environments like
_production_, _integration_, etc...

All of the file names must start with `docker-env`. For example,
you can use a base `docker-env.yml` file like this:

```YAML
# docker-env.yml
machines:
	master:
	  instances:     1
	  swarm:
	    master:      true
	database-$(#):
	  instances:     $(NUM_DATABASES)
	  engine:
	    label:       class.database
	worker-$(#):
	  instances:     $(NUM_WORKERS)
	  driver:
	    softlayer:
	      cpu:       4
	      disk-size: 100
	      memory:    8192
	      region:    [ tor01, dal05, sjc01 ]
```

and then you can specify a _OpenStack_ configuration
in a `docker-env-openstack.yml` file:

```YAML
# docker-env-openstack.yml
vars:
  NUM_DATABASES:      4
  NUM_WORKERS:        10
engine:
  environment:
    MY_APP_ENV:       production
driver:
  openstack:
    flavor-name:      tiny
    image-name:       Ubuntu 14.04 LTS
    floatingip-pool:  myfloatingips
swarm:
  discovery:          token://1234
```

and the equivalent for testing in local VirtualBox'es

```YAML
# docker-env-virtualbox.yml
engine:
  env:
    MY_APP_ENV:       development
driver:
  virtualbox:
    cpu-count:        1
    memory:           1024
    no-share:         true
swarm:
  master:             true
  discovery:          token://1234
```


Creating the machines
---------------------

So running `docker-env create openstack` will merge the contents of
`docker-machine.ymĺ` and `docker-env-openstack.yml` and produce:

```YAML
# docker-machine.ymĺ + docker-env-openstack.yml
machines:
	master:
	  instances:            1
	  engine:
	    env:
	      APP_ENV:          production
	  driver:
	    openstack:
	      flavor-name:      tiny
	      image-name:       Ubuntu 14.04 LTS
	      floatingip-pool:  myfloatingips
	  swarm:
	    master:             true
	    discovery:          token://1234
	database-$(#):
	  instances:            4
	  engine:
	    label:              class.database
	    env:
	      APP_ENV:          production
	  driver:
	    openstack:
	      flavor-name:      tiny
	      image-name:       Ubuntu 14.04 LTS
	      floatingip-pool:  myfloatingips
	  swarm:
	    discovery:          token://1234
	worker-$(#):
	  instances:            10
	  engine:
	    env:
	      APP_ENV:          production
	  driver:
	    softlayer:
	      cpu:              4
	      disk-size:        100
	      memory:           8192
	      region:           [ tor01, dal05, sjc01 ]
	  swarm:
	    discovery:          token://1234
```

while running a simple `docker-env create` would produce an incomplete
configuration file, as it would use just `docker-env.ymĺ` (where 
nothing about the driver has been specified).
