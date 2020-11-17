# hello-dependencies

This is an adaptation of the [CNCF buildpacks samples repo](https://github.com/buildpacks/samples)
to demonstrate a buildpack with dependencies.

# Build this buildpack

This buildpack can be built (from the root of the repo) with:

```shell
pack package-buildpack my-buildpack --config ./package.toml
```

# Use this buildpack

```shell
# This uses the v0.0.1 release, to use a locally build image (e.g. my-buildpack)
# replace the ghcr.io reference below.
pack build blah --buildpack ghcr.io/mattmoor/hello-dependencies:v0.0.1
```

# Sample output

```
===> DETECTING
hello-buildpacks 0.0.1
hello-dependent  0.0.1
===> ANALYZING
===> RESTORING
===> BUILDING
---> Buildpack (dependent) Template
     platform_dir files:
       /platform:
       total 12
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 .
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 ..
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 env

       /platform/env:
       total 20
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 .
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 ..
       -rw-r--r-- 1 root root    4 Jan  1  1980 VAR1
       -rw-r--r-- 1 root root    4 Jan  1  1980 VAR2
       -rw-r--r-- 1 root root    4 Jan  1  1980 VAR3
     env_dir: /platform/env
     env vars:
       HOSTNAME=b45dca7b365d
       CNB_STACK_ID=io.buildpacks.stacks.bionic
       PWD=/workspace
       HOME=/home/cnb
       CNB_BUILDPACK_DIR=/cnb/buildpacks/hello-dependent/0.0.1
       VAR1=val1
       VAR3=val3
       VAR2=val2
       SHLVL=1
       PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
       _=/usr/bin/env
     layers_dir: /layers/hello-dependent
     plan_path: /tmp/plan.793582662/hello-dependent/plan.toml
     plan contents:
       [[entries]]
         name = "some-other-thing"

---> Done
---> Buildpack Template
     platform_dir files:
       /platform:
       total 12
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 .
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 ..
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 env

       /platform/env:
       total 20
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 .
       drwxr-xr-x 1 root root 4096 Nov 17 06:04 ..
       -rw-r--r-- 1 root root    4 Jan  1  1980 VAR1
       -rw-r--r-- 1 root root    4 Jan  1  1980 VAR2
       -rw-r--r-- 1 root root    4 Jan  1  1980 VAR3
     env_dir: /platform/env
     env vars:
       HOSTNAME=b45dca7b365d
       CNB_STACK_ID=io.buildpacks.stacks.bionic
       PWD=/workspace
       HOME=/home/cnb
       CNB_BUILDPACK_DIR=/cnb/buildpacks/hello-buildpacks/0.0.1
       VAR1=val1
       VAR3=val3
       VAR2=val2
       SHLVL=1
       PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
       BAR=foo
       _=/usr/bin/env
     layers_dir: /layers/hello-buildpacks
     plan_path: /tmp/plan.793582662/hello-buildpacks/plan.toml
     plan contents:
       [[entries]]
         name = "some-thing"

---> Done
===> EXPORTING
Reusing 1/1 app layer(s)
Reusing layer 'launcher'
Adding layer 'config'
Adding label 'io.buildpacks.lifecycle.metadata'
Adding label 'io.buildpacks.build.metadata'
Adding label 'io.buildpacks.project.metadata'
Warning: default process type 'web' not present in list []
*** Images (7f5ae43d13af):
      blah
Successfully built image blah
```
