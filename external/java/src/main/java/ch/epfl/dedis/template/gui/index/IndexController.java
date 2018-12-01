package ch.epfl.dedis.template.gui.index;

import javafx.event.ActionEvent;
import javafx.fxml.Initializable;
import javafx.scene.control.*;
import  javafx.fxml.FXML;
import javafx.scene.image.Image;
import javafx.scene.image.ImageView;

import java.io.File;
import java.net.URL;
import java.util.ResourceBundle;

import static ch.epfl.dedis.template.gui.ByzSetup.getAdminName;

public class IndexController implements Initializable {


    @FXML
    private Label sclogoLabel;

    @FXML
    private Label aslogoLabel;

    @FXML
    private MenuButton roleButton;

    @FXML
    private Button submitRoleButton;



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

            MenuItem item = new MenuItem(getAdminName());
            item.setOnAction(this::onRoleChange);
            roleButton.getItems().add(item);

        }
        catch (Exception e) {
            System.out.println(e.toString());
        }

    }

    private void onRoleChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        roleButton.setText(Identity);
    }

}
