package ch.epfl.dedis.template.gui.json;

public class Person {
    public String name;
    public byte[] darc;
    public String publicKey;
    public String privateKey;


    public Person(String name, byte[] darc, String publicKey, String privateKey){
        this.name = name;
        this.darc = darc;
        this.publicKey = publicKey;
        this.privateKey = privateKey;
    }

    public  Person(){

    }

}
