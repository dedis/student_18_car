package ch.epfl.dedis.template.gui.signUp;

import ch.epfl.dedis.template.gui.index.Main;
import javafx.beans.binding.Bindings;
import javafx.beans.property.BooleanProperty;
import javafx.beans.property.SimpleBooleanProperty;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.fxml.Initializable;
import javafx.scene.control.*;

import java.net.URL;
import java.util.ResourceBundle;

public class SignUpController implements Initializable {


    @FXML
    private TextField vinField;

    @FXML
    private ToggleGroup roleGroup;

    @FXML
    private TextField emailField;

    @FXML
    private Button submitButton;

    BooleanProperty disable = new SimpleBooleanProperty();



    @Override
    public void initialize(URL location, ResourceBundle resources){

        disable.bind(roleGroup.selectedToggleProperty().isNull().or(Bindings.createBooleanBinding(() ->
                        emailField.getText().trim().isEmpty(), emailField.textProperty())).or(Bindings.createBooleanBinding(() ->
                vinField.getText().trim().isEmpty(), vinField.textProperty())));

        submitButton.disableProperty().bind(disable);

        vinField.promptTextProperty().setValue("1FMCU59H98KA04664");

        emailField.promptTextProperty().setValue("email@example.com");

        submitButton.setOnAction(event -> {
            Main.window.setScene(Main.signUpResultScene);
        });

    }

    @FXML
    void onSelection(ActionEvent event){
        RadioButton r = (RadioButton)event.getTarget();

        if (r.isSelected())
        {
            System.out.println(r.getText());
        }
    }


    @FXML
    void signUp(ActionEvent event) {
        RadioButton selectedRole = (RadioButton) roleGroup.getSelectedToggle();

        if (selectedRole.getText() == "Garage")
        {

        }

    }

}
