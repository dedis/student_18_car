package ch.epfl.dedis.template.gui.home;

import ch.epfl.dedis.template.gui.index.Main;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.Button;
import javafx.scene.control.Label;
import javafx.scene.image.Image;
import javafx.scene.image.ImageView;

import java.io.File;
import java.net.URL;
import java.util.ResourceBundle;

public class HomeController implements Initializable {



    @FXML
    private Button addReportButton;

    @FXML
    private Button readHistoryButton;

    @FXML
    private Button signUpButton;

    @FXML
    private Label sclogoLabel;

    @FXML
    private Label aslogoLabel;



    @Override
    public void initialize(URL location, ResourceBundle resources){

        try {

            URL urlsc = new File("src/main/java/ch/epfl/dedis/template/gui/index/sclogo.png").toURL();
            Image sclogo = new Image(urlsc.openStream(), 50, 50, true, true);
            ImageView viewSC = new ImageView(sclogo);
            sclogoLabel.setGraphic(viewSC);

            URL urlas = new File("src/main/java/ch/epfl/dedis/template/gui/index/autosenselogo.png").toURL();
            Image aslogo = new Image(urlas.openStream(), 45, 45, true, true);
            ImageView viewAS = new ImageView(aslogo);
            aslogoLabel.setGraphic(viewAS);

            //signInButton.setStyle("-fx-base: #001155;");
            addReportButton.setStyle("-fx-text-fill: #001155");
            readHistoryButton.setStyle("-fx-text-fill: #001155");


            //signUpButton.setStyle("-fx-background-color: #001155");
            signUpButton.setStyle("-fx-text-fill: #001155");

            signUpButton.setOnAction(event -> {
                Main.window.setScene(Main.signUpScene);
            });

            addReportButton.setOnAction(event -> {
                Main.window.setScene(Main.addReportScene);
            });

            //addReportButton.setOnAction(this::signIn);

        }
        catch (Exception e) {
            System.out.println(e.toString());
        }

    }

    @FXML
    private void signIn(ActionEvent actionEvent) {

    }

}
