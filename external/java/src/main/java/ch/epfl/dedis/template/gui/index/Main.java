package ch.epfl.dedis.template.gui.index;

import ch.epfl.dedis.calypso.CalypsoRPC;
import ch.epfl.dedis.template.gui.json.ByzC;
import javafx.application.Application;
import javafx.scene.Scene;
import javafx.stage.Stage;
import javafx.fxml.FXMLLoader;
import javafx.scene.Parent;

import java.io.*;
import java.net.URL;

import static ch.epfl.dedis.template.gui.ByzSetup.*;

public class Main extends Application {

    public static Scene indexScene, adminScene, signUpResultScene, addReportScene, userScreenScene,
            readHistoryScene, errorScene;
    public static Stage window;

    public static ByzC byzC;
    public static CalypsoRPC calypsoRPC;
    public static String homePath = System.getProperty("user.home");

    public static String errorMsg = "";

    /**
     * Starts the ByzCoin Blockchain when it's running for the first time,
     * or reads the config files in order to connect to it.
     *
     * Then, we start the GUI
     *
     * @param primaryStage
     * @throws Exception
     */
    @Override
    public void start(Stage primaryStage) throws Exception {

        byzC = setup();
        //connects to an existing byzcoin and an existing Long Term Secret
        calypsoRPC = CalypsoRPC.fromCalypso(getRoster(byzC), getByzId(byzC), getLTSId(byzC));

        window = primaryStage;

        URL urlindex = new File("src/main/java/ch/epfl/dedis/template/gui/index/index.fxml").toURL();
        Parent rootIndex = FXMLLoader.load(urlindex);
        indexScene = new Scene(rootIndex, 600, 400);

        URL urlSignUp = new File("src/main/java/ch/epfl/dedis/template/gui/Register/Register.fxml").toURL();
        Parent rootSignUp = FXMLLoader.load(urlSignUp);
        adminScene = new Scene(rootSignUp, 600, 400);

        URL urlSignUpResult = new File("src/main/java/ch/epfl/dedis/template/gui/Register/signUpResult.fxml").toURL();
        Parent rootSignUpResult = FXMLLoader.load(urlSignUpResult);
        signUpResultScene = new Scene(rootSignUpResult, 600, 400);

        URL urlUserScreen = new File("src/main/java/ch/epfl/dedis/template/gui/userScreen/userScreen.fxml").toURL();
        Parent rootUserScreen = FXMLLoader.load(urlUserScreen);
        userScreenScene = new Scene(rootUserScreen, 600, 400);

        URL urlAddReport = new File("src/main/java/ch/epfl/dedis/template/gui/addReport/addReport.fxml").toURL();
        Parent rootAddReport = FXMLLoader.load(urlAddReport);
        addReportScene = new Scene(rootAddReport, 600, 400);

        URL urlReadHistory = new File("src/main/java/ch/epfl/dedis/template/gui/readHistory/readHistory.fxml").toURL();
        Parent rootReadHistory = FXMLLoader.load(urlReadHistory);
        readHistoryScene = new Scene(rootReadHistory, 600, 400);

        URL urlErrorScene = new File("src/main/java/ch/epfl/dedis/template/gui/errorScene/errorScene.fxml").toURL();
        Parent rootErrorScene = FXMLLoader.load(urlErrorScene);
        errorScene = new Scene(rootErrorScene, 600, 400);

        //Parent root = FXMLLoader.load(getClass().getResource("index.fxml"));
        primaryStage.setTitle("Car Maintenance History");
        primaryStage.setScene(indexScene);

        primaryStage.show();

    }

    public static void loadErrorScene(){
        try {
            URL urlErrorScene = new File("src/main/java/ch/epfl/dedis/template/gui/errorScene/errorScene.fxml").toURL();
            Parent rootErrorScene = FXMLLoader.load(urlErrorScene);
            errorScene = new Scene(rootErrorScene, 600, 400);
        }
        catch (Exception e) {
            errorMsg = e.toString();
            e.printStackTrace();
        }
    }

}

