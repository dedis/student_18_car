package ch.epfl.dedis.template;

import ch.epfl.dedis.integration.TestServerController;
import ch.epfl.dedis.integration.TestServerInit;
import ch.epfl.dedis.lib.Roster;
import ch.epfl.dedis.lib.SkipblockId;
import ch.epfl.dedis.byzcoin.ByzCoinRPC;
import ch.epfl.dedis.calypso.*;
import ch.epfl.dedis.byzcoin.Proof;
import ch.epfl.dedis.lib.darc.*;
import ch.epfl.dedis.lib.exception.CothorityCommunicationException;
import ch.epfl.dedis.lib.exception.CothorityException;
import ch.epfl.dedis.byzcoin.InstanceId;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import com.google.protobuf.InvalidProtocolBufferException;
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
    static ByzCoinRPC bc;
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
        System.out.println(genesisDarc.toString());

        bc = new ByzCoinRPC(testInstanceController.getRoster(), genesisDarc, Duration.of(500, MILLIS));
        if (!bc.checkLiveness()) {
            throw new CothorityCommunicationException("liveness check failed");
        }

        //bc.update();

        genesisDarcInstance = bc.getGenesisDarcInstance();

        //genesisDarcInstance = DarcInstance.fromByzCoin(bc, genesisDarc);
        //genesisDarcInstance = DarcInstance.fromByzCoin(bc, genesisDarc.getId());
        //genesisDarcInstance = new DarcInstance(bc, genesisDarc, genAdmin, Darc newDarc)




        //Spawning admin darc with the spawn:darc rule for a new signer.
        admin = new SignerEd25519();
        adminDarc = new Darc(Arrays.asList(admin.getIdentity()), Arrays.asList(admin.getIdentity()), "Admin darc".getBytes());
        adminDarc.setRule("spawn:darc", admin.getIdentity().toString().getBytes());
        adminDarcInstance = genesisDarcInstance.spawnDarcAndWait(adminDarc, genAdmin, 10);




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
        assertTrue(bc.checkLiveness());
    }

    /**
     * Evolves the darc to give spawn-rights to create a keyValue contract, as well as the right to invoke the
     * update command from the contract.
     * Then it will store a first key/value pair and verify it's correctly stored.
     * Finally it updates the key/value pair to a new value.
     *
     * @throws Exception
     */
 /*  @Test
    void spawnAndUpdateCar() throws Exception {

       //spawn
       Car c = new Car("123A456");
       CarInstance ci = new CarInstance(bc, carDarcInstance, admin, c);
       assertEquals(c, ci);

       //update
       String secret = "this is a secret";
       Document doc = new Document(secret.getBytes(), 16, null, genesisDarc.getBaseId());
       WriteInstance wi = new WriteInstance(calypso,
               userDarc.getBaseId(), Arrays.asList(user), doc.getWriteData(calypso.getLTS()));
       Proof p = calypso.getProof(wi.getInstance().getId());
       assertTrue(p.matches());

       List<Report> reports = new ArrayList<>();
       Report report = new Report("15.02.1994", "1234523", wi.getInstance().getId().getId());
       reports.add(report);

       ci.addReportAndWait(reports, user, 10);
       assertEquals(report, ci.getReports().get(0));

       //mKV.setValue("27".getBytes());
       //vi.updateKeyValueAndWait(Arrays.asList(mKV), admin, 10);

       //assertEquals(mKV, vi.getKeyValues().get(0));
    }*/


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
