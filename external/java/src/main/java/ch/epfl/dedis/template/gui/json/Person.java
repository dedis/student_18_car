package ch.epfl.dedis.template.gui.json;

import ch.epfl.dedis.lib.darc.SignerEd25519;

public class Person {
    public String name;
    public byte[] darc;
    public byte[] signer;


    public Person(String name, byte[] darc, byte[] signer){
        this.name = name;
        this.darc = darc;
        this.signer = signer;
    }

    public  Person(){
    }

}
