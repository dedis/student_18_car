package ch.epfl.dedis.template.gui.index;

import ch.epfl.dedis.byzcoin.ByzCoinRPC;
import ch.epfl.dedis.byzcoin.contracts.DarcInstance;
import ch.epfl.dedis.calypso.CalypsoRPC;
import ch.epfl.dedis.lib.Roster;
import ch.epfl.dedis.lib.darc.Darc;
import ch.epfl.dedis.lib.darc.DarcId;
import ch.epfl.dedis.lib.darc.Signer;
import ch.epfl.dedis.lib.darc.SignerEd25519;
import ch.epfl.dedis.lib.proto.DarcProto;
import ch.epfl.dedis.lib.proto.OnetProto;
import ch.epfl.dedis.template.gui.json.ByzC;
import ch.epfl.dedis.template.gui.json.Person;
import com.cedarsoftware.util.io.JsonReader;
import com.cedarsoftware.util.io.JsonWriter;
import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.databind.ObjectMapper;
import javafx.application.Application;
import javafx.scene.Scene;
import javafx.stage.Stage;
import javafx.fxml.FXMLLoader;
import javafx.scene.Parent;

import java.io.*;
import java.net.URL;
import java.time.Duration;
import java.util.Arrays;

import static ch.epfl.dedis.byzcoin.contracts.DarcInstance.fromByzCoin;
import static ch.epfl.dedis.template.gui.ByzSetup.*;
import static java.time.temporal.ChronoUnit.MILLIS;

public class Main extends Application {

    public static Scene loginScene, signUpScene, signUpResultScene, addReportScene;
    public static Stage window;


//    public static void main(String[] args) {
//        launch(args);
//    }
    public static void main() {

    }

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

        ByzC byzC = setup();

        Darc adminDarc = getAdminDarc();

        CalypsoRPC calypsoRPC = CalypsoRPC.fromCalypso(getRoster(byzC), getByzId(byzC), getLTSId(byzC));

        DarcInstance darcInstance = fromByzCoin(calypsoRPC, adminDarc.getId());



        window = primaryStage;

        URL urlLogin = new File("src/main/java/ch/epfl/dedis/template/gui/index/index.fxml").toURL();
        Parent rootLogin = FXMLLoader.load(urlLogin);
        loginScene = new Scene(rootLogin, 600, 400);

        URL urlSignUp = new File("src/main/java/ch/epfl/dedis/template/gui/signUp/signUp.fxml").toURL();
        Parent rootSignUp = FXMLLoader.load(urlSignUp);
        signUpScene = new Scene(rootSignUp, 600, 400);

        URL urlSignUpResult = new File("src/main/java/ch/epfl/dedis/template/gui/signUp/signUpResult.fxml").toURL();
        Parent rootSignUpResult = FXMLLoader.load(urlSignUpResult);
        signUpResultScene = new Scene(rootSignUpResult, 600, 400);

        URL urlAddReport = new File("src/main/java/ch/epfl/dedis/template/gui/addReport/addReport.fxml").toURL();
        Parent rootAddReport = FXMLLoader.load(urlAddReport);
        addReportScene = new Scene(rootAddReport, 600, 400);

        //Parent root = FXMLLoader.load(getClass().getResource("index.fxml"));
        primaryStage.setTitle("Car Maintenance History");
        primaryStage.setScene(loginScene);

        primaryStage.show();

    }



}

