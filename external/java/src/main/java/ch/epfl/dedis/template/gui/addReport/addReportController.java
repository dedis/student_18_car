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
import ch.epfl.dedis.template.gui.index.Main;
import javafx.beans.binding.Bindings;
import javafx.beans.property.BooleanProperty;
import javafx.beans.property.SimpleBooleanProperty;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.*;

import java.net.URL;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.ResourceBundle;

import static ch.epfl.dedis.byzcoin.contracts.DarcInstance.fromByzCoin;
import static ch.epfl.dedis.template.gui.ByzSetup.*;
import static ch.epfl.dedis.template.gui.index.Main.calypsoRPC;
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

    BooleanProperty disable = new SimpleBooleanProperty();


    @Override
    public void initialize(URL location, ResourceBundle resources){


        disable.bind(Bindings.createBooleanBinding(() ->
                scoreTextField.getText().trim().isEmpty(), scoreTextField.textProperty()).or(Bindings.createBooleanBinding(() ->
                mileageTextField.getText().trim().isEmpty(), mileageTextField.textProperty())).or(Bindings.createBooleanBinding(() ->
                notesTextArea.getText().trim().isEmpty(), notesTextArea.textProperty())));

        try{
            if(getVIN()!=null)
            {
                MenuItem VINItem = new MenuItem(getVIN());
                VINItem.setOnAction(this::onVINChange);
                chooseVINButton.getItems().add(VINItem);
            }
        }
        catch (Exception e){
            e.printStackTrace();
        }

        submitButton.setOnAction(this::submitReport);
    }

    private void submitReport(ActionEvent event) {

        RadioButton selectedRadioButton = (RadioButton) warranty.getSelectedToggle();
        String war = selectedRadioButton.getText();
        boolean hasWarranty = (war.equals("Yes"));

        SecretData secret = new SecretData(scoreTextField.getText(),
                mileageTextField.getText(), hasWarranty, notesTextArea.getText());
        try{

            Document doc = new Document(secret.toProto().toByteArray(), 16, null, getCarDarc().getBaseId());

            WriteInstance w = new WriteInstance(calypsoRPC, getCarDarc().getId(), Arrays.asList(getGarage()), doc.getWriteData(calypsoRPC.getLTS()));
            Proof p = calypsoRPC.getProof(w.getInstance().getId());

            List<Report> reports = new ArrayList<>();
            Report report = new Report("15.02.1994", getGarageName(), w.getInstance().getId().getId());
            reports.add(report);


            CarInstance ci = CarInstance.fromCalypso(calypsoRPC,  getCarInstanceId());
            ci.addReportAndWait(reports, getGarage(), 10);
        }
        catch (Exception e){
            e.printStackTrace();
        }

        Main.window.setScene(signUpResultScene);
        scoreTextField.setText("");
        mileageTextField.setText("");
    }

    private void onVINChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        chooseVINButton.setText(Identity);
    }


}
