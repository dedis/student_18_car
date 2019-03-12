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
import ch.epfl.dedis.template.gui.index.Main;
import ch.epfl.dedis.template.gui.json.ByzC;
import ch.epfl.dedis.template.gui.json.CarJson;
import ch.epfl.dedis.template.gui.json.Person;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.File;
import java.time.Duration;
import java.util.Arrays;
import java.util.HashMap;

import static ch.epfl.dedis.template.gui.index.Main.homePath;
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

    public static String tomlStrEPFL = "[[servers]]\n" +
            "  Address = \"tcp://dedis-ns2.epfl.ch:7002\"\n" +
            "  Public = \"d829a0790ffa8799e4bbd1bee8da0507c9166b665660baba72dd8610fca27cc1\"\n" +
            "  Description = \"Conode_1\"\n" +
            "[[servers]]\n" +
            "  Address = \"tcp://dedis-ns2.epfl.ch:7004\"\n" +
            "  Public = \"d750a30daa44713d1a4b44ca4ef31142b3b53c0c36a558c0d610cc4108bb4ecb\"\n" +
            "  Description = \"Conode_2\"\n" +
            "[[servers]]\n" +
            "  Address = \"tcp://dedis-ns2.epfl.ch:7006\"\n" +
            "  Public = \"7f47f33084c3ecc233f8b05b8f408bbd1c2e4a129aae126f92becacc73576bc7\"\n" +
            "  Description = \"Conode_3\"\n" +
            "[[servers]]\n" +
            "  Address = \"tcp://dedis-ns2.epfl.ch:7008\"\n" +
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
            File byzFile = new File(homePath + "/json/byzcoin.json");
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
                mapper.writeValue(new File(homePath + "/json/byzcoin.json"), byzC);


                //creating admin and admin darc
                SignerEd25519 admin = new SignerEd25519();
                Darc adminDarc = new Darc(Arrays.asList(admin.getIdentity()), Arrays.asList(admin.getIdentity()), "Admin darc".getBytes());
                adminDarc.setRule("spawn:darc", admin.getIdentity().toString().getBytes());
                DarcInstance adminDarcInstance = genesisDarcInstance.spawnDarcAndWait(adminDarc, genAdmin, 10);

                //storing admin details on disk
                Person admPerson = new Person("admin", adminDarc.toProto().toByteArray(), admin.serialize());
                File adminFile = new File(homePath + "/json/admin.json");
                mapper.writeValue(adminFile, admPerson);

                return byzC;
            }
            else
            {
                byzC = mapper.readValue(new File(homePath + "/json/byzcoin.json"), ByzC.class);
                return byzC;
            }
        }
        catch (Exception e) {
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
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
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
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
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
            return null;
        }
    }

    public static Darc getAdminDarc() throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File adminFile = new File(homePath + "/json/admin.json");
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
        File adminFile = new File(homePath + "/json/admin.json");
        if(adminFile.exists()){
            Person admPerson = mapper.readValue(adminFile, Person.class);
            return admPerson.name;
        }
        else
            return null;
    }

    public static SignerEd25519 getAdmin() throws Exception{
        ObjectMapper mapper = new ObjectMapper();
        File adminFile = new File(homePath + "/json/admin.json");
        if(adminFile.exists()){
            Person admPerson = mapper.readValue(adminFile, Person.class);
            return (SignerEd25519)SignerFactory.New(admPerson.signer);
        }
        else
            return null;
    }

    public static Darc getPersonDarc(String name, String role) throws  Exception{

        String path = homePath + "/json/" + role + ".json";
        ObjectMapper mapper = new ObjectMapper();
        File personFile = new File(path);


        if(personFile.exists()){
            TypeReference<HashMap<String, Person>> typeRef
                    = new TypeReference<HashMap<String, Person>>() {};
            HashMap<String, Person> personMap = mapper.readValue(personFile, typeRef);

            if (personMap.containsKey(name)){
                Person person = new Person(personMap.get(name));
                DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(person.darc);
                Darc darc = new Darc(darcProto);
                return  darc;
            }
            else
                throw new Exception("No person with that name exists");
        }
        else
            throw new Exception("No person with that name exists");
    }

    public static SignerEd25519 getPersonSigner(String name, String role) throws Exception{

        String path = homePath + "/json/" + role + ".json";

        ObjectMapper mapper = new ObjectMapper();
        File personFile = new File(path);
        if(personFile.exists()){
            TypeReference<HashMap<String, Person>> typeRef
                    = new TypeReference<HashMap<String, Person>>() {};
            HashMap<String, Person> personMap = mapper.readValue(personFile, typeRef);
            if (personMap.containsKey(name))
                return (SignerEd25519)SignerFactory.New(personMap.get(name).signer);
            else
                throw new Exception("No person with that name exists");
        }
        else
            return null;
    }

    public static Darc getCarReaderDarc(String VIN) throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File(homePath + "/json/car.json");
        if(carFile.exists()){

            TypeReference<HashMap<String, CarJson>> typeRef
                    = new TypeReference<HashMap<String, CarJson>>() {};
            HashMap<String, CarJson> carMap = mapper.readValue(carFile, typeRef);

            if (carMap.containsKey(VIN)){
                DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(carMap.get(VIN).readerDarc);
                Darc darc = new Darc(darcProto);
                return  darc;
            }
            else
                throw new Exception("No car with that VIN exists");
        }
        else
            throw new Exception("No car with that VIN exists");
    }


    public static Darc getCarOwnerDarc(String VIN) throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File(homePath + "/json/car.json");
        if(carFile.exists()){

            TypeReference<HashMap<String, CarJson>> typeRef
                    = new TypeReference<HashMap<String, CarJson>>() {};
            HashMap<String, CarJson> carMap = mapper.readValue(carFile, typeRef);

            if (carMap.containsKey(VIN)){
                DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(carMap.get(VIN).ownerDarc);
                Darc darc = new Darc(darcProto);
                return  darc;
            }
            else
                throw new Exception("No car with that VIN exists");
        }
        else
            throw new Exception("No car with that VIN exists");
    }

