package ch.epfl.dedis.template.gui.addReport;

import ch.epfl.dedis.template.gui.index.Main;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.Button;
import javafx.scene.control.TextField;

import java.net.URL;
import java.util.ResourceBundle;

public class addReportController implements Initializable {

    @FXML
    private TextField vinTextField;


    @Override
    public void initialize(URL location, ResourceBundle resources){

    vinTextField.promptTextProperty().setValue("1FMCU59H98KA04664");

    }


}
