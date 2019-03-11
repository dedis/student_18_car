package ch.epfl.dedis.template.gui.readHistory;

import ch.epfl.dedis.byzcoin.InstanceId;
import ch.epfl.dedis.calypso.*;
import ch.epfl.dedis.template.CarInstance;
import ch.epfl.dedis.template.Report;
import ch.epfl.dedis.template.SecretData;
import ch.epfl.dedis.template.gui.errorScene.ErrorSceneController;
import ch.epfl.dedis.template.gui.index.IndexController;
import ch.epfl.dedis.template.gui.index.Main;
import ch.epfl.dedis.template.gui.json.CarJson;
import com.fasterxml.jackson.databind.ObjectMapper;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.FXMLLoader;
import javafx.fxml.Initializable;
import javafx.scene.Parent;
import javafx.scene.Scene;
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

    @FXML
    private Button goBackButton;

    @FXML
    private TextField roleID;

    @Override
    public void initialize(URL location, ResourceBundle resources){

        goBackButton.setOnAction(this::switchScene);
        goBackButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");

        try{
            ObjectMapper mapper = new ObjectMapper();
            roleID.setText("ID: " + IndexController.role);
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
            readHistoryButton.setOnAction(this::getHistory);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }

        readHistoryButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");
    }

    private void switchScene(ActionEvent event) {

        try {
            //refreshing scenes
            URL url = new File("src/main/java/ch/epfl/dedis/template/gui/index/index.fxml").toURL();
            Parent Index = FXMLLoader.load(url);
            Scene iScene = new Scene(Index, 600, 400);

            URL urlSignUp = new File("src/main/java/ch/epfl/dedis/template/gui/Register/Register.fxml").toURL();
            Parent rootSignUp = FXMLLoader.load(urlSignUp);
            Main.adminScene = new Scene(rootSignUp, 600, 400);

            URL urlUserScreen = new File("src/main/java/ch/epfl/dedis/template/gui/userScreen/userScreen.fxml").toURL();
            Parent rootUserScreen = FXMLLoader.load(urlUserScreen);
            Main.userScreenScene = new Scene(rootUserScreen, 600, 400);

            URL urlAddReport = new File("src/main/java/ch/epfl/dedis/template/gui/addReport/addReport.fxml").toURL();
            Parent rootAddReport = FXMLLoader.load(urlAddReport);
            Main.addReportScene = new Scene(rootAddReport, 600, 400);

            URL urlReadHistory = new File("src/main/java/ch/epfl/dedis/template/gui/readHistory/readHistory.fxml").toURL();
            Parent rootReadHistory = FXMLLoader.load(urlReadHistory);
            Main.readHistoryScene = new Scene(rootReadHistory, 600, 400);

            Main.window.setScene(iScene);
        }
        catch (Exception e){
            Main.errorMsg = e.toString();
            Main.loadErrorScene();
            Main.window.setScene(Main.errorScene);
            e.printStackTrace();
        }
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