//    public static String getVIN() throws Exception{
//        ObjectMapper mapper = new ObjectMapper();
//        File carFile = new File(homePath + "/json/car.json");
//        if(carFile.exists()){
//            CarJson carJson = mapper.readValue(carFile, CarJson.class);
//            return carJson.VIN;
//        }
//        else
//            return null;
//    }


    public static Darc getCarDarc(String VIN) throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File(homePath + "/json/car.json");
        if(carFile.exists()){
            TypeReference<HashMap<String, CarJson>> typeRef
                    = new TypeReference<HashMap<String, CarJson>>() {};
            HashMap<String, CarJson> carMap = mapper.readValue(carFile, typeRef);
            if (carMap.containsKey(VIN)){
                DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(carMap.get(VIN).darc);
                Darc darc = new Darc(darcProto);
                return  darc;
            }
            else
                throw new Exception("No car with that VIN exists");
        }
        else
            throw new Exception("No car with that VIN exists");
    }


    public static Darc getCarGarageDarc(String VIN) throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File(homePath + "/json/car.json");
        if(carFile.exists()){
            TypeReference<HashMap<String, CarJson>> typeRef
                    = new TypeReference<HashMap<String, CarJson>>() {};
            HashMap<String, CarJson> carMap = mapper.readValue(carFile, typeRef);
            if (carMap.containsKey(VIN)){
                DarcProto.Darc darcProto = DarcProto.Darc.parseFrom(carMap.get(VIN).garageDarc);
                Darc darc = new Darc(darcProto);
                return  darc;
            }
            else
                throw new Exception("No car with that VIN exists");
        }
        else
            throw new Exception("No car with that VIN exists");
    }

    public static InstanceId getCarInstanceId(String VIN) throws  Exception{

        ObjectMapper mapper = new ObjectMapper();
        File carFile = new File(homePath + "/json/car.json");
        if(carFile.exists()){
            TypeReference<HashMap<String, CarJson>> typeRef
                    = new TypeReference<HashMap<String, CarJson>>() {};
            HashMap<String, CarJson> carMap = mapper.readValue(carFile, typeRef);
            if (carMap.containsKey(VIN)){
                CarJson carJson = carMap.get(VIN);
                InstanceId id = new InstanceId(carJson.instanceId);
                return id;
            }
            else
                throw new Exception("No car with that VIN exists");
        }
        else
            throw new Exception("No car with that VIN exists");
    }



}
