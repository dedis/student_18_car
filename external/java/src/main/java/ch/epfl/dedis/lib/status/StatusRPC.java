package ch.epfl.dedis.lib.status;

import ch.epfl.dedis.lib.ServerIdentity;
import ch.epfl.dedis.lib.exception.CothorityCommunicationException;
import ch.epfl.dedis.proto.OnetProto;
import ch.epfl.dedis.proto.StatusProto;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;

import java.util.Map;

/**
 * The RPC to get the status of a conode.
 */
public class StatusRPC {
    /**
     * Make an RPC to the conode identified by its server identity (sid) to get its status.
     * @param sid The server identity of the conode.
     */
    public static Map<String, OnetProto.Status> getStatus(ServerIdentity sid) throws CothorityCommunicationException {
        StatusProto.Request.Builder b = StatusProto.Request.newBuilder();
        ByteString msg = ByteString.copyFrom(sid.SendMessage("Status/Request", b.build().toByteArray()));

        try {
            StatusProto.Response resp = StatusProto.Response.parseFrom(msg);
            return resp.getStatusMap();
        } catch (InvalidProtocolBufferException e) {
            throw new CothorityCommunicationException(e);
        }
    }
}
