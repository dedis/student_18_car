package ch.epfl.dedis.template;

import ch.epfl.dedis.integration.TestServerController;
import ch.epfl.dedis.integration.TestServerInit;
import ch.epfl.dedis.byzcoin.ByzCoinRPC;
import ch.epfl.dedis.calypso.*;
import ch.epfl.dedis.byzcoin.Proof;
import ch.epfl.dedis.lib.darc.*;
import ch.epfl.dedis.lib.exception.CothorityCommunicationException;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Duration;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import static java.time.temporal.ChronoUnit.MILLIS;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

public class CarTest {
    //static ByzCoinRPC bc;
    static CalypsoRPC calypso;

    static Signer genAdmin;
    static Darc genesisDarc;
    static DarcInstance genesisDarcInstance;

    static Signer admin;
    static Darc adminDarc;
    static DarcInstance adminDarcInstance;
    static Signer user;
    static Darc userDarc;
    static DarcInstance userDarcInstance;

    static Darc readerDarc;
    static DarcInstance readerDarcInstance;
    static Darc garageDarc;
    static DarcInstance garageDarcInstance;
    static Darc carDarc;
    static DarcInstance carDarcInstance;

    private final static Logger logger = LoggerFactory.getLogger(CarTest.class);
    private TestServerController testInstanceController;

    /**
     * Initializes a new ByzCoin ledger and adds a genesis darc with evolve rights to the admin.
     * The new ledger is empty and will create new blocks every 500ms, which is good for tests,
     * but in a real implementation would be more like 5s.
     *
     * @throws Exception
     */
    @BeforeEach
    void initAll() throws Exception {
        testInstanceController = TestServerInit.getInstance();

        //creating genesis darc
        genAdmin = new SignerEd25519();
        genesisDarc = ByzCoinRPC.makeGenesisDarc(genAdmin, testInstanceController.getRoster());


        //bc = new ByzCoinRPC(testInstanceController.getRoster(), genesisDarc, Duration.of(500, MILLIS));
        calypso = new CalypsoRPC(testInstanceController.getRoster(), genesisDarc, Duration.of(500, MILLIS));

        if (!calypso.checkLiveness()) {
            throw new CothorityCommunicationException("liveness check failed");
        }
        System.out.println(genesisDarc.toString());
        //bc.update();

        genesisDarcInstance = calypso.getGenesisDarcInstance();

        //genesisDarcInstance = DarcInstance.fromByzCoin(bc, genesisDarc);
        //genesisDarcInstance = DarcInstance.fromByzCoin(bc, genesisDarc.getId());
        //genesisDarcInstance = new DarcInstance(bc, genesisDarc, genAdmin, Darc newDarc)




        //Spawning admin darc with the spawn:darc rule for a new signer.
        admin = new SignerEd25519();
        adminDarc = new Darc(Arrays.asList(admin.getIdentity()), Arrays.asList(admin.getIdentity()), "Admin darc".getBytes());
        adminDarc.setRule("spawn:darc", admin.getIdentity().toString().getBytes());
        adminDarcInstance = genesisDarcInstance.spawnDarcAndWait(adminDarc, genAdmin, 10);


        //Main.main();

        //Spawning user darc with invoke:evolve and _sign rules
        user = new SignerEd25519();
        userDarc = new Darc(Arrays.asList(user.getIdentity()), Arrays.asList(user.getIdentity()), "User darc".getBytes());
        userDarcInstance = adminDarcInstance.spawnDarcAndWait(userDarc, admin, 10);



        //Spawning reader darc with invoke:evolve and _sign rules
        readerDarc = new Darc(Arrays.asList(userDarc.getIdentity()), Arrays.asList(userDarc.getIdentity()), "Reader darc".getBytes());
        readerDarcInstance = adminDarcInstance.spawnDarcAndWait(readerDarc, admin, 10);



        //Spawning garage darc with invoke:evolve and _sign rules
        garageDarc = new Darc(Arrays.asList(userDarc.getIdentity()), Arrays.asList(userDarc.getIdentity()), "Garage darc".getBytes());
        garageDarcInstance = adminDarcInstance.spawnDarcAndWait(garageDarc, admin, 10);



        //Spawning car darc with spawn:car, invoke:addReport, spawn:calypsoWrite and spawn:calypsoRead rules
        Rules rs = new Rules();
        rs.addRule("spawn:car", adminDarc.getIdentity().toString().getBytes());
        rs.addRule("invoke:addReport", garageDarc.getIdentity().toString().getBytes());
        rs.addRule("spawn:calypsoWrite", garageDarc.getIdentity().toString().getBytes());
        rs.addRule("spawn:calypsoRead", readerDarc.getIdentity().toString().getBytes());
        carDarc = new Darc(rs, "Car darc".getBytes());
        carDarcInstance = adminDarcInstance.spawnDarcAndWait(carDarc, admin, 10);

        System.out.println(carDarcInstance.getDarc().toString());
    }

