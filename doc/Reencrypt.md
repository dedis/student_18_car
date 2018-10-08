Navigation: [DEDIS](https://github.com/dedis/doc) ::
[Cothority](https://github.com/dedis/cothority) ::
[Building Blocks](https://github.com/dedis/cothority/tree/documentation/doc/BuildingBlocks.md) ::
Distributed Reencryption

# Distributed Reencryption

Once a [DKG](../evoting/protocol/DKG.md) has been set up, its aggregated public
key can be used to encrypt data, for example using ElGamal encryption. In some
circumstances you don't want to directly decrypt that data, but merely give
access to another user, without the distributed setup seeing what the original
data is.

We call this _re-encryption_, because it takes encrypted data and outputs
an encrypted blob that can be decrypted by another private key than the one
used in the DKG. This is done by having each node decrypting the data with
his share of the key, and then encrypting it to the new key. As each no only
has a share of the key, the original data is never revealed. However, the end
result is encrypted to a new public key and can be decrypted using the corresponding
private key.

We use this _re-encryption_ in our onchain-secrets implementation that will
soon be added to the cothority.

## Research Papers

- [SCARAB](https://eprint.iacr.org/2018/209) - Hidden in Plain Sight: Storing
and Managing Secrets on a Public Ledger
