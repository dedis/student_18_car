package ch.epfl.dedis.template.gui.userScreen;

import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Identity;
import ch.epfl.dedis.lib.darc.Rule;
import ch.epfl.dedis.lib.darc.Rules;
import ch.epfl.dedis.template.gui.errorScene.ErrorSceneController;
import ch.epfl.dedis.template.gui.index.IndexController;
import ch.epfl.dedis.template.gui.index.Main;
import ch.epfl.dedis.template.gui.json.CarJson;
import ch.epfl.dedis.template.gui.json.Person;
import com.fasterxml.jackson.databind.ObjectMapper;
import javafx.collections.ObservableList;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.*;

import java.io.File;
import java.net.URL;
import java.util.*;
import java.util.stream.Collectors;

import static ch.epfl.dedis.byzcoin.contracts.DarcInstance.fromByzCoin;
import static ch.epfl.dedis.template.gui.ByzSetup.*;
import static ch.epfl.dedis.template.gui.ByzSetup.getPersonDarc;
import static ch.epfl.dedis.template.gui.index.Main.calypsoRPC;
import static ch.epfl.dedis.template.gui.index.Main.homePath;
import static ch.epfl.dedis.template.gui.index.Main.signUpResultScene;

public class userScreenController implements Initializable {

    @FXML
    private ListView<CheckBox> readersList;

    @FXML
    private MenuButton changeOwnerButton;

    @FXML
    private Button updateButton;

    @FXML
    private MenuButton chooseVinButton;

    @FXML
    private TextField roleID;

    @FXML
    private ListView<CheckBox> garagesList;

    @Override
    public void initialize(URL location, ResourceBundle resources){


        readersList.getSelectionModel().setSelectionMode(SelectionMode.MULTIPLE);
        garagesList.getSelectionModel().setSelectionMode(SelectionMode.MULTIPLE);

        try{

            roleID.setText("ID: " + IndexController.role);
            ObjectMapper mapper = new ObjectMapper();

            File carFile = new File(homePath + "/json/car.json");
            if(carFile.exists())
            {
                HashMap<String, CarJson> carMap = mapper.readValue(carFile, HashMap.class);

                for (HashMap.Entry<String, CarJson> entry : carMap.entrySet()) {
//                    if (getCarOwnerDarc(entry.getKey()).equals(getPersonDarc(IndexController.role,"user")))
//                    {
                        MenuItem VINItem = new MenuItem(entry.getKey());
                        VINItem.setOnAction(this::onVINChange);
                        chooseVinButton.getItems().add(VINItem);
//                    }

                }
            }

            File userFile = new File(homePath + "/json/user.json");
            if(userFile.exists())
            {
                HashMap<String, Person> userMap = mapper.readValue(userFile, HashMap.class);

                for (HashMap.Entry<String, Person> entry : userMap.entrySet()) {
                    if (!entry.getKey().equals(IndexController.role)){
                        MenuItem personItem = new MenuItem(entry.getKey());
                        personItem.setOnAction(this::onOwnerChange);
                        changeOwnerButton.getItems().add(personItem);
                    }
                }
            }

            File readerFile = new File(homePath + "/json/reader.json");
            if(readerFile.exists())
            {
                HashMap<String, Person> personMap = mapper.readValue(readerFile, HashMap.class);

                for (HashMap.Entry<String, Person> entry : personMap.entrySet()) {
                    CheckBox readerBox = new CheckBox(entry.getKey());
                    readersList.getItems().add(readerBox);
                }
            }

            File garageFile = new File(homePath + "/json/garage.json");
            if(garageFile.exists())
            {
                HashMap<String, Person> personMap = mapper.readValue(garageFile, HashMap.class);

                for (HashMap.Entry<String, Person> entry : personMap.entrySet()) {
                    CheckBox garageBox = new CheckBox(entry.getKey());
                    garagesList.getItems().add(garageBox);
                }
            }
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
        
        updateButton.setOnAction(this::update);
        updateButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");

    }

    private void onOwnerChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        changeOwnerButton.setText(Identity);
    }


