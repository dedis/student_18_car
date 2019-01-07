package ch.epfl.dedis.template.gui.userScreen;

import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.Rules;
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
import java.util.HashMap;
import java.util.ResourceBundle;

import static ch.epfl.dedis.byzcoin.contracts.DarcInstance.fromByzCoin;
import static ch.epfl.dedis.template.gui.ByzSetup.*;
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
    private ListView<CheckBox> garagesList;

    @Override
    public void initialize(URL location, ResourceBundle resources){


        readersList.getSelectionModel().setSelectionMode(SelectionMode.MULTIPLE);
        garagesList.getSelectionModel().setSelectionMode(SelectionMode.MULTIPLE);

        try{
            ObjectMapper mapper = new ObjectMapper();

            File carFile = new File(homePath + "/json/car.json");
            if(carFile.exists())
            {
                HashMap<String, CarJson> carMap = mapper.readValue(carFile, HashMap.class);

                for (HashMap.Entry<String, CarJson> entry : carMap.entrySet()) {
                    MenuItem VINItem = new MenuItem(entry.getKey());
                    VINItem.setOnAction(this::onVINChange);
                    chooseVinButton.getItems().add(VINItem);
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
            e.printStackTrace();
        }
        
        updateButton.setOnAction(this::update);
        updateButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");

    }


    //todo make it work with removal of access
    private void update(ActionEvent event) {
        ObservableList<CheckBox> readers = readersList.getItems();
        ObservableList<CheckBox> garages = garagesList.getItems();
        try {

            for (CheckBox reader : readers) {
                if (reader.isSelected()) {

                    Darc carReaderDarc = getCarReaderDarc(chooseVinButton.getText());
                    DarcInstance carReaderDarcInstance = fromByzCoin(calypsoRPC, carReaderDarc.getId());
                    carReaderDarc.addIdentity(Darc.RuleSignature,
                            getPersonDarc(reader.getText(), "reader").getIdentity(), Rules.OR);
                    carReaderDarcInstance.evolveDarcAndWait(carReaderDarc,
                            getPersonSigner(IndexController.role, "user"), 10);
                }
            }

            for (CheckBox garage : garages) {
                if (garage.isSelected()) {

                    Darc carGarageDarc = getCarGarageDarc(chooseVinButton.getText());
                    DarcInstance carGarageDarcInstance = fromByzCoin(calypsoRPC, carGarageDarc.getId());
                    carGarageDarc.addIdentity(Darc.RuleSignature,
                            getPersonDarc(garage.getText(), "garage").getIdentity(), Rules.OR);
                    carGarageDarcInstance.evolveDarcAndWait(carGarageDarc,
                            getPersonSigner(IndexController.role, "user"), 10);
                }
            }
        }
        catch (Exception e){
            e.printStackTrace();
        }

        Main.window.setScene(signUpResultScene);
    }

    private void onVINChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        chooseVinButton.setText(Identity);
    }

}
