# Why we need Redis Streans

## Our need of an event bus

Cluster Registry is responsible for registering clusters to sentinel control plane. It exposes an HTTP API which can be used to register 
clusters using kubeconfig, clusterName and context.

However, this information about cluster is needed across the services of control plane. Each service is decoupled and has its own storage layer.
To propagate the information about newly registered clusters to all services, we use Redis Streams.

Similarly, agent controller, the controller that is responsible for installing sentinel agents on the cluster and ensuring they are healthy. This
also needs information about cluster, and when it has cluster name it evaluates the status of the cluster. It evaluates cluster reachability, agent
installation status etc. This information is also needed across multiple services. An example is policy registry which allows users to push their service
policy to control plane needs the information about cluster and agent. In policy registry it is called cluster runtime status.
To propagate this information across services, we again use Redis Streams.

## Why Redis Streams

We know that across services we produce an event, and consumers consume it. But we need a stable bus that ensures atmoic + successful read and write or pub and sub.
If the consumer crashes by half reading, it should re-read it. If two consumers read the same log, they need co-ordination for who is responsible, who retries on failure and
who owns which event. And if consumer is slow, we need to avoid spinning and batch efficiently.

Redis Streams provides us with a simple and efficient way to implement this event bus. 
There are three levels of event bus

##### Level 1: The Log

A bus adds ordered and persistent log. It is append only. By log, I mean events here.

##### Level 2: Consumer Groups

Consumer groups do XREAD stream 0, which gives them all the messages, without co-ordination and safety.



### Why Sentinel Requires Level 3

Sentinel services are controllers, not log readers.

They:

* Materialize state from events
* Maintain system invariants
* Must converge to the correct state even after crashes
* Using Redis Streams with consumer groups ensures:
* Events are processed exactly once per service responsibility
* Failures do not corrupt or lose state
* Services can scale horizontally without coordination logic
* New services can subscribe without impacting existing ones