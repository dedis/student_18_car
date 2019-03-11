package ch.epfl.dedis.template.gui.addReport;

import ch.epfl.dedis.byzcoin.Instance;
import ch.epfl.dedis.byzcoin.InstanceId;
import ch.epfl.dedis.byzcoin.Proof;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.calypso.CalypsoRPC;
import ch.epfl.dedis.calypso.Document;
import ch.epfl.dedis.calypso.WriteInstance;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.template.CarInstance;
import ch.epfl.dedis.template.Report;
import ch.epfl.dedis.template.SecretData;
import ch.epfl.dedis.template.gui.errorScene.ErrorSceneController;
import ch.epfl.dedis.template.gui.index.IndexController;
import ch.epfl.dedis.template.gui.index.Main;
import ch.epfl.dedis.template.gui.json.CarJson;
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
import static ch.epfl.dedis.template.gui.ByzSetup.*;
import static ch.epfl.dedis.template.gui.index.Main.calypsoRPC;
import static ch.epfl.dedis.template.gui.index.Main.homePath;
import static ch.epfl.dedis.template.gui.index.Main.signUpResultScene;

public class addReportController implements Initializable {

    @FXML
    private TextArea notesTextArea;

    @FXML
    private TextField scoreTextField;

    @FXML
    private MenuButton chooseVINButton;

    @FXML
    private TextField mileageTextField;

    @FXML
    private Button submitButton;

    @FXML
    private ToggleGroup warranty;

    @FXML
    private TextField roleID;

    BooleanProperty disable = new SimpleBooleanProperty();

    @Override
    public void initialize(URL location, ResourceBundle resources){

        disable.bind(Bindings.createBooleanBinding(() ->
                scoreTextField.getText().trim().isEmpty(), scoreTextField.textProperty()).or(Bindings.createBooleanBinding(() ->
                mileageTextField.getText().trim().isEmpty(), mileageTextField.textProperty())).or(Bindings.createBooleanBinding(() ->
                notesTextArea.getText().trim().isEmpty(), notesTextArea.textProperty())));

        try{

            roleID.setText("ID: " + IndexController.role);
            ObjectMapper mapper = new ObjectMapper();

            File carFile = new File(homePath + "/json/car.json");
            if(carFile.exists())
            {
                HashMap<String, CarJson> carMap = mapper.readValue(carFile, HashMap.class);

                for (HashMap.Entry<String, CarJson> entry : carMap.entrySet()) {
                    MenuItem VINItem = new MenuItem(entry.getKey());
                    VINItem.setOnAction(this::onVINChange);
                    chooseVINButton.getItems().add(VINItem);
                }
            }
            submitButton.setOnAction(this::submitReport);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }


        submitButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");
    }

    private void submitReport(ActionEvent event) {

        RadioButton selectedRadioButton = (RadioButton) warranty.getSelectedToggle();
        String war = selectedRadioButton.getText();
        boolean hasWarranty = (war.equals("Yes"));

        SecretData secret = new SecretData(scoreTextField.getText(),
                mileageTextField.getText(), hasWarranty, notesTextArea.getText());
        try{

            Document doc = new Document(secret.toProto().toByteArray(), 16,
                    null, getCarDarc(chooseVINButton.getText()).getBaseId());

            WriteInstance w = new WriteInstance(calypsoRPC, getCarDarc(chooseVINButton.getText()).getId(),
                    Arrays.asList(getPersonSigner(IndexController.role,
                            "garage")), doc.getWriteData(calypsoRPC.getLTS()));
            Proof p = calypsoRPC.getProof(w.getInstance().getId());

            List<Report> reports = new ArrayList<>();
            Report report = new Report(new Date().toString(),
                    IndexController.role, w.getInstance().getId().getId());
            reports.add(report);


            CarInstance ci = CarInstance.fromCalypso(calypsoRPC,  getCarInstanceId(chooseVINButton.getText()));
            ci.addReportAndWait(reports,
                    getPersonSigner(IndexController.role, "garage"), 10);

            Main.window.setScene(signUpResultScene);
            scoreTextField.setText("");
            mileageTextField.setText("");
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
        chooseVINButton.setText(Identity);
    }
}
