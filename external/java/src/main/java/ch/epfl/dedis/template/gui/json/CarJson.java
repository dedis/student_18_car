package ch.epfl.dedis.template.gui.json;

public class CarJson {
    public String VIN;
    public byte[] darc;
    public byte[] ownerDarc;
    public byte[] readerDarc;
    public byte[] garageDarc;
    public byte[] instanceId;


    public CarJson(String VIN, byte[] darc, byte[] ownerDarc, byte[] readerDarc,
                   byte[] garageDarc, byte[] instanceId) {
        this.VIN = VIN;
        this.darc = darc;
        this.ownerDarc = ownerDarc;
        this.readerDarc = readerDarc;
        this.garageDarc = garageDarc;
        this.instanceId = instanceId;
    }

    public CarJson(){
    }
}