    //todo make it work with removal of access
    //todo if the same are selected it throws an exception
    private void update(ActionEvent event) {
        ObservableList<CheckBox> readers = readersList.getItems();
        ObservableList<CheckBox> garages = garagesList.getItems();
        try {

            ObjectMapper mapper = new ObjectMapper();
            File userFile = new File(homePath + "/json/user.json");
            if(userFile.exists()) {
                HashMap<String, Person> userMap = mapper.readValue(userFile, HashMap.class);
                Darc carOwnerDarc = getCarOwnerDarc(chooseVinButton.getText());

                List<Identity> readerList = new ArrayList<>();
                readerList.add(carOwnerDarc.getIdentity());
                List<Identity> garageList =  new ArrayList<>();
                garageList.add(carOwnerDarc.getIdentity());

                for (CheckBox reader : readers) {
                    if (reader.isSelected()) {
                        readerList.add(getPersonDarc(reader.getText(), "reader").getIdentity());

                    }
                }

                Darc carReaderDarc = getCarReaderDarc(chooseVinButton.getText());
                DarcInstance carReaderDarcInstance = fromByzCoin(calypsoRPC, carReaderDarc.getId());
                List<String> readerSignerIDs = readerList.stream().map(Identity::toString).collect(Collectors.toList());
                Darc newDarcR = carReaderDarcInstance.getDarc();
                newDarcR.setRule(Darc.RuleSignature, String.join(" | ", readerSignerIDs).getBytes());
                carReaderDarcInstance.evolveDarcAndWait(newDarcR,
                        getPersonSigner(IndexController.role, "user"), 5);

                for (CheckBox garage : garages) {
                    if (garage.isSelected()) {

                        garageList.add(getPersonDarc(garage.getText(), "garage").getIdentity());
                    }
                }

                Darc carGarageDarc = getCarGarageDarc(chooseVinButton.getText());
                DarcInstance carGarageDarcInstance = fromByzCoin(calypsoRPC, carGarageDarc.getId());
                List<String> garageSignerIDs = garageList.stream().map(Identity::toString).collect(Collectors.toList());
                Darc newDarcG = carGarageDarcInstance.getDarc();
                newDarcG.setRule(Darc.RuleSignature, String.join(" | ", garageSignerIDs).getBytes());
                carGarageDarcInstance.evolveDarcAndWait(newDarcG,
                        getPersonSigner(IndexController.role, "user"), 5);


                if(userMap.containsKey(changeOwnerButton.getText())){
                    DarcInstance carOwnerDarcInstance = fromByzCoin(calypsoRPC, carOwnerDarc.getId());
                    Darc newOwnerDarc = getPersonDarc(changeOwnerButton.getText(), "user");
                    Darc newDarcO = carOwnerDarcInstance.getDarc();
                    newDarcO.setRule(Darc.RuleSignature, newOwnerDarc.getIdentity().toString().getBytes());
                    newDarcO.setRule("invoke:evolve", newOwnerDarc.getIdentity().toString().getBytes());
                    carOwnerDarcInstance.evolveDarcAndWait(newDarcO,
                            getPersonSigner(IndexController.role, "user"), 10);

                    //we also need to replace the previous owner from the read and garage darc with the new owner

//                    newDarcR = carReaderDarcInstance.getDarc();
//                    String newReaderExp = replaceSigner(carReaderDarcInstance,
//                            getPersonDarc(IndexController.role, "user").getIdentity(),
//                            getPersonDarc(changeOwnerButton.getText(), "user").getIdentity());
//                    newDarcR.setRule(Darc.RuleSignature, newReaderExp.getBytes());
//                    carReaderDarcInstance.evolveDarcAndWait(newDarcR,
//                            getPersonSigner(changeOwnerButton.getText(), "user"), 5);
//
//
//                    newDarcG = carGarageDarcInstance.getDarc();
//                    String newGarageExp = replaceSigner(carGarageDarcInstance,
//                            getPersonDarc(IndexController.role, "user").getIdentity(),
//                            getPersonDarc(changeOwnerButton.getText(), "user").getIdentity());
//                    newDarcG.setRule(Darc.RuleSignature, newGarageExp.getBytes());
//                    carGarageDarcInstance.evolveDarcAndWait(newDarcG,
//                            getPersonSigner(changeOwnerButton.getText(), "user"), 5);


                }
            }

            Main.window.setScene(signUpResultScene);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
    }

    private void onVINChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        chooseVinButton.setText(Identity);
    }
//
//    private String replaceSigner(DarcInstance darcInstance, Identity oldIdentity, Identity newIdentity) throws Exception{
//
//        Darc newDarc = darcInstance.getDarc();
//        String expr = newDarc.getExpression("_sign").toString();
//
//        if (expr.contains(oldIdentity.toString() + " | ") ){
//
//            expr.replaceAll(oldIdentity.toString() + " | ", newIdentity.toString() + " | ");
//            return expr;
//        }
//
//        if (expr.contains(" | " + oldIdentity.toString()) ){
//
//            expr.replaceAll(" | " + oldIdentity.toString(), " | " + newIdentity.toString());
//            return expr;
//        }
//
//        if (expr.contains(oldIdentity.toString()) ){
//
//            expr.replaceAll(oldIdentity.toString(), newIdentity.toString());
//            return expr;
//        }
//
//        return expr;
//    }

}
