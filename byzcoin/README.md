Navigation: [DEDIS](https://github.com/dedis/doc/tree/master/README.md) ::
[Cothority](https://github.com/dedis/cothority/tree/master/README.md) ::
[Building Blocks](https://github.com/dedis/cothority/tree/master/doc/BuildingBlocks.md) ::
ByzCoin

# ByzCoin

This implementation of ByzCoin has its goal to implement the protocol
described in the [OmniLedger Paper](https://eprint.iacr.org/2017/406.pdf).
As the paper is only describing the network interaction and very few of the
details of how the transactions themselves are handled, we will include
them as seem fit.

This document describes the part of ByzCoin that are implemented and how to
use them. It should grow over time as more parts of the system are implemented.

## Overview

Here is a graphical overview of the current implementation in the cothority
framework:

![ByzCoin Implementation](ByzCoin.png?raw=true "ByzCoin")
As an svg: [ByzCoin Implementation](ByzCoin.svg). This image has been
created with https://draw.io and can be imported there.

Our ByzCoin service currently implements:

1. multiple transactions per block
2. queuing of transactions at each node and periodic creation of a new
block by the leader
3. contracts that define the behaviour of how to change the global state
4. view-change in case the leader fails

The following points are scheduled to be done before end of '18:

5. sharding of the nodes
6. inter-shard transactions

Items 5 and 6 are the 'real' ByzCoin improvements as described in the
[ByzCoin Paper](https://eprint.iacr.org/2017/406.pdf).

## Transaction collection and View Change

Transactions can be submitted by end-users to any conode in the roster for
the Skipchain that is holding the ByzCoin.

At the beginning of each block creation, the leader launches a protocol to
contact all the followers in parallel and to request the outstanding
transactions they have. Once a follower answers this request, they are
counting on the leader to faithfully attempt to include their transaction.
There is no retry mechanism.

With the collected transactions now in the leader, it runs them in order
to find out how many it can fit into 1/2 of a block interval. It then sends
the proposed block to the followers for them to validate. If there are transactions
remaining to be run, they will be prepended to the next collected set of
transactions when the next block interval expires.

A "view change" (change of leader) is needed when the leader stops performing
its duties correctly. Followers notice the need for a new leader if the leader
stops sending heartbeat messages within some time window or detect a malicious
behaviour (not implemented yet).

The design is similar to the view-change protocol in PBFT (OSDI99). We keep the
view-change message that followers send when they detect an anomaly. But we
replace the new-view message with the ftcosi protocol and block creation. The
result of ftcosi is an aggregate signature of all the nodes that agree to
perform the view-change. The signature is included in the block which nodes
accept if the aggregate signature is correct. This technique enables nodes to
synchronise and replay blocks to compute the most up-to-date leader.

# Structure Definitions

Following is an overview of the most important structures defined in ByzCoin.
For a more programmatic description of these structures, go to the
[DataStructures](DataStructures.md) file.

## Skipchain Block

Whenever ByzCoin stores a new Skipchain Block, the header will only contain
hashes, while the ClientTransactions will be stored in the body. This allows
for a reduced proof size.

Block header:
- Merkle tree root of the global state
- Hash of all ClientTransactions in this block
- Hash of all StateChanges resulting from the clientTransactions

Block body:
- List of all ClientTransactions

## Smart Contracts in ByzCoin

A contract defines how to interpret the methods sent by the client. It is
identified by the contractID which is a string pointing to a given contract.

Contracts receive as an input a list of coins that are available to them. As
an output, a contract needs to give the new list of coins that is available.

After all contracts have been run, the leftover coins are given to the leader as
a mining reward.

Input arguments:
- pointer to database for read-access
- Instruction from the client
- key/value pairs of coins available

Output arguments:
- one StateChange (might be empty)
- updated key/value pairs of coins still available
- error that will abort the clientTransaction if it is non-zero. No global
state will be changed if any of the contracts returns non-zero.

The contracts are compiled into the conode binary. A set of conodes making
up a cothority may have differing implementations of a given contract,
but if they do not create the same output StateChanges, the cothority might not
be able to reach the threshold of agreeing conodes in order to commit the
transactions onto the ByzCoin. If one conode is creating differing contract outputs
(for example, it is cheating), it's output will not be integrated into the
global shared state.

## From Client to the Collection

In ByzCoin we define the following path from client instructions to
global state changes:

* _Instruction_ is one of Spawn, Invoke or Delete that is called upon an
existing object
* _ClientTransaction_ is a set of instructions sent by a client
* _StateChange_ is calculated at the leader and verified by every node. It
contains the new key/contractID/value triplets to create/update/delete.

A block in ByzCoin contains zero or more ByzCoinTransactions. Every
one of these transactions can be valid or not and will be marked as such by
the leader. Every node has to verify whether it accepts or refuses the
decisions made by the leader.

### Authentication and Coins

Current authentications support darc-signatures, later authentications will also
support use of coins. It is the contracts' responsibility to verify that enough
coins are available.

## Collection

The collection is a Merkle-tree based data structure to securely and
verifiably store key / value associations on untrusted nodes. The library
in this package focuses on ease of use and flexibility, allowing to easily
develop applications ranging from simple client-server storage to fully
distributed and decentralized ledgers with minimal bootstrapping time.

Our collection used is a library that has been
[developed for a PhD project](collection/README.md) and
can do much more than simple Merkle-trees. Depending on the future direction
of the project, it might be replaced by a simpler Merkle-tree implementation.

## Darc

Package darc in most of our projects we need some kind of access control to
protect resources. Instead of having a simple password or public key for
authentication, we want to have access control that can be: evolved with a
threshold number of keys be delegated. So instead of having a fixed list of
identities that are allowed to access a resource, the goal is to have an
evolving description of who is allowed or not to access a certain resource.

For more information, see [darc/README.md](darc/README.md).

## Contracts

- [Contracts](Contracts.md) gives a short overview how contracts work and
some examples how to use them.
