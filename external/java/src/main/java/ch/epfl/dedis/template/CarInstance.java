package ch.epfl.dedis.template;

import ch.epfl.dedis.byzcoin.transaction.Argument;
import ch.epfl.dedis.byzcoin.transaction.ClientTransaction;
import ch.epfl.dedis.byzcoin.transaction.Instruction;
import ch.epfl.dedis.byzcoin.transaction.Invoke;
import ch.epfl.dedis.lib.Hex;
import ch.epfl.dedis.calypso.*;
import ch.epfl.dedis.lib.darc.*;
import ch.epfl.dedis.lib.exception.CothorityCryptoException;
import ch.epfl.dedis.lib.exception.CothorityException;
import ch.epfl.dedis.lib.exception.CothorityNotFoundException;
import ch.epfl.dedis.byzcoin.*;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.template.proto.CarProto;
import com.google.protobuf.InvalidProtocolBufferException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;


/**
 * CarInstance represents a car on ByzCoin. It can be initialised either by the
 * instanceID, in which case it will fetch itself the needed data from ByzCoin. Or it is initialized
 * with a proof, then it will simply copy the values stored in the proof to create a new CarInstance.
 */
public class  CarInstance {
    private Instance instance;
    private CalypsoRPC bc;
    private Car car;

    private final static Logger logger = LoggerFactory.getLogger(CarInstance.class);

    /**
     * Instantiates a new KeyValueInstance given a working ByzCoin instance and
     * an instanceId. This instantiator will contact ByzCoin and try to get
     * the current valueInstance. If the instance is not found, or is not of
     * contractId "Value", an exception will be thrown.
     *
     * @param ol is a link to a ByzCoin instance that is running
     * @param id of the value-instance to connect to
     * @throws CothorityException
     */
    public CarInstance(CalypsoRPC ol, InstanceId id) throws CothorityException {
        this.bc = ol;
        update(id);
    }

    /**
     * Instantiates a KeyValueInstance given a proof.
     *
     * @param bc is a link to a ByzCoin instance that is running
     * @param p is a proof for a valid KeyValue instance
     * @throws CothorityException
     */
    public CarInstance(CalypsoRPC bc, Proof p) throws CothorityException {
        this.bc = bc;
        update(p);
    }

    /**
     * Spawns a new Car on ByzCoin.
     *
     * @param bc a working ByzCoin instance
     * @param darcInstance a darcInstance with a rule "spawn:car"
     * @param signer with the right to execute the "spawn:car" rule from the darcInstance
     * @param c a Car to include in the data of the instance
     * @throws CothorityException
     */
    public CarInstance(CalypsoRPC bc, DarcInstance darcInstance,
                       Signer signer, Car c) throws CothorityException{
        this.bc = bc;
        List<Argument> args = new ArrayList<>();
        args.add(c.toArgument());
        Proof p = darcInstance.spawnInstanceAndWait("car", signer, args, 10);
        update(p);
    }

    /**
     * Updates an existing KeyValueInstance in case it has been updated in ByzCoin.
     * @throws CothorityException
     */
    public void update() throws CothorityException {
        if (instance == null || bc == null || car == null){
            throw new CothorityException("instance not initialized yet");
        }
        update(instance.getId());
    }

    /**
     * updates the keyvalue instance from a live ByzCoin.
     *
     * @throws CothorityException
     */
    public void update(InstanceId id) throws CothorityException {
        update(bc.getProof(id));
    }

    /**
     * Updates the keyvalue instance from a given proof.
     *
     * @param pr the proof to the KeyValue instance
     * @throws CothorityException
     */
    public void update(Proof pr) throws CothorityException{
        if (!pr.matches()){
            throw new CothorityException("cannot use non-matching proof for update");
        }
        instance = Instance.fromProof(pr);
        if (!instance.getContractId().equals("car")) {
            logger.error("wrong instance: {}", instance.getContractId());
            throw new CothorityNotFoundException("this is not a car instance");
        }
        try {
            CarProto.Car cP = CarProto.Car.parseFrom(instance.getData());
            car = new Car(cP);

        } catch (InvalidProtocolBufferException e) {
            throw new CothorityException(e);
        }
    }

