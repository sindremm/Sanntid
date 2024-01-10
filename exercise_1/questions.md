Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> Parallelism is when the process is happening at the same time, and concurrency is when they switch between threads and the threads do not act at the same time. 

What is the difference between a *race condition* and a *data race*? 
> A data race is an error that happens when a thread tries to read from or write to a memory location that is already being written to by another thread.
A race condition is when the outcome of an operation depends on the timing of two (or more) different threads. 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> A scheduler is the thing that decides what thread to run next. It usually selects among the runnable threads. One way to do it is to pick a random one. 


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> It is good to use when you want to have concurrent executions. You can only use the interactions that you need to have clearer code. 

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> Coroutines are more practical when you need a lightweight solution, as they bypass the OS's thread management system by only having to communicate inbetween themselves.


Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> Creating concurrent programs makes solving problems of independent tasks easier, but adds more overhead to the program overall. The inclusion of threads can also introduce race conditions
and timing errors, which will have to be taken into account by the programmer.

What do you think is best - *shared variables* or *message passing*?
> Sharing variables is easier to implement than message passing, as it is only necessary to keep track of which thread is 'locking' a resource at the time. When threads are sharing variables however,
the resource is unavailable to other threads for longer than with message passing. 


