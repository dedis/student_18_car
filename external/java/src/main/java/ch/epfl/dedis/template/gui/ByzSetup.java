package ch.epfl.dedis.template.gui;

import ch.epfl.dedis.byzcoin.ByzCoinRPC;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.calypso.CalypsoRPC;
import ch.epfl.dedis.calypso.LTSId;
import ch.epfl.dedis.lib.Roster;
import ch.epfl.dedis.lib.SkipblockId;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Signer;
import ch.epfl.dedis.lib.darc.SignerEd25519;
import ch.epfl.dedis.lib.proto.DarcProto;
import ch.epfl.dedis.lib.proto.OnetProto;
import ch.epfl.dedis.lib.proto.SkipchainProto;
import ch.epfl.dedis.template.gui.json.ByzC;
import ch.epfl.dedis.template.gui.json.Person;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.File;
import java.time.Duration;
import java.util.Arrays;

import static java.time.temporal.ChronoUnit.MILLIS;

public class ByzSetup {

    public static String tomlStr = "[[servers]]\n" +
            "  Address = \"tcp://127.0.0.1:7002\"\n" +
            "  Public = \"d829a0790ffa8799e4bbd1bee8da0507c9166b665660baba72dd8610fca27cc1\"\n" +
            "  Description = \"Conode_1\"\n" +
            "[[servers]]\n" +
            "  Address = \"tcp://127.0.0.1:7004\"\n" +
            "  Public = \"d750a30daa44713d1a4b44ca4ef31142b3b53c0c36a558c0d610cc4108bb4ecb\"\n" +
            "  Description = \"Conode_2\"\n" +
            "[[servers]]\n" +
            "  Address = \"tcp://127.0.0.1:7006\"\n" +
            "  Public = \"7f47f33084c3ecc233f8b05b8f408bbd1c2e4a129aae126f92becacc73576bc7\"\n" +
            "  Description = \"Conode_3\"\n" +
            "[[servers]]\n" +
            "  Address = \"tcp://127.0.0.1:7008\"\n" +
            "  Public = \"8b25f8ac70b85b2e9aa7faf65507d4f7555af1c872240305117b7659b1e58a1e\"\n" +
            "  Description = \"Conode_4\"";

    /**
     * Setting up the blockchain when the app for the demo is run.
     * Create a new Byzcoin Blockchain if it's the first time the app is started and
     * store in a local file: Roster, GenesisSkipchain and LTSId.
     * Also create and store Admin Darc and Admin in another file.
     *
     * If the configuration files already exist, read the file and return:
     *
     * @return ByzC that contains byte[] version of Roster, GenesisSkipchain and LTSId
     */
    public static ByzC setup() throws Exception{

        ByzC byzC;
        ObjectMapper mapper = new ObjectMapper();
        File byzFile = new File("/Users/Iva/byzcoin.json");
        if(!byzFile.exists()) {
            Signer genAdmin = new SignerEd25519();
            Darc genesisDarc = ByzCoinRPC.makeGenesisDarc(genAdmin, Roster.FromToml(tomlStr));

            CalypsoRPC calypso = new CalypsoRPC(Roster.FromToml(tomlStr), genesisDarc, Duration.of(500, MILLIS));
            DarcInstance genesisDarcInstance = calypso.getGenesisDarcInstance();

            byzC = new ByzC(Roster.FromToml(tomlStr).toProto().toByteArray(),
                    calypso.getGenesisBlock().getId().toProto().toByteArray(),
                    calypso.getLTSId().toProto().toByteArray());

            mapper.writeValue(new File("/Users/Iva/byzcoin.json"), byzC);


            Signer admin = new SignerEd25519();
            Darc adminDarc = new Darc(Arrays.asList(admin.getIdentity()), Arrays.asList(admin.getIdentity()), "Admin darc".getBytes());
            adminDarc.setRule("spawn:darc", admin.getIdentity().toString().getBytes());
            DarcInstance adminDarcInstance = genesisDarcInstance.spawnDarcAndWait(adminDarc, genAdmin, 10);

            Person admPerson = new Person("admin", adminDarc.toProto().toByteArray(),
                    admin.getPublic().toString(), admin.getPrivate().toString());

            File adminFile = new File("/Users/Iva/admin.json");
            mapper.writeValue(new File("/Users/Iva/admin.json"), admPerson);

            return byzC;
        }
        else
        {

            byzC = mapper.readValue(new File("/Users/Iva/byzcoin.json"), ByzC.class);
            return byzC;

        }

    }

    /**
     *
     * @param byzC that contains byte[] version of the Roster
     * @return Roster object
     * @throws Exception
     */
    public static Roster getRoster(ByzC byzC) throws Exception{
        OnetProto.Roster rosterProto = OnetProto.Roster.parseFrom(byzC.roster);
        Roster roster = new Roster(rosterProto);
        return roster;
    }

    /**
     *
     * @param byzC that contains byte[] version of the SkipblockId
     * @return SkipblockId object
     * @throws Exception
     */
    public static SkipblockId getByzId(ByzC byzC) throws Exception{
        SkipblockId ByzId  = new SkipblockId(byzC.skipblockId);
        return ByzId;
    }

    /**
     *
     * @param byzC that contains byte[] version of the LTSId
     * @return LTSId object
     * @throws Exception
     */
    public static LTSId getLTSId(ByzC byzC) throws Exception{
        LTSId ltsId  = new LTSId(byzC.ltsId);
        return ltsId;
    }

    public static Darc getAdminDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File adminFile = new File("/Users/Iva/admin.json");
        if(adminFile.exists()){
            Person admPerson = mapper.readValue(new File("/Users/Iva/admin.json"), Person.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(admPerson.darc);
            Darc darcAdmin = new Darc(darcProto);
            return  darcAdmin;
        }
        else
            return null;

    }

    public static String getAdminName() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File adminFile = new File("/Users/Iva/admin.json");
        if(adminFile.exists()){
            Person admPerson = mapper.readValue(new File("/Users/Iva/admin.json"), Person.class);
            return admPerson.name;
        }
        else
            return null;
    }


}
