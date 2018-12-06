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

import static ch.epfl.dedis.template.gui.ByzSetup.*;

public class IndexController implements Initializable {


    @FXML
    private Label sclogoLabel;

    @FXML
    private Label aslogoLabel;

    @FXML
    public MenuButton roleButton;

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

            MenuItem adminItem = new MenuItem(getAdminName());
            adminItem.setOnAction(this::onRoleChange);
            roleButton.getItems().add(adminItem);

            if(getUserName()!=null)
            {
                MenuItem userItem = new MenuItem(getUserName());
                userItem.setOnAction(this::onRoleChange);
                roleButton.getItems().add(userItem);
            }

            if(getReaderName()!=null)
            {
                MenuItem readerItem = new MenuItem(getReaderName());
                readerItem.setOnAction(this::onRoleChange);
                roleButton.getItems().add(readerItem);
            }

            if(getGarageName()!=null)
            {
                MenuItem garageItem = new MenuItem(getGarageName());
                garageItem.setOnAction(this::onRoleChange);
                roleButton.getItems().add(garageItem);
            }

            submitRoleButton.setStyle("-fx-background-color: #001155; -fx-text-fill: white");
            //submitRoleButton.setStyle("-fx-text-fill: white");
            submitRoleButton.setOnAction(this::onRoleSubmit);
        }
        catch (Exception e) {
            e.printStackTrace();
        }
    }

    private void onRoleSubmit(ActionEvent event)
    {
        try{
            if (roleButton.getText().equals(getAdminName())){

                Main.window.setScene(Main.adminScene);
            }
            else if (roleButton.getText().equals(getUserName())){

                Main.window.setScene(Main.userScreenScene);
            }
            else if(roleButton.getText().equals(getReaderName())){

                Main.window.setScene(Main.readHistoryScene);
            }
            else if(roleButton.getText().equals(getGarageName())){

                Main.window.setScene(Main.addReportScene);
            }
            roleButton.setText("Identity");
        }
        catch (Exception e) {
            e.printStackTrace();
        }
    }

    private void onRoleChange(ActionEvent event) {
        String Identity = ((MenuItem) event.getTarget()).getText();
        roleButton.setText(Identity);
    }

}
