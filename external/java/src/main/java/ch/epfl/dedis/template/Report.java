package ch.epfl.dedis.template;

import ch.epfl.dedis.byzcoin.transaction.Argument;
import ch.epfl.dedis.template.proto.CarProto;
import ch.epfl.dedis.byzcoin.Instance;
import com.google.protobuf.ByteString;

import java.util.Arrays;

/**
 * Report is one element of the Car instance. It holds a key of type string
 * and a value of type byte[].
 */
public class Report {
    private String date;
    private String garageId;
    private byte[] writeInstanceID;

    /**
     * Create a Report object given its protobuf representation.
     *
     * @param rep the protobuf representation of the Report
     */
    public Report(CarProto.Report rep) {
        date = rep.getDate();
        garageId = rep.getGarageid();
        writeInstanceID = rep.getWriteinstanceid().toByteArray();
    }

    /**
     * Create a Report object given the date and garageId
     *
     * @param date the date for the object
     * @param garageId the garageId for the object
     */
    public Report(String date, String garageId, byte[] writeInstanceID) {
        this.date = date;
        this.garageId = garageId;
        this.writeInstanceID = writeInstanceID;
    }

    /**
     * @return the date of the object.
     */
    public String getDate() {
        return date;
    }

    /**
     * @return the garageID of the object.
     */
    public String getGarageId() {
        return garageId;
    }

    /**
     * @return the writeInstance of the object.
     */
    public byte[] getWriteInstanceID() {
        return writeInstanceID;
    }

    /**
     * @param date the new date
     */
    public void setDate(String date) {
        this.date = date;
    }

    /**
     * @param garageId the new garageId
     */
    public void setGarageId(String garageId) {
        this.garageId = garageId;
    }

    /**
     * @param writeInstanceID the new writeInstanceID
     */
    public void setWriteInstanceID(byte[] writeInstanceID) {
        this.writeInstanceID = writeInstanceID;
    }

    /**
     * Converts this object to the protobuf representation.
     * @return The protobuf representation.
     */
    public CarProto.Report toProto() {
        CarProto.Report.Builder b = CarProto.Report.newBuilder();
        b.setDate(this.date);
        b.setGarageid(this.garageId);
        ByteString s= ByteString.copyFrom(this.writeInstanceID);
        b.setWriteinstanceid(s);
        return b.build();
    }

    /**
     * @return an argument representing the key/value pair.
     */
    //TODO how to encode the report into []byte
    public Argument toArgument() {
        return new Argument("report", this.toProto().toByteArray());
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        Report rep = (Report) o;

        //TODO how to compare instances
        return date.equals(rep.getDate()) && garageId.equals(rep.getGarageId());
    }
}