    /**
     * Simply checks the liveness of the conodes. Can often catch a badly set up system.
     *
     * @throws Exception
     */
    @Test
    void ping() throws Exception {
        assertTrue(calypso.checkLiveness());
    }

    /**
     * Evolves the darc to give spawn-rights to create a keyValue contract, as well as the right to invoke the
     * update command from the contract.
     * Then it will store a first key/value pair and verify it's correctly stored.
     * Finally it updates the key/value pair to a new value.
     *
     * @throws Exception
//     */
//    @Test
//    void addReader() throws Exception {
//
//        Signer reader = new SignerEd25519();
//        readerDarc.addIdentity("_sign", reader.getIdentity(), Rules.AND);
//        calypso.getGenesisDarcInstance().evolveDarcAndWait(readerDarc, admin, 10);
//        System.out.println(readerDarc.toString());
//
//    }



    /**
     * Evolves the darc to give spawn-rights to create a keyValue contract, as well as the right to invoke the
     * update command from the contract.
     * Then it will store a first key/value pair and verify it's correctly stored.
     * Finally it updates the key/value pair to a new value.
     *
     * @throws Exception
     */
   @Test
    void spawnCar() throws Exception {

       //spawn
       Car c = new Car("123A456");
       CarInstance ci = new CarInstance(calypso, carDarcInstance, admin, c);
       System.out.println("Car Instance:");
       System.out.println(ci.getInstance().getId().toString());
       Car c2 = new Car (ci.getInstance().getData());
       assertEquals(c, c2);
    }

    @Test
    void spawnCarAddAndReadReport() throws Exception {

        //spawn
        Car c = new Car("123A46");
        CarInstance ci = new CarInstance(calypso, carDarcInstance, admin, c);
        System.out.println("Car Instance:");
        System.out.println(ci.getInstance().getId().toString());
        Car c2 = new Car (ci.getInstance().getData());
        assertEquals(c, c2);

        //update
        SecretData secret = new SecretData("1090", "100 000", true, "tires changed");
        Document doc = new Document(secret.toProto().toByteArray(), 16, "sdf".getBytes(), carDarc.getBaseId());
        //WriteData wd = new WriteData(calypso.getLTS(), secret.getBytes(),
        //      new Encryption.keyIv(16).getKeyMaterial() , null,carDarc.getBaseId());
        WriteInstance wi = new WriteInstance(calypso,
                carDarc.getBaseId(), Arrays.asList(user), doc.getWriteData(calypso.getLTS()));
        Proof p = calypso.getProof(wi.getInstance().getId());
        assertTrue(p.matches());

        List<Report> reports = new ArrayList<>();
        Report report = new Report("15.02.1994", "1234523", wi.getInstance().getId().getId());
        reports.add(report);

        ci.addReportAndWait(reports, user, 10);
        assertEquals(report, ci.getReports().get(0));

        ReadInstance ri = new ReadInstance(calypso, wi, Arrays.asList(user));

        DecryptKeyReply dkr = calypso.tryDecrypt(calypso.getProof(wi.getInstance().getId()), calypso.getProof(ri.getInstance().getId()));
        // And derive the symmetric key, using the user's private key to decrypt it:
        byte[] keyMaterial = dkr.getKeyMaterial(user.getPrivate());

        // Finally get the document back:
        Document doc2 = Document.fromWriteInstance(wi, keyMaterial);

        SecretData s2 = new SecretData(doc2.getData());
        //assertEquals(secret.getEcoScore(), s2.getEcoScore());

        // And check it's the same.
        assertTrue(doc.equals(doc2));

    }


    /**
     * We only give the client the roster and the genesis ID. It should be able to find the configuration, latest block
     * and the genesis darc.
     */
/*    @Test
    void reconnect() throws Exception {
        KeyValue mKV = new KeyValue("value", "314159".getBytes());
        KeyValueInstance vi = new KeyValueInstance(bc, genesisDarcInstance, admin, Arrays.asList(mKV));
        assertEquals(mKV, vi.getKeyValues().get(0));

        reconnect_client(bc.getRoster(), bc.getGenesis().getSkipchainId(), vi.getId());
    }*/

    /**
     * Re-connects to a ByzCoin ledger and verifies the value stored in the keyValue instance. This shows
     * how to use the minimal information necessary to get the data from an instance.
     *
     * @param ro   the roster of ByzCoin
     * @param scId the Id of ByzCoin
     * @param kvId the Id of the instance to retrieve
     */
   /* void reconnect_client(Roster ro, SkipblockId scId, InstanceId kvId) throws CothorityException, InvalidProtocolBufferException {
        ByzCoinRPC bc = new ByzCoinRPC(ro, scId);
        assertTrue(bc.checkLiveness());

        KeyValueInstance localKvi = new KeyValueInstance(bc, kvId);
        KeyValue testKv = new KeyValue("value", "314159".getBytes());
        assertEquals(testKv, localKvi.getKeyValues().get(0));
    }*/
}
