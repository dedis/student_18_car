package ch.epfl.dedis.template.gui.Register;

import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Rules;
import ch.epfl.dedis.lib.darc.Signer;
import ch.epfl.dedis.lib.darc.SignerEd25519;
import ch.epfl.dedis.template.Car;
import ch.epfl.dedis.template.CarInstance;
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
import java.util.Arrays;
import java.util.ResourceBundle;

import static ch.epfl.dedis.byzcoin.contracts.DarcInstance.fromByzCoin;
//import static ch.epfl.dedis.template.gui.ByzSetup.getAdmin;
import static ch.epfl.dedis.template.gui.ByzSetup.*;
import static ch.epfl.dedis.template.gui.index.Main.calypsoRPC;

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

        try{
            if(getUserName()!=null)
            {
                MenuItem userItem = new MenuItem(getUserName());
                userItem.setOnAction(this::onRoleChange);
                chooseOwnerButton.getItems().add(userItem);
            }
        }
        catch (Exception e){
            e.printStackTrace();
        }
    }

    private void onRoleChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        chooseOwnerButton.setText(Identity);
    }


    private void createUser(ActionEvent event) {

        try{
            Darc adminDarc = getAdminDarc();
            DarcInstance adminDarcInstance = fromByzCoin(calypsoRPC, adminDarc.getId());

            SignerEd25519 user = new SignerEd25519();
            Darc userDarc = new Darc(Arrays.asList(user.getIdentity()), Arrays.asList(user.getIdentity()), "User darc".getBytes());
            DarcInstance userDarcInstance = adminDarcInstance.spawnDarcAndWait(userDarc, getAdmin(), 10);

            Person userPerson = new Person(createUserText.getText(), userDarc.toProto().toByteArray(), user.serialize());

            ObjectMapper mapper = new ObjectMapper();
            mapper.writeValue(new File("/Users/Iva/json/user.json"), userPerson);

            createUserText.setText("");
            Main.window.setScene(Main.signUpResultScene);

        }
        catch (Exception e){
            e.printStackTrace();
        }
    }

    private void createReader(ActionEvent event) {
        try{
            Darc adminDarc = getAdminDarc();
            DarcInstance adminDarcInstance = fromByzCoin(calypsoRPC, adminDarc.getId());

            SignerEd25519 reader = new SignerEd25519();
            Darc userDarc = new Darc(Arrays.asList(reader.getIdentity()), Arrays.asList(reader.getIdentity()), "User darc".getBytes());
            DarcInstance userDarcInstance = adminDarcInstance.spawnDarcAndWait(userDarc, getAdmin(), 10);

            Person readerPerson = new Person(createReaderText.getText(), userDarc.toProto().toByteArray(), reader.serialize());

            ObjectMapper mapper = new ObjectMapper();
            mapper.writeValue(new File("/Users/Iva/json/reader.json"), readerPerson);

            createReaderText.setText("");
            Main.window.setScene(Main.signUpResultScene);
        }
        catch (Exception e){
            e.printStackTrace();
        }
    }

    private void createGarage(ActionEvent event) {
        try{
            //getting the admin details from the local file
            Darc adminDarc = getAdminDarc();
            DarcInstance adminDarcInstance = fromByzCoin(calypsoRPC, adminDarc.getId());

            SignerEd25519 garage = new SignerEd25519();
            Darc userDarc = new Darc(Arrays.asList(garage.getIdentity()),
                    Arrays.asList(garage.getIdentity()), "User darc".getBytes());
            DarcInstance userDarcInstance = adminDarcInstance.spawnDarcAndWait(userDarc, getAdmin(), 10);

            Person garagePerson = new Person(createGarageText.getText(),
                    userDarc.toProto().toByteArray(), garage.serialize());

            ObjectMapper mapper = new ObjectMapper();
            mapper.writeValue(new File("/Users/Iva/json/garage.json"), garagePerson);

            createGarageText.setText("");
            Main.window.setScene(Main.signUpResultScene);
        }
        catch (Exception e){
            e.printStackTrace();
        }
    }

    private void createCar(ActionEvent event) {
        try{
            //todo: check the owner from the VIN and set the initial rules with that owner

            //getting the admin details from the local file
            Darc adminDarc = getAdminDarc();
            DarcInstance adminDarcInstance = fromByzCoin(calypsoRPC, adminDarc.getId());

            //Spawning owner darc with invoke:evolve and _sign rules
            Darc ownerDarc = new Darc(Arrays.asList(getUserDarc().getIdentity()),
                    Arrays.asList(getUserDarc().getIdentity()), "User darc".getBytes());
            DarcInstance ownerDarcInstance = adminDarcInstance.spawnDarcAndWait(ownerDarc, getAdmin(), 10);

            //Spawning reader darc with invoke:evolve and _sign rules
            Darc readerDarc = new Darc(Arrays.asList(getUserDarc().getIdentity()),
                    Arrays.asList(getUserDarc().getIdentity()), "Reader darc".getBytes());
            DarcInstance readerDarcInstance = adminDarcInstance.spawnDarcAndWait(readerDarc, getAdmin(), 10);

            //Spawning garage darc with invoke:evolve and _sign rules
            Darc garageDarc = new Darc(Arrays.asList(getUserDarc().getIdentity()),
                    Arrays.asList(getUserDarc().getIdentity()), "Garage darc".getBytes());
            DarcInstance garageDarcInstance = adminDarcInstance.spawnDarcAndWait(garageDarc, getAdmin(), 10);

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
            System.out.println(ci.getId());

            CarJson carJson = new CarJson(createCarText.getText(), carDarc.toProto().toByteArray(),
                    ownerDarc.toProto().toByteArray(), readerDarc.toProto().toByteArray(),
                    garageDarc.toProto().toByteArray(), ci.getId().getId());

            ObjectMapper mapper = new ObjectMapper();
            mapper.writeValue(new File("/Users/Iva/json/car.json"), carJson);

            System.out.println(getCarInstanceId());

            createGarageText.setText("");
            Main.window.setScene(Main.signUpResultScene);
        }
        catch (Exception e){
            e.printStackTrace();
        }
    }


    /*@FXML
    void onSelection(ActionEvent event){
        RadioButton r = (RadioButton)event.getTarget();

        if (r.isSelected())
        {
            System.out.println(r.getText());
        }
    }


    @FXML
    void signUp(ActionEvent event) {
        RadioButton selectedRole = (RadioButton) roleGroup.getSelectedToggle();

        if (selectedRole.getText() == "Garage")
        {

        }

    }*/

}
