Navigation: [DEDIS](https://github.com/dedis/doc/tree/master/README.md) ::
[Cothority](../README.md) ::
[Applications](../doc/Applications.md) ::
Calypso

# Calypso

Calypso is the implementation of the upcoming "Calypso - Auditable Sharing of
Private Data over Blockchains". The paper can be found
[here](https://eprint.iacr.org/2018/209).

In short, Calypso allows to store symmetric keys in ByzCoin, protected by a
sharded key, and controls access to this symmetric keys using Darcs,
Distributed Access Rights Control.

It implements both the access-control cothority and the secret-management
cothority:
- The access-control cothority is implemented using ByzCoin with two
  contracts, `calypsoWrite` and `calypsoRead`
- The secret-management cothority uses an onet service with methods to set up a
  Long Term Secret (LTS) distributed key and to request a re-encryption

The workflow is the following:
1. secret-management: Administrator sets up a new LTS for all his clients. It
   does so by calling the `CreateLTS` service endpoint. The resulting `LTSID`
   will be used by all clients.
2. access-control: Administrator gives document creation rights to a writer
3. access-control: Writer creates new Darcs for customers and for documents.
4. access-control: Writer spawns a `Write` instance from a document Darc
5. access-control: Reader requests that a `Read` instance is spawned from a
   `Write` instance
6. secret-management: Reader requests a re-encryption to the `DecryptKey`
   service endpoint.

![Workflow Overview](CalypsoByzCoin.png?raw=true "Workflow Overview")

## Darcs, Instances, Instructions and Contracts

Here is a very short overview of the three most important elements of
ByzCoin. For a more thorough documentation, refer to
[ByzCoin](../byzcoin/README.md) documentation.

The current ByzCoin service is a batching implementation of the previous
skipchain service. It has a global state that holds _Instances_, where every
instance is tied to a _Contract_ and holds a blob of data. The contract defines
how the data is to be interpreted and allows different _Instructions_ sent from
the user.

Access control is done using _Darcs_, which define what public keys can verify
an action. Each instruction received by ByzCoin is mapped to an action and
then verified if the given signature is correct. Also, every instance is linked
to one darc that defines what actions are allowed to be done to that instance.

All instructions sent to ByzCoin are batched in a new block that is created
every `blockInterval` seconds.

## CreateLTS

The CreateLTS endpoint is currently unsecured. It takes a list of nodes as
input and then asks all these nodes to create a distributed key using DKG. For
this operation, all nodes must be online. Per default, a threshold of 2/3 of
the nodes must be present for the decryption.

The CreateLTS service endpoint returns a `LTSID` in the form of a 32 byte
slice. This ID represents the group that created the distributed key. Any node
can participate in as many DKGs as you want and will get a random `LTSID`
assigned.

## Write Contract

The write contract verifies that the request has been correctly created, so
that no malicious writer can send an encrypted key without knowing the secret.
It then creates a new write-instance that contains the write request.

A read request must also be sent to the write contract, which will forward it
to the read contract. This is so that every instruction sent to ByzCoin has
as a target an existing instance.

## Read Contract

The read contract verifies that the request is valid and points to the write
instance. It stores the reader's public key in the instance, so that the
secret-management cothority can re-encrypt to this reader's public key.
