package ch.epfl.dedis.lib.calypso;

import ch.epfl.dedis.lib.exception.CothorityException;
import ch.epfl.dedis.lib.exception.CothorityNotFoundException;
import ch.epfl.dedis.lib.byzcoin.*;
import ch.epfl.dedis.lib.byzcoin.darc.DarcId;
import ch.epfl.dedis.lib.byzcoin.darc.Signer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Arrays;
import java.util.List;

/**
 * ReaderInstance represents an instance created by the calypsoRead contract.
 */
public class ReaderInstance {
    private static String ContractId = "calypsoRead";
    private Instance instance;
    private ByzCoinRPC bc;
    private final static Logger logger = LoggerFactory.getLogger(ReaderInstance.class);

    /**
     * Constructor used for when a new instance is needed.
     *
     * @param bc      The ByzCoinRPC object.
     * @param signers Signers who are allowed to spawn this instance.
     * @param darcId  The darc ID that has the signers.
     * @param rr      The ReadRequest that is sent to the contract.
     * @throws CothorityException
     */
    public ReaderInstance(ByzCoinRPC bc, List<Signer> signers, DarcId darcId, ReadRequest rr) throws CothorityException {
        this.bc = bc;
        InstanceId id = this.read(rr, darcId, signers);
        this.setInstance(id);
    }

    /**
     * Constructor used to connect to an existing instance.
     *
     * @param bc The ByzCoinRPC object.
     * @param id The identity of the instance.
     * @throws CothorityException
     */
    public ReaderInstance(ByzCoinRPC bc, InstanceId id) throws CothorityException {
        this.bc = bc;
        this.setInstance(id);
    }

    /**
     * Get the instance object.
     */
    public Instance getInstance() {
        return instance;
    }

    /**
     * Create a spawn instruction with a read request and send it to the ledger.
     */
    private InstanceId read(ReadRequest rr, DarcId darcID, List<Signer> signers) throws CothorityException {
        Argument arg = new Argument("read", rr.toProto().toByteArray());

        Spawn spawn = new Spawn(ContractId, Arrays.asList(arg));
        Instruction instr = new Instruction(new InstanceId(darcID.getId()), Instruction.genNonce(), 0, 1, spawn);
        instr.signBy(darcID, signers);

        ClientTransaction tx = new ClientTransaction(Arrays.asList(instr));
        bc.sendTransactionAndWait(tx, 5);

        return instr.deriveId("");
    }

    // TODO same as what's in EventLogInstance, make a super class?
    private void setInstance(InstanceId id) throws CothorityException {
        Proof p = bc.getProof(id);
        Instance inst = new Instance(p);
        if (!inst.getContractId().equals(ContractId)) {
            logger.error("wrong instance: {}", inst.getContractId());
            throw new CothorityNotFoundException("this is not an " + ContractId + " instance");
        }
        this.instance = inst;
        logger.info("new " + ContractId + " instance: " + inst.getId().toString());
    }
}