    /**
     * Creates an instruction to evolve the car in ByzCoin. The signer must have its identity in the current
     * darc as "invoke:addReport" rule.
     * <p>
     * TODO: allow for evolution if the expression has more than one identity.
     *
     * @param reports the reports to be added to the list of reports in the car instance.
     * @param signer     must have its identity in the "invoke:addReport" rule
     * @param pos       position of the instruction in the ClientTransaction
     * @param len       total number of instructions in the ClientTransaction
     * @return Instruction to be sent to ByzCoin
     * @throws CothorityCryptoException
     */
    public Instruction addReportInstruction(List<Report> reports, Signer signer, int pos, int len) throws CothorityCryptoException {
        List<Argument> args = new ArrayList<>();
        for (Report rep : reports) {
            args.add(rep.toArgument());
        }
        Invoke inv = new Invoke("addReport", args);
        Instruction inst = new Instruction(instance.getId(), Instruction.genNonce(), pos, len, inv);
        try {
            Request r = new Request(instance.getDarcId(), "invoke:addReport", inst.hash(),
                    Arrays.asList(signer.getIdentity()), null);
            logger.info("Signing: {}", Hex.printHexBinary(r.hash()));
            Signature sign = new Signature(signer.sign(r.hash()), signer.getIdentity());
            inst.setSignatures(Arrays.asList(sign));
        } catch (Signer.SignRequestRejectedException e) {
            throw new CothorityCryptoException(e.getMessage());
        }
        return inst;
    }

    /**
     * Sends a request to update the keyvalue instance but doesn't wait for the request to be delivered.
     *
     * @param reports the keyValues to replace/delete/add to the list.
     * @param owner must have its identity in the "invoke:update" rule
     * @return a TransactionId pointing to the transaction that should be included
     * @throws CothorityException
     */
    public void addReport(List<Report> reports, Signer owner) throws CothorityException {
        Instruction inst = addReportInstruction(reports, owner, 0, 1);
        ClientTransaction ct = new ClientTransaction(Arrays.asList(inst));
        bc.sendTransaction(ct);
    }

    /**
     * Asks ByzCoin to update the value and waits until the new value has
     * been stored in the global state.
     *
     * @param reports the value to replace the old value.
     * @param owner     is the owner that can sign to evolve the darc
     * @param wait      is the number of blocks to wait for an inclusion
     * @throws CothorityException
     */
    public void addReportAndWait(List<Report> reports, Signer owner, int wait) throws CothorityException {
        Instruction inst = addReportInstruction(reports, owner, 0, 1);
        ClientTransaction ct = new ClientTransaction(Arrays.asList(inst));
        bc.sendTransactionAndWait(ct, wait);
        update();
    }

    /**
     * @return the id of the instance
     */
    public InstanceId getId() {
        return instance.getId();
    }

    /**
     * @return a copy of the reports stored in this instance.
     */
    public List<Report> getReports() throws CothorityCryptoException {
        List<Report> ret = new ArrayList<>();
        for (Report rep : car.getReportsList()) {
            ret.add(rep);
        }
        return ret;
    }

    public String getVin() throws CothorityCryptoException {
        return car.getVin();
    }

    /**
     * @return the instance used.
     */
    public Instance getInstance() {
        return instance;
    }

    /**
     * Fetches an already existing writeInstance from Calypso and returns it.
     *
     * @param calypso the Calypso instance
     * @param carInstId the car instance to load
     * @return the new carInstance
     * @throws CothorityException if something goes wrong
     */
    public static CarInstance fromCalypso(CalypsoRPC calypso, InstanceId carInstId) throws CothorityException {
        return new CarInstance(calypso, carInstId);
    }
}
