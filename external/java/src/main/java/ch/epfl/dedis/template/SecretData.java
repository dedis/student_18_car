package ch.epfl.dedis.template;

import ch.epfl.dedis.byzcoin.transaction.Argument;
import ch.epfl.dedis.template.proto.CarProto;


import java.util.Arrays;

/**
 * SecretData is one element of the Report. It holds a key of type string
 * and a value of type byte[].
 */
public class SecretData {
    private String ecoScore;
    private String mileage;
    private Boolean warranty;
    private String checkNote;

    /**
     * Create a KeyValue object given its protobuf representation.
     *
     * @param sd the protobuf representation of the SecretData
     */
    public SecretData(CarProto.SecretData sd) {
        ecoScore = sd.getEcoscore();
        mileage = sd.getMileage();
        warranty = sd.getWarranty();
        checkNote = sd.getChecknote();
    }

    /**
     * Create a SecretData object given the ecoScore, mileage, warranty and checkNote
     *
     * @param ecoScore the ecoScore for the object
     * @param mileage the mileage for the object
     * @param warranty the warranty for the object
     * @param checkNote the checkNote for the object
     */
    public SecretData(String ecoScore, String mileage, Boolean warranty, String checkNote) {
        this.ecoScore = ecoScore;
        this.mileage = mileage;
        this.warranty = warranty;
        this.checkNote = checkNote;
    }

    /**
     * @return the ecoScore of the object.
     */
    public String getEcoScore() {
        return ecoScore;
    }

    /**
     * @return the mileage of the object.
     */
    public String getMileage() {
        return mileage;
    }

    /**
     * @return the warranty of the object.
     */
    public Boolean getWarranty() {
        return warranty;
    }

    /**
     * @return the checkNote of the object.
     */
    public String getCheckNote() {
        return checkNote;
    }


    /**
     * @param ecoScore the new ecoScore
     */
    public void setEcoScore(String ecoScore) {
        this.ecoScore = ecoScore;
    }

    /**
     * @param mileage the new mileage
     */
    public void setMileage(String mileage) {
        this.mileage = mileage;
    }

    /**
     * @param warranty the new warranty
     */
    public void setWarranty(Boolean warranty) {
        this.warranty = warranty;
    }

    /**
     * @param checkNote the new checkNote
     */
    public void setCheckNote(String checkNote) {
        this.checkNote = checkNote;
    }

    /**
     * Converts this object to the protobuf representation.
     * @return The protobuf representation.
     */
    public CarProto.SecretData toProto() {
        CarProto.SecretData.Builder b = CarProto.SecretData.newBuilder();
        b.setEcoscore(this.ecoScore);
        b.setMileage(this.mileage);
        b.setWarranty(this.warranty);
        b.setChecknote(this.checkNote);
        return b.build();
    }

    /**
     * @return an argument representing the key/value pair.
     */
    //TODO i don't think i need this
    //public Argument toArgument() {
      //  return new Argument(getKey(), getValue());
    //}

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        SecretData sd = (SecretData) o;

        return ecoScore.equals(sd.getEcoScore()) && mileage.equals(sd.getMileage())
                && warranty.equals(sd.getWarranty()) && checkNote.equals(sd.getCheckNote());
    }
}
