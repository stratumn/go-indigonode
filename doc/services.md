# Services

## Introduction

A service is a fundamental concept. Almost everything is implemented as a
service. In fact, the core of Alice is essentially a service manager. Without
any services running, it doesn't do much!

A service runs until it is no longer needed and adds functionality to an Alice
node. For instance, the CLI (Command Line Interface) connects to a node via its
gRPC API. If the gRPC API service is stopped, it is no longer possible to
control the node from the CLI.

From the CLI prompt, you can list the available services using the
`manager-list` command.

```
Alice> manager-list
ID       STATUS   NEEDS            STOPPABLE  PRUNABLE  NAME             DESC
api      RUNNING  [grpcapi]        false      false     API Services     Starts API services.
boot     RUNNING  [api system]     true       false     Boot Services    Starts boot services.
grpcapi  RUNNING  [manager]        false      true      gRPC API         Exposes a gRPC API.
manager  RUNNING  []               false      true      Service Manager  Manages services.
pruner   RUNNING  [manager]        false      true      Service Pruner   Prunes unused services.
signal   RUNNING  [manager]        false      true      Signal Handler   Handles exit signals.
system   RUNNING  [pruner signal]  false      true      System Services  Starts system services.
```

There are low-level system services, such as `signal`, which takes care of
shutting down the node if requested. Services can be more complex, such as
a chat application. Other services, such as `boot`, do nothing on their own,
and exist solely to ensure other services are started.

A service can depend on other services, as you can see in the `NEEDS` column
back in the CLI. When you start a service that needs other services, Alice
starts all of them. If a service is not needed by any other services, it is
said to be `STOPPABLE`. A service is said to be `PRUNABLE` if it wasn't
started directly but rather because it was needed by another service. Every
now and then, services that are both `PRUNABLE` and `STOPPABLE` are stopped
in an effort to save resources. The `pruner`, which is itself a service,
takes care of stopping unused services.

You can prevent a service from being pruned by starting it explicitly, even
if it is already running. Using the CLI, let's make sure the gRPC API service
never gets pruned (don't worry, it should never happen under normal
circumstances).

```
Alice> manager-start grpcapi
ID         grpcapi
STATUS     RUNNING
NEEDS      [manager]
STOPPABLE  false
PRUNABLE   false
NAME       gRPC API
DESC       Exposes a gRPC API.
```

The gRPC API service is no longer `PRUNABLE`. In the spirit of minimalism,
let's stop everything but the gRPC API services.

```
Alice> manager-stop boot
ID         boot
STATUS     STOPPED
NEEDS      [api system]
STOPPABLE  true
PRUNABLE   false
NAME       Boot Services
DESC       Starts boot services.

Alice> manager-prune
ID       STATUS   NEEDS            STOPPABLE  PRUNABLE  NAME             DESC
api      RUNNING  [grpcapi]        true       false     API Services     Starts API services.
boot     STOPPED  [api system]     true       false     Boot Services    Starts boot services.
grpcapi  RUNNING  [manager]        false      false     gRPC API         Exposes a gRPC API.
manager  RUNNING  []               false      true      Service Manager  Manages services.
pruner   STOPPED  [manager]        true       true      Service Pruner   Prunes unused services.
signal   STOPPED  [manager]        true       true      Signal Handler   Handles exit signals.
system   STOPPED  [pruner signal]  true       true      System Services  Starts system services.
```

What happened there? We stopped the `boot` service. It is a very special
service that tells Alice what to start when it boots. It does so by declaring
that it "needs" other services. Since we just shut it down, nothing except the
services required by the gRPC API is needed anymore. Good thing we made sure
the gRPC API service wouldn't be stopped! We also told Alice to immediately
prune the services so they would be removed right away.

You can turn the services back on by starting the `boot` service again.

```
Alice> manager-start boot
ID         boot
STATUS     RUNNING
NEEDS      [api system]
STOPPABLE  true
PRUNABLE   false
NAME       Boot Services
DESC       Starts boot services.

Alice> manager-list
ID       STATUS   NEEDS            STOPPABLE  PRUNABLE  NAME             DESC
api      RUNNING  [grpcapi]        false      false     API Services     Starts API services.
boot     RUNNING  [api system]     true       false     Boot Services    Starts boot services.
grpcapi  RUNNING  [manager]        false      false     gRPC API         Exposes a gRPC API.
manager  RUNNING  []               false      true      Service Manager  Manages services.
pruner   RUNNING  [manager]        false      true      Service Pruner   Prunes unused services.
signal   RUNNING  [manager]        false      true      Signal Handler   Handles exit signals.
system   RUNNING  [pruner signal]  false      true      System Services  Starts system services.
```
