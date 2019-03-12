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
    static Signer garage;

    static Darc readerDarc;
    static DarcInstance readerDarcInstance;
    static Darc garageDarc;
    static DarcInstance garageDarcInstance;
    static Darc carDarc;
    static DarcInstance carDarcInstance;

    static Darc carReaderDarc;
    static DarcInstance carReaderDarcInstance;
    static Darc carGarageDarc;
    static DarcInstance carGarageDarcInstance;
    static Darc carOwnerDarc;
    static DarcInstance carOwnerDarcInstance;


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
//        testInstanceController = TestServerInit.getInstanceManual();
        testInstanceController = TestServerInit.getInstance();

        //creating genesis darc
        genAdmin = new SignerEd25519();
        genesisDarc = ByzCoinRPC.makeGenesisDarc(genAdmin, testInstanceController.getRoster());

        calypso = new CalypsoRPC(testInstanceController.getRoster(), genesisDarc, Duration.of(1000, MILLIS));

        if (!calypso.checkLiveness()) {
            throw new CothorityCommunicationException("liveness check failed");
        }

        genesisDarcInstance = calypso.getGenesisDarcInstance();


        //Spawning admin darc with the spawn:darc rule for a new signer.
        admin = new SignerEd25519();
        adminDarc = new Darc(Arrays.asList(admin.getIdentity()), Arrays.asList(admin.getIdentity()), "Admin darc".getBytes());
        adminDarc.setRule("spawn:darc", admin.getIdentity().toString().getBytes());
        adminDarcInstance = genesisDarcInstance.spawnDarcAndWait(adminDarc, genAdmin, 10);


        //Spawning user darc with invoke:evolve and _sign rules
        user = new SignerEd25519();
        userDarc = new Darc(Arrays.asList(user.getIdentity()), Arrays.asList(user.getIdentity()), "User darc".getBytes());
        userDarcInstance = adminDarcInstance.spawnDarcAndWait(userDarc, admin, 10);


        carOwnerDarc = new Darc(Arrays.asList(userDarc.getIdentity()),
                Arrays.asList(userDarc.getIdentity()), "Car Owner darc".getBytes());
        carOwnerDarcInstance = adminDarcInstance.spawnDarcAndWait(carOwnerDarc, admin, 10);


        //Spawning reader darc with invoke:evolve and _sign rules
        carReaderDarc = new Darc(Arrays.asList(carOwnerDarc.getIdentity()),
                Arrays.asList(carOwnerDarc.getIdentity()), "Car Reader darc".getBytes());
        carReaderDarcInstance = adminDarcInstance.spawnDarcAndWait(carReaderDarc, admin, 10);


        //Spawning garage darc with invoke:evolve and _sign rules
        carGarageDarc = new Darc(Arrays.asList(carOwnerDarc.getIdentity()),
                Arrays.asList(carOwnerDarc.getIdentity()), " Car Garage darc".getBytes());
        carGarageDarcInstance = adminDarcInstance.spawnDarcAndWait(carGarageDarc, admin, 10);


        //Spawning car darc with spawn:car, invoke:addReport, spawn:calypsoWrite and spawn:calypsoRead rules
        Rules rs = new Rules();
        rs.addRule("spawn:car", adminDarc.getIdentity().toString().getBytes());
        rs.addRule("invoke:addReport", carGarageDarc.getIdentity().toString().getBytes());
        rs.addRule("spawn:calypsoWrite", carGarageDarc.getIdentity().toString().getBytes());
        rs.addRule("spawn:calypsoRead", carReaderDarc.getIdentity().toString().getBytes());
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
        assertTrue(calypso.checkLiveness());
    }


    /**
     * Spawns a new car Instance
     *
     * @throws Exception
     */
    @Test
    void spawnCar() throws Exception {

        Car c = new Car("123A456");
        CarInstance ci = new CarInstance(calypso, carDarcInstance, admin, c);

        Car c2 = new Car(ci.getInstance().getData());
        assertEquals(c, c2);
    }


    /**
     * Spawns a new car Instance
     * Evolves the car garage Darc by adding a new Garage
     * Evolves the car instance by adding a report
     * Reads the report
     *
     * @throws Exception
     */
    @Test
    void spawnCarAddAndReadReport() throws Exception {

        //spawn car
        Car c = new Car("123A46");
        CarInstance ci = new CarInstance(calypso, carDarcInstance, admin, c);
        Car c2 = new Car(ci.getInstance().getData());
        assertEquals(c, c2);

        //spawn garage darc
        garage = new SignerEd25519();
        garageDarc = new Darc(Arrays.asList(garage.getIdentity()),
                Arrays.asList(garage.getIdentity()), "Garage darc".getBytes());
        garageDarcInstance = adminDarcInstance.spawnDarcAndWait(garageDarc, admin, 10);

        // evolve the car garage Darc
        carGarageDarc.addIdentity(Darc.RuleSignature, garageDarc.getIdentity(), Rules.OR);
        carGarageDarcInstance.evolveDarcAndWait(carGarageDarc, user, 10);

        //update car instance by adding a report
        SecretData secret = new SecretData("1090", "100 000", true, "tires changed");
        Document doc = new Document(secret.toProto().toByteArray(), 16, "sdf".getBytes(), carDarc.getBaseId());
        WriteInstance wi = new WriteInstance(calypso,
                carDarc.getBaseId(), Arrays.asList(user), doc.getWriteData(calypso.getLTS()));

        Proof p = calypso.getProof(wi.getInstance().getId());
        assertTrue(p.matches());

        List<Report> reports = new ArrayList<>();
        Report report = new Report("15.02.1994", "1234523", wi.getInstance().getId().getId());
        reports.add(report);

        ci.addReportAndWait(reports, user, 10);
        assertEquals(report, ci.getReports().get(0));

        //read the report
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
     * Evolves the car owner darc
     *
     * @throws Exception
     */
    @Test
    void changeOwner() throws Exception {

        //spawn car instance
        Car c = new Car("123A46");
        CarInstance ci = new CarInstance(calypso, carDarcInstance, admin, c);
        Car c2 = new Car(ci.getInstance().getData());
        assertEquals(c, c2);

        Signer newOwner = new SignerEd25519();
        Darc newOwnerDarc = new Darc(Arrays.asList(newOwner.getIdentity()),
                Arrays.asList(newOwner.getIdentity()), "New Owner darc".getBytes());
        carOwnerDarc.setRule("_sign", newOwnerDarc.getIdentity().toString().getBytes());
        carOwnerDarc.setRule("invoke:evolve", newOwnerDarc.getIdentity().toString().getBytes());
        carOwnerDarcInstance.evolveDarcAndWait(carOwnerDarc, user, 10);

    }

}
