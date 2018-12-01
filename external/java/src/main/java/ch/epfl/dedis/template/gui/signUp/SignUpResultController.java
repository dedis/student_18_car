package ch.epfl.dedis.template.gui.signUp;

import ch.epfl.dedis.template.gui.index.Main;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.Button;

import java.net.URL;
import java.util.ResourceBundle;

public class SignUpResultController implements Initializable {


    @FXML
    private Button backButton;

    @Override
    public void initialize(URL location, ResourceBundle resources){

        backButton.setOnAction(event -> {
            Main.window.setScene(Main.loginScene);
        });


    }


}
