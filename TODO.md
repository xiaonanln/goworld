# TODO

* Run GoWorld server on Docker

* Adopt Kafka
    ** Support For Reliable Call?

* Run processes in Docker
    * Dispatcher, Gate, Game should run in different docker container 
    * Processes connect each other and other services using Container Network
    * Processes discover each other using etcd ?

* Optimize callall and 'AllClients' attribute broadcasting ?

* Multiple service entities on multiple Games to remove SPOF in service architecture

* Service Registry using Etcd

* Better AOI algorithm that enables entities to have different AOI distances

* Read config using tag (maybe use yaml)
