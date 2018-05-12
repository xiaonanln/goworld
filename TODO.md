# TODO

* Service Architecture Change

* Global Information Management (broadcast global info to all games)

* Global Message Mechanism (call multiple entities on multiple games)

* Use Kafka for reliable messaging between entities

* Fault Recovery
    * Fault Recovery of Gates 
    * Fault Recovery of Dispatchers
        * Restoring connections from games & gates
        * Restoring client infos from gates 
        * Restoring game & entity infos from games
    * Fault Recovery of Games 
        * Clear entity infos on dispatchers
        * Restoring Services
            * Restore informations on Services (using Redis?)
            
* Stateless Server (making server as stateless as possible)
