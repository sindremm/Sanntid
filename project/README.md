# Elevator project

## Units
###  Master-slave
The master-slave works by having every elevator being a slave while at all times, one of the elevators are assigned master.

The responsiblities of master: 
- Collecitng calls
- Assigning orders to elevators
- Broadcast message containing all the info of the system to all the slaves

The broadcasted message contains a counter to keep track of the latest state of the system. 

The slaves:
- Continually send messages to master to signify that they are alive and receiving orders
- Listen to master signal and update internal info

#### Reassigning master:
The master is reassigned using the raft consensus algorithm, where the elevator with the highest counter (and then priority) is elected as the new master. The master is reassigned in two cases:
1. The master dies
    - The elevator with the next highest priority takes over
2. An elevator of higher priority than the current master appears
    - The new elevator sends signal to master. The master recognizes that the elevator is of higher priority, sends data, and reassigns master.


###  Scheduler



## Layout
- All received orders 
    - int buttons [4] bool
    - up bottons [4] bool
    - down buttons [4] bool
- Is Master bool
- Working elevators [4] bool
- Elevator states[4] elevator states (current floor, target floor)
- Counter [int] 