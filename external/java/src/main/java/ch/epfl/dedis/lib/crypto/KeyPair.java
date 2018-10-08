package ch.epfl.dedis.lib.crypto;

import java.security.SecureRandom;

public class KeyPair {
    public Scalar scalar;
    public Point point;

    public KeyPair() {
        byte[] seed = new byte[Ed25519.field.getb() / 8];
        new SecureRandom().nextBytes(seed);
        scalar = new Ed25519Scalar(seed);
        point = Ed25519Point.base().mul(scalar);
    }
}
