package ch.epfl.dedis.template;

import ch.epfl.dedis.integration.TestServerController;
import ch.epfl.dedis.integration.TestServerInit;
import ch.epfl.dedis.lib.Roster;
import ch.epfl.dedis.lib.SkipblockId;
import ch.epfl.dedis.byzcoin.ByzCoinRPC;
import ch.epfl.dedis.lib.exception.CothorityCommunicationException;
import ch.epfl.dedis.lib.exception.CothorityException;
import ch.epfl.dedis.byzcoin.InstanceId;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Rules;
import ch.epfl.dedis.lib.darc.Signer;
import ch.epfl.dedis.lib.darc.SignerEd25519;
import com.google.protobuf.InvalidProtocolBufferException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Duration;
import java.util.Arrays;

import static java.time.temporal.ChronoUnit.MILLIS;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

public class CarTest {
    static ByzCoinRPC bc;

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

    private final static Logger logger = LoggerFactory.getLogger(KeyValueTest.class);
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
        Rules rules = Darc.initRules(Arrays.asList(genAdmin.getIdentity()),
                Arrays.asList(genAdmin.getIdentity()));

        genesisDarc = ByzCoinRPC.makeGenesisDarc(genAdmin, testInstanceController.getRoster());
        //genesisDarc = new Darc(rules, "genesis".getBytes());

        bc = new ByzCoinRPC(testInstanceController.getRoster(), genesisDarc, Duration.of(500, MILLIS));
        if (!bc.checkLiveness()) {
            throw new CothorityCommunicationException("liveness check failed");
        }
        genesisDarcInstance = DarcInstance.fromByzCoin(bc, genesisDarc);

        // Show how to evolve a darc to add new rules. We could've also create a correct genesis darc in the
        // lines above by adding all rules. But for testing purposes this shows how to add new rules to a darc.
        /*Darc darc2 = genesisDarc.copy();
        darc2.setRule("spawn:keyValue", admin.getIdentity().toString().getBytes());
        darc2.setRule("invoke:update", admin.getIdentity().toString().getBytes());
        genesisDarcInstance.evolveDarcAndWait(darc2, genAdmin);*/

        //Spawning admin darc with the spawn:darc rule for a new signer.
        Signer admin = new SignerEd25519();
        adminDarc = new Darc(Arrays.asList(admin.getIdentity()), Arrays.asList(admin.getIdentity()), "Admin darc".getBytes());
        adminDarc.setRule("spawn:darc", admin.getIdentity().toString().getBytes());
        adminDarcInstance = genesisDarcInstance.spawnDarcAndWait(adminDarc, genAdmin, 10);
        //not working --> ERROR: cannot find symbol
        //[ERROR]   symbol:   method spawnDarcAndWait(ch.epfl.dedis.lib.byzcoin.darc.Darc,ch.epfl.dedis.lib.byzcoin.darc.Signer,int)

        //adminDarcInstance = new DarcInstance(bc, adminDarc);

        //Spawning user darc with invoke:evolve and _sign rules
        user = new SignerEd25519();
        userDarc = new Darc(Arrays.asList(user.getIdentity()), Arrays.asList(user.getIdentity()), "User darc".getBytes());
        userDarcInstance = adminDarcInstance.spawnDarcAndWait(userDarc, admin, 10);
        //userDarcInstance = new DarcInstance(bc, userDarc);

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
/*    @Test
    void spawnValue() throws Exception {
        KeyValue mKV = new KeyValue("value", "314159".getBytes());

        KeyValueInstance vi = new KeyValueInstance(bc, genesisDarcInstance, admin, Arrays.asList(mKV));
        assertEquals(mKV, vi.getKeyValues().get(0));

        mKV.setValue("27".getBytes());
        vi.updateKeyValueAndWait(Arrays.asList(mKV), admin, 10);

        assertEquals(mKV, vi.getKeyValues().get(0));
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
