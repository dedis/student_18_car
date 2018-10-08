package ch.epfl.dedis.integration;

import ch.epfl.dedis.byzgen.OcsFactory;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;


public class ManualTestServerController extends TestServerController {
    @Override
    public int countRunningConodes() throws IOException, InterruptedException {
        Process p = Runtime.getRuntime().exec("pgrep conode");
        int returnCode = p.waitFor();
        if (returnCode != 0) {
            throw new IllegalStateException("unable to count running conodes");
        }
        return countLines(inputStreamToString(p.getInputStream()));
    }

    @Override
    public void startConode(int nodeNumber) throws InterruptedException, IOException {
        Runtime.getRuntime().exec("../scripts/start_4th_conode.sh");
        Thread.sleep(1000);
    }

    @Override
    public void killConode(int nodeNumber) throws IOException, InterruptedException {
        if (nodeNumber!=4) {
            throw new IllegalArgumentException("I'm a manual controller and I'm able only to kill node4");
        }

        // kill the last conode and try to make a request
        int exitValue = Runtime.getRuntime().exec("pkill -n conode").waitFor();

        if ( exitValue != 0 ) {
            throw new IllegalStateException("something is wrong I'm not able to kill node");
        }
    }

    @Override
    public List<OcsFactory.ConodeAddress> getConodes() {
        return Arrays.asList(
                new OcsFactory.ConodeAddress(buildURI("tcp://localhost:7002"), CONODE_PUB_1),
                new OcsFactory.ConodeAddress(buildURI("tcp://localhost:7004"), CONODE_PUB_2),
                new OcsFactory.ConodeAddress(buildURI("tcp://localhost:7006"), CONODE_PUB_3),
                new OcsFactory.ConodeAddress(buildURI("tcp://localhost:7008"), CONODE_PUB_4));
    }

    private static int countLines(String str){
        String[] lines = str.split("\r\n|\r|\n");
        return  lines.length;
    }

    private static String inputStreamToString(InputStream in) {
        return new BufferedReader(new InputStreamReader(in))
                .lines().collect(Collectors.joining("\n"));
    }
}
