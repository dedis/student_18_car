package ch.epfl.dedis.lib;

import ch.epfl.dedis.lib.crypto.Hex;
import ch.epfl.dedis.lib.exception.CothorityCryptoException;
import com.google.protobuf.ByteString;

import java.util.Arrays;

/**
 * Implementation of {@link HashId}. This implementation is immutable and is can be used as key for collections
 */
public class Sha256id implements HashId {
    private final byte[] id;
    public final static int length = 32;

    public Sha256id(ByteString bs) throws CothorityCryptoException{
        this(bs.toByteArray());
    }

    public Sha256id(byte[] id) throws CothorityCryptoException {
        if (id.length != length) {
            throw new CothorityCryptoException("need 32 bytes for sha256-hash, only got " + id.length);
        }
        this.id = Arrays.copyOf(id, id.length);
    }

    @Override
    public byte[] getId() {
        return Arrays.copyOf(id, id.length);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        return Arrays.equals(id, ((Sha256id) o).id);
    }

    @Override
    public int hashCode() {
        return Arrays.hashCode(id);
    }

    @Override
    public String toString(){
        return Hex.printHexBinary(id);
    }

    public ByteString toProto(){
        return ByteString.copyFrom(id);
    }
}
