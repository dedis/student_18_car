package ch.epfl.dedis.byzgen;

import java.security.PublicKey;
import java.security.Signature;

public interface UserSignature {
    /**
     * Return signature body
     * @return signature body
     * @see Signature#sign()
     */
    byte [] getSignature();

    /**
     * Return type of algorithm used for creating signature.
     * <a href="http://docs.oracle.com/javase/8/docs/technotes/guides/security/StandardNames.html#Signature">
     *     security/StandardNames/Signature</a>}
     *
     * @return algorithm used for signature
     */
    String getAlgorithm();

    /**
     * Returns public key used for signing document
     * @return public key used for signing document
     */
    PublicKey getPublicKey();
}
