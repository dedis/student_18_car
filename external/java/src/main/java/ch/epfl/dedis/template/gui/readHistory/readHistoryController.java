package ch.epfl.dedis.template.gui.readHistory;

import ch.epfl.dedis.byzcoin.InstanceId;
import ch.epfl.dedis.calypso.*;
import ch.epfl.dedis.template.CarInstance;
import ch.epfl.dedis.template.Report;
import ch.epfl.dedis.template.SecretData;
import ch.epfl.dedis.template.gui.index.IndexController;
import ch.epfl.dedis.template.gui.index.Main;
import ch.epfl.dedis.template.gui.json.CarJson;
import com.fasterxml.jackson.databind.ObjectMapper;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.*;

import java.io.File;
import java.net.URL;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.ResourceBundle;

import static ch.epfl.dedis.template.gui.ByzSetup.*;
import static ch.epfl.dedis.template.gui.index.Main.calypsoRPC;
import static ch.epfl.dedis.template.gui.index.Main.homePath;
import static ch.epfl.dedis.template.gui.index.Main.signUpResultScene;

public class readHistoryController implements Initializable {

    @FXML
    private TextArea historyTextArea;

    @FXML
    private MenuButton chooseVINButton;

    @FXML
    private Button readHistoryButton;


    @Override
    public void initialize(URL location, ResourceBundle resources){

        try{
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
        }
        catch (Exception e){
            e.printStackTrace();
        }

        readHistoryButton.setOnAction(this::getHistory);
        readHistoryButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");
    }

    private void getHistory(ActionEvent event) {

        try{
            CarInstance ci = CarInstance.fromCalypso(calypsoRPC,  getCarInstanceId(chooseVINButton.getText()));

            List<Report> reports = ci.getReports();
            for(int i=0; i <reports.size(); i++){
                Report report = reports.get(i);
                InstanceId id = new InstanceId(report.getWriteInstanceID());
                WriteInstance wi = WriteInstance.fromCalypso(calypsoRPC, id);
                ReadInstance ri = new ReadInstance(calypsoRPC, wi,
                        Arrays.asList(getPersonSigner(IndexController.role, "reader")));
                DecryptKeyReply dkr = calypsoRPC.tryDecrypt(calypsoRPC.getProof(wi.getInstance().getId()),
                        calypsoRPC.getProof(ri.getInstance().getId()));
                // And derive the symmetric key, using the user's private key to decrypt it:
                byte[] keyMaterial = dkr.getKeyMaterial(getPersonSigner(IndexController.role,
                        "reader").getPrivate());

                //get the document back:
                Document doc = Document.fromWriteInstance(wi, keyMaterial);

                SecretData secretData = new SecretData(doc.getData());

                historyTextArea.appendText("Report by: " + report.getGarageId() + "\n");
                historyTextArea.appendText("Date: " + report.getDate()+ "\n");
                historyTextArea.appendText("Eco Score: " + secretData.getEcoScore() + "\n");
                historyTextArea.appendText("Mielage: " + secretData.getMileage() + "\n");
                historyTextArea.appendText("Check Note: " + secretData.getCheckNote() + "\n");
                historyTextArea.appendText("\n");

            }
        }
        catch (Exception e) {
            e.printStackTrace();
        }


    }

    private void onVINChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        chooseVINButton.setText(Identity);
    }

}
