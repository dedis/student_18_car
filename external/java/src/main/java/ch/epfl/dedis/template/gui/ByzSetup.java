package ch.epfl.dedis.template.gui;

import ch.epfl.dedis.byzcoin.ByzCoinRPC;
import ch.epfl.dedis.byzcoin.InstanceId;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.calypso.CalypsoRPC;
import ch.epfl.dedis.calypso.LTSId;
import ch.epfl.dedis.lib.Roster;
import ch.epfl.dedis.lib.SkipblockId;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Signer;
import ch.epfl.dedis.lib.darc.SignerEd25519;
import ch.epfl.dedis.lib.darc.SignerFactory;
import ch.epfl.dedis.lib.proto.DarcProto;
import ch.epfl.dedis.lib.proto.OnetProto;
import ch.epfl.dedis.lib.proto.SkipchainProto;
import ch.epfl.dedis.template.gui.json.ByzC;
import ch.epfl.dedis.template.gui.json.CarJson;
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
    public static ByzC setup(){

        try{
            ByzC byzC;
            ObjectMapper mapper = new ObjectMapper();
            File byzFile = new File("/Users/Iva/json/byzcoin.json");
            if(!byzFile.exists()) {

                //creating genesis admin, darc and the byzcoin blockchain itself
                Signer genAdmin = new SignerEd25519();
                Darc genesisDarc = ByzCoinRPC.makeGenesisDarc(genAdmin, Roster.FromToml(tomlStr));
                CalypsoRPC calypso = new CalypsoRPC(Roster.FromToml(tomlStr), genesisDarc, Duration.of(500, MILLIS));
                DarcInstance genesisDarcInstance = calypso.getGenesisDarcInstance();


                //storing configuration details on disk, as json object
                byzC = new ByzC(Roster.FromToml(tomlStr).toProto().toByteArray(),
                        calypso.getGenesisBlock().getId().toProto().toByteArray(),
                        calypso.getLTSId().toProto().toByteArray());
                mapper.writeValue(new File("/Users/Iva/json/byzcoin.json"), byzC);


                //creating admin and admin darc
                SignerEd25519 admin = new SignerEd25519();
                Darc adminDarc = new Darc(Arrays.asList(admin.getIdentity()), Arrays.asList(admin.getIdentity()), "Admin darc".getBytes());
                adminDarc.setRule("spawn:darc", admin.getIdentity().toString().getBytes());
                DarcInstance adminDarcInstance = genesisDarcInstance.spawnDarcAndWait(adminDarc, genAdmin, 10);

                //storing admin details on disk
                Person admPerson = new Person("admin", adminDarc.toProto().toByteArray(), admin.serialize());
                File adminFile = new File("/Users/Iva/json/admin.json");
                mapper.writeValue(adminFile, admPerson);

                return byzC;
            }
            else
            {
                byzC = mapper.readValue(new File("/Users/Iva/json/byzcoin.json"), ByzC.class);
                return byzC;
            }
        }
        catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    /**
     *
     * @param byzC that contains byte[] version of the Roster
     * @return Roster object
     * @throws Exception
     */
    public static Roster getRoster(ByzC byzC){
        try{
            OnetProto.Roster rosterProto = OnetProto.Roster.parseFrom(byzC.roster);
            Roster roster = new Roster(rosterProto);
            return roster;
        }
        catch (Exception e){
            e.printStackTrace();
            return null;
        }
    }

    /**
     *
     * @param byzC that contains byte[] version of the SkipblockId
     * @return SkipblockId object
     * @throws Exception
     */
    public static SkipblockId getByzId(ByzC byzC){
        SkipblockId ByzId  = new SkipblockId(byzC.skipblockId);
        return ByzId;
    }

    /**
     *
     * @param byzC that contains byte[] version of the LTSId
     * @return LTSId object
     * @throws Exception
     */
    public static LTSId getLTSId(ByzC byzC){
        try{
            LTSId ltsId  = new LTSId(byzC.ltsId);
            return ltsId;
        }
        catch (Exception e){
            e.printStackTrace();
            return null;
        }
    }

    public static Darc getAdminDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File adminFile = new File("/Users/Iva/json/admin.json");
        if(adminFile.exists()){
            Person admPerson = mapper.readValue(adminFile, Person.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(admPerson.darc);
            Darc darcAdmin = new Darc(darcProto);
            return  darcAdmin;
        }
        else
            return null;
    }

    public static String getAdminName() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File adminFile = new File("/Users/Iva/json/admin.json");
        if(adminFile.exists()){
            Person admPerson = mapper.readValue(adminFile, Person.class);
            return admPerson.name;
        }
        else
            return null;
    }

    public static SignerEd25519 getAdmin() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File adminFile = new File("/Users/Iva/json/admin.json");
        if(adminFile.exists()){
            Person admPerson = mapper.readValue(adminFile, Person.class);
            return (SignerEd25519)SignerFactory.New(admPerson.signer);
        }
        else
            return null;
    }

    public static Darc getGarageDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File garageFile = new File("/Users/Iva/json/garage.json");
        if(garageFile.exists()){
            Person garagePerson = mapper.readValue(garageFile, Person.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(garagePerson.darc);
            Darc darcGarage = new Darc(darcProto);
            return  darcGarage;
        }
        else
            return null;
    }

    public static SignerEd25519 getGarage() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File garageFile = new File("/Users/Iva/json/garage.json");
        if(garageFile.exists()){
            Person garagePerson = mapper.readValue(garageFile, Person.class);
            return (SignerEd25519)SignerFactory.New(garagePerson.signer);
        }
        else
            return null;
    }

    public static String getGarageName() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File garageFile = new File("/Users/Iva/json/garage.json");
        if(garageFile.exists()){
            Person garagePerson = mapper.readValue(garageFile, Person.class);
            return garagePerson.name;
        }
        else
            return null;
    }

    public static Darc getUserDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File userFile = new File("/Users/Iva/json/user.json");
        if(userFile.exists()){
            Person userPerson = mapper.readValue(userFile, Person.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(userPerson.darc);
            Darc darcUser = new Darc(darcProto);
            return  darcUser;
        }
        else
            return null;

    }

    public static SignerEd25519 getUser() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File userFile = new File("/Users/Iva/json/user.json");
        if(userFile.exists()){
            Person userPerson = mapper.readValue(userFile, Person.class);
            return (SignerEd25519)SignerFactory.New(userPerson.signer);
        }
        else
            return null;
    }

    public static String getUserName() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File userFile = new File("/Users/Iva/json/user.json");
        if(userFile.exists()){
            Person userPerson = mapper.readValue(userFile, Person.class);
            return userPerson.name;
        }
        else
            return null;
    }

    public static Darc getReaderDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File readerFile = new File("/Users/Iva/json/reader.json");
        if(readerFile.exists()){
            Person readerPerson = mapper.readValue(readerFile, Person.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(readerPerson.darc);
            Darc darcReader = new Darc(darcProto);
            return  darcReader;
        }
        else
            return null;
    }


    public static Darc getCarReaderDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File("/Users/Iva/json/car.json");
        if(carFile.exists()){
            CarJson carJson = mapper.readValue(carFile, CarJson.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(carJson.readerDarc);
            Darc darcReader = new Darc(darcProto);
            return  darcReader;
        }
        else
            return null;
    }





    public static SignerEd25519 getReader() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File readerFile = new File("/Users/Iva/json/reader.json");
        if(readerFile.exists()){
            Person readerPerson = mapper.readValue(readerFile, Person.class);
            return (SignerEd25519)SignerFactory.New(readerPerson.signer);
        }
        else
            return null;
    }

    public static String getReaderName() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File readerFile = new File("/Users/Iva/json/reader.json");
        if(readerFile.exists()){
            Person userPerson = mapper.readValue(readerFile, Person.class);
            return userPerson.name;
        }
        else
            return null;
    }

    public static String getVIN() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File("/Users/Iva/json/car.json");
        if(carFile.exists()){
            CarJson carJson = mapper.readValue(carFile, CarJson.class);
            return carJson.VIN;
        }
        else
            return null;
    }


    public static Darc getCarDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File("/Users/Iva/json/car.json");
        if(carFile.exists()){
            CarJson carJson = mapper.readValue(carFile, CarJson.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(carJson.darc);
            Darc darcCar = new Darc(darcProto);
            return  darcCar;
        }
        else
            return null;
    }


    public static Darc getCarGarageDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File("/Users/Iva/json/car.json");
        if(carFile.exists()){
            CarJson carJson = mapper.readValue(carFile, CarJson.class);
            DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(carJson.garageDarc);
            Darc darcGarage = new Darc(darcProto);
            return  darcGarage;
        }
        else
            return null;
    }

    public static InstanceId getCarInstanceId() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File("/Users/Iva/json/car.json");
        if(carFile.exists()){
            CarJson carJson = mapper.readValue(carFile, CarJson.class);
            InstanceId id = new InstanceId(carJson.instanceId);
            return id;
        }
        else
            return null;
    }



}
