package ch.epfl.dedis.template.gui.json;

import ch.epfl.dedis.calypso.LTSId;
import ch.epfl.dedis.lib.Roster;
import ch.epfl.dedis.lib.SkipBlock;
import ch.epfl.dedis.lib.SkipblockId;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Signer;
import ch.epfl.dedis.lib.proto.DarcProto;
import ch.epfl.dedis.lib.proto.OnetProto;

public class ByzC {
//    public Roster roster;
//    public SkipBlock skipBlock;
//    public SkipblockId skipblockId;
//    public LTSId ltsId;
//    public Darc genDarc;
//    public byte[] genAdminIdentity;
//
//    public ByzC(Roster roster, SkipBlock skipBlock){
//        this.roster = roster;
//        this.skipBlock = skipBlock;
//        this.skipblockId = skipblockId;
//        this.ltsId = ltsId;
//        this.genDarc = genDarc;
//        this.genAdminIdentity = genAdminIdentity;
//    }
    public byte[] roster;
    public byte[] skipblockId;
    public byte[] ltsId;


    public ByzC( byte[] roster, byte[] skipblock, byte[] ltsId){
        this.roster = roster;
        this.skipblockId = skipblock;
        this.ltsId = ltsId;
    }

    public ByzC() {
    }
}
