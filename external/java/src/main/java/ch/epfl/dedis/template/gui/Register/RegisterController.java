package ch.epfl.dedis.template.gui.Register;

import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Rules;
import ch.epfl.dedis.lib.darc.Signer;
import ch.epfl.dedis.lib.darc.SignerEd25519;
import ch.epfl.dedis.template.Car;
import ch.epfl.dedis.template.CarInstance;
import ch.epfl.dedis.template.gui.errorScene.ErrorSceneController;
import ch.epfl.dedis.template.gui.index.Main;
import ch.epfl.dedis.template.gui.json.CarJson;
import ch.epfl.dedis.template.gui.json.Person;
import com.fasterxml.jackson.databind.ObjectMapper;
import javafx.beans.binding.Bindings;
import javafx.beans.property.BooleanProperty;
import javafx.beans.property.SimpleBooleanProperty;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.*;

import java.io.File;
import java.net.URL;
import java.util.*;

import static ch.epfl.dedis.byzcoin.contracts.DarcInstance.fromByzCoin;
//import static ch.epfl.dedis.template.gui.ByzSetup.getAdmin;
import static ch.epfl.dedis.template.gui.ByzSetup.*;
import static ch.epfl.dedis.template.gui.index.Main.calypsoRPC;
import static ch.epfl.dedis.template.gui.index.Main.homePath;

public class RegisterController implements Initializable {

    @FXML
    private TextField createReaderText;

    @FXML
    private TextField createCarText;

    @FXML
    private Button createGarageButton;

    @FXML
    private Button createReaderButton;

    @FXML
    private TextField createGarageText;

    @FXML
    private Button createCarButton;

    @FXML
    private Button createUserButton;

    @FXML
    private TextField createUserText;

    @FXML
    private MenuButton chooseOwnerButton;

    BooleanProperty disableUser = new SimpleBooleanProperty();
    BooleanProperty disableCar = new SimpleBooleanProperty();
    BooleanProperty disableReader = new SimpleBooleanProperty();
    BooleanProperty disableGarage = new SimpleBooleanProperty();

    ObjectMapper mapper = new ObjectMapper();

    @Override
    public void initialize(URL location, ResourceBundle resources){

        disableUser.bind(Bindings.createBooleanBinding(() ->
                createUserText.getText().trim().isEmpty(), createUserText.textProperty()));
        disableCar.bind(Bindings.createBooleanBinding(() ->
                createCarText.getText().trim().isEmpty(), createCarText.textProperty()));
        disableReader.bind(Bindings.createBooleanBinding(() ->
                createReaderText.getText().trim().isEmpty(), createReaderText.textProperty()));
        disableGarage.bind(Bindings.createBooleanBinding(() ->
                createGarageText.getText().trim().isEmpty(), createGarageText.textProperty()));

        createUserButton.disableProperty().bind(disableUser);
        createCarButton.disableProperty().bind(disableCar);
        createReaderButton.disableProperty().bind(disableReader);
        createGarageButton.disableProperty().bind(disableGarage);

        createUserButton.setOnAction(this::createUser);
        createReaderButton.setOnAction(this::createReader);
        createGarageButton.setOnAction(this::createGarage);
        createCarButton.setOnAction(this::createCar);

        createUserButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");
        createCarButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");
        createReaderButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");
        createGarageButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");

        try{
            //populate choose owner button
            File userFile = new File(homePath + "/json/user.json");
            if(userFile.exists())
            {
                HashMap<String, Person> userMap = mapper.readValue(userFile, HashMap.class);
                for (HashMap.Entry<String, Person> entry : userMap.entrySet()) {
                    MenuItem userItem = new MenuItem(entry.getKey());
                    userItem.setOnAction(this::onRoleChange);
                    chooseOwnerButton.getItems().add(userItem);
                }
            }
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
    }

    private void onRoleChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        chooseOwnerButton.setText(Identity);
    }


