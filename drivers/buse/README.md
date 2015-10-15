## What is BUSE?
BUSE is a block driver in user space.  It leverages `NBD` to provide block volume access to a container.  In the back, it writes data out to a local file.

## Roadmap
The next rev of BUSE will include replication to a host based on the cluster API.  It will log transactions to a local database and synchronously replicate blocks to a peer BUSE driver.

With this version of BUSE you should be able to use Docker in a multi host - shared nothing storage cluster.
