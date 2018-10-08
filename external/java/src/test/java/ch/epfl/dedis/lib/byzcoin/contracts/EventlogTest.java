package ch.epfl.dedis.lib.byzcoin.contracts;

import ch.epfl.dedis.integration.TestServerController;
import ch.epfl.dedis.integration.TestServerInit;
import ch.epfl.dedis.lib.byzcoin.ByzCoinRPC;
import ch.epfl.dedis.lib.byzcoin.InstanceId;
import ch.epfl.dedis.lib.byzcoin.darc.Darc;
import ch.epfl.dedis.lib.byzcoin.darc.Rules;
import ch.epfl.dedis.lib.byzcoin.darc.Signer;
import ch.epfl.dedis.lib.byzcoin.darc.SignerEd25519;
import ch.epfl.dedis.lib.eventlog.Event;
import ch.epfl.dedis.lib.eventlog.SearchResponse;
import ch.epfl.dedis.lib.exception.CothorityCommunicationException;
import ch.epfl.dedis.lib.exception.CothorityCryptoException;
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

class EventlogTest {
    private static ByzCoinRPC bc;
    private static EventLogInstance el;
    private static Signer admin;

    private final static Logger logger = LoggerFactory.getLogger(EventlogTest.class);
    private TestServerController testInstanceController;

    @BeforeEach
    void initAll() throws Exception {
        testInstanceController = TestServerInit.getInstance();
        admin = new SignerEd25519();
        Rules rules = Darc.initRules(Arrays.asList(admin.getIdentity()),
                Arrays.asList(admin.getIdentity()));
        rules.addRule("spawn:eventlog", admin.getIdentity().toString().getBytes());
        rules.addRule("invoke:eventlog", admin.getIdentity().toString().getBytes());
        Darc genesisDarc = new Darc(rules, "genesis".getBytes());

        bc = new ByzCoinRPC(testInstanceController.getRoster(), genesisDarc, Duration.of(500, MILLIS));
        if (!bc.checkLiveness()) {
            throw new CothorityCommunicationException("liveness check failed");
        }

        el = new EventLogInstance(bc, Arrays.asList(admin), genesisDarc.getId());
    }

    @Test
    void log() throws Exception {
        Event e = new Event("hello", "goodbye");
        InstanceId key = el.log(e, bc.getGenesisDarc().getBaseId(), Arrays.asList(admin));
        Thread.sleep(5 * bc.getConfig().getBlockInterval().toMillis());
        Event loggedEvent = el.get(key);
        assertEquals(loggedEvent, e);
    }

    @Test
    void testLogMore() throws Exception {
        int n = 50;
        List<InstanceId> keys = new ArrayList<>(n);
        Event event = new Event("login", "alice");
        for (int i = 0; i < n; i++) {
            // The timestamps in these event are all the same, but doing el.log takes time and it may not be possible to
            // add all the events. So we have to limit the number of events that we add.
            keys.add(el.log(event, bc.getGenesisDarc().getBaseId(), Arrays.asList(admin)));
        }
        boolean allOK = true;
        for (int i = 0; i < 4; i++) {
            allOK = true;
            Thread.sleep(5 * bc.getConfig().getBlockInterval().toMillis());
            for (InstanceId key : keys) {
                try {
                    logger.info("ok");
                    Event event2 = el.get(key);
                    assertEquals(event, event2);
                } catch (CothorityCryptoException e){
                    logger.info("bad");
                    allOK = false;
                    break;
                }
            }
            if (allOK){
                break;
            }
        }
        assertTrue(allOK, "one of the events failed");
    }

    @Test
    void testSearch() throws Exception {
        long now = System.currentTimeMillis() * 1000 * 1000;
        Event event = new Event(now, "login", "alice");
        el.log(event, bc.getGenesisDarc().getBaseId(), Arrays.asList(admin));

        Thread.sleep(5 * bc.getConfig().getBlockInterval().toMillis());

        // finds the event under any topic
        SearchResponse resp = el.search("", now - 1000, now + 1000);
        assertEquals(1, resp.events.size());
        assertEquals(resp.events.get(0), event);
        assertTrue(!resp.truncated);

        // finds the event under the right topic
        resp = el.search("login", now - 1000, now + 1000);
        assertEquals(1, resp.events.size());
        assertEquals(resp.events.get(0), event);
        assertTrue(!resp.truncated);

        // event does not exist
        resp = el.search("", now - 2000, now - 1000);
        assertEquals(0, resp.events.size());
    }
}