    private void createUser(ActionEvent event) {

        try{
            createPerson("user", createUserText.getText());
            createUserText.setText("");
            Main.window.setScene(Main.signUpResultScene);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
    }

    private void createReader(ActionEvent event) {
        try{
            createPerson("reader", createReaderText.getText());
            createReaderText.setText("");
            Main.window.setScene(Main.signUpResultScene);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
    }

    private void createGarage(ActionEvent event) {
        try{
            createPerson("garage", createGarageText.getText());
            createGarageText.setText("");
            Main.window.setScene(Main.signUpResultScene);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
    }

    private void createCar(ActionEvent event) {
        try{
            //todo: check the owner from the VIN and set the initial rules with that owner

            String path = homePath + "/json/car.json";

            //getting the admin details from the local file
            Darc adminDarc = getAdminDarc();
            DarcInstance adminDarcInstance = fromByzCoin(calypsoRPC, adminDarc.getId());

            //Spawning owner darc with invoke:evolve and _sign rules
            Darc ownerDarc = new Darc(Arrays.asList(getPersonDarc(chooseOwnerButton.getText(), "user").getIdentity()),
                    Arrays.asList(getPersonDarc(chooseOwnerButton.getText(), "user").getIdentity()), "Car Owner darc".getBytes());
            adminDarcInstance.spawnDarcAndWait(ownerDarc, getAdmin(), 5);

            //Spawning reader darc with invoke:evolve and _sign rules
            Darc readerDarc = new Darc(Arrays.asList(ownerDarc.getIdentity()),
                    Arrays.asList(ownerDarc.getIdentity()), "Car Reader darc".getBytes());
            adminDarcInstance.spawnDarcAndWait(readerDarc, getAdmin(), 5);

            //Spawning garage darc with invoke:evolve and _sign rules
            Darc garageDarc = new Darc(Arrays.asList(ownerDarc.getIdentity()),
                    Arrays.asList(ownerDarc.getIdentity()), "Car Garage darc".getBytes());
            adminDarcInstance.spawnDarcAndWait(garageDarc, getAdmin(), 5);

            //Spawning car darc with spawn:car, invoke:addReport, spawn:calypsoWrite and spawn:calypsoRead rules
            Rules rs = new Rules();
            rs.addRule("spawn:car", adminDarc.getIdentity().toString().getBytes());
            rs.addRule("invoke:addReport", garageDarc.getIdentity().toString().getBytes());
            rs.addRule("spawn:calypsoWrite", garageDarc.getIdentity().toString().getBytes());
            rs.addRule("spawn:calypsoRead", readerDarc.getIdentity().toString().getBytes());
            Darc carDarc = new Darc(rs, "Car darc".getBytes());
            DarcInstance carDarcInstance = adminDarcInstance.spawnDarcAndWait(carDarc, getAdmin(), 10);

            //Creating car instance from the given VIN
            Car c = new Car(createCarText.getText());
            CarInstance ci = new CarInstance(calypsoRPC, carDarcInstance, getAdmin(), c);

            CarJson carJson = new CarJson(createCarText.getText(), carDarc.toProto().toByteArray(),
                    ownerDarc.toProto().toByteArray(), readerDarc.toProto().toByteArray(),
                    garageDarc.toProto().toByteArray(), ci.getId().getId());

            File carFile = new File(path);
            if(!carFile.exists()) {
                HashMap<String, CarJson> carMap = new HashMap<>();
                carMap.put(carJson.VIN, carJson);
                mapper.writeValue(carFile, carMap);
            }
            else {
                HashMap<String, CarJson> carMap = mapper.readValue(carFile, HashMap.class);
                if(!carMap.containsKey(carJson.VIN)){
                    carMap.put(carJson.VIN, carJson);
                    mapper.writeValue(carFile, carMap);
                }
                else{
                    throw new Exception("A car with the provided VIN already exists");
                }
            }

            createGarageText.setText("");
            Main.window.setScene(Main.signUpResultScene);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
    }

    /**
     * @param role can be either user, reader or garage
     *
     */
    private void createPerson(String role, String name) throws Exception{
        String path = homePath + "/json/" + role + ".json";

        Darc adminDarc = getAdminDarc();
        DarcInstance adminDarcInstance = fromByzCoin(calypsoRPC, adminDarc.getId());

        SignerEd25519 signer = new SignerEd25519();
        Darc personDarc = new Darc(Arrays.asList(signer.getIdentity()), Arrays.asList(signer.getIdentity()), (role + " darc").getBytes());
        adminDarcInstance.spawnDarcAndWait(personDarc, getAdmin(), 10);

        Person person = new Person(name, personDarc.toProto().toByteArray(), signer.serialize());

        //if the file doesn't exist, we need to create a new list of users/readers/garages
        //otherwise take the list from the file and update it
        File personFile = new File(path);
        if(!personFile.exists()) {
            HashMap<String, Person> personMap = new HashMap<>();
            personMap.put(person.name, person);
            mapper.writeValue(personFile, personMap);
        }
        else {
            HashMap<String, Person> personMap = mapper.readValue(personFile, HashMap.class);
            if(!personMap.containsKey(person.name)){
                personMap.put(person.name, person);
                mapper.writeValue(personFile, personMap);
            }
            else{
                throw new Exception("The name is already taken");
            }
        }
    }



}
