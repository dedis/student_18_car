package ch.epfl.dedis.lib.byzcoin.darc;

import ch.epfl.dedis.lib.Sha256id;
import ch.epfl.dedis.lib.exception.CothorityCryptoException;
import com.google.protobuf.ByteString;

/**
 * This class represents a DarcId, which is the hash of the fixed fields.
 */
public class DarcId extends Sha256id {
    /**
     * Constructs a darc ID from a byte array.
     * @param id the darc ID
     * @throws CothorityCryptoException
     */
    public DarcId(byte[] id) throws CothorityCryptoException {
        super(id);
    }

    /**
     * Constructs a darc ID from ByteString.
     * @param id the darc ID
     * @throws CothorityCryptoException
     */
    public DarcId(ByteString id) throws CothorityCryptoException {
        this(id.toByteArray());
    }

    /**
     * Creates a darc ID with all zeros.
     * @return the darc ID
     * @throws CothorityCryptoException
     */
    public static DarcId zero() throws CothorityCryptoException {
        return new DarcId(new byte[32]);
    }
}
