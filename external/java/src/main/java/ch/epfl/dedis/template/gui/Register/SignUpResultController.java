package ch.epfl.dedis.template.gui.Register;

import ch.epfl.dedis.template.gui.index.IndexController;
import ch.epfl.dedis.template.gui.index.Main;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.FXMLLoader;
import javafx.fxml.Initializable;
import javafx.scene.Parent;
import javafx.scene.Scene;
import javafx.scene.control.Button;

import java.io.File;
import java.net.URL;
import java.util.ResourceBundle;

public class SignUpResultController implements Initializable {


    @FXML
    private Button backButton;

    @Override
    public void initialize(URL location, ResourceBundle resources){

            backButton.setOnAction(this::switchScene);
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


            Main.window.setScene(iScene);
        }
        catch (Exception e){
            e.printStackTrace();
        }
    }


}
