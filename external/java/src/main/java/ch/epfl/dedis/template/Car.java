package ch.epfl.dedis.template;

import ch.epfl.dedis.byzcoin.Instance;
import ch.epfl.dedis.template.proto.CarProto;
import com.google.protobuf.InvalidProtocolBufferException;
import ch.epfl.dedis.byzcoin.transaction.Argument;
import com.google.protobuf.ByteString;


import java.util.ArrayList;
import java.util.List;

/**
 * Car represents the data stored in a Car instance. A Car instance
 * stores a VIN(vehicle id number) attribute and a list of reports and lets you add reports.
 */
public class Car {

    private String vin;
    private List<Report> reportList;

    /**
     * Create a Car object given its protobuf representation.
     *
     * @param carProto the protobuf representation.
     */
    public Car(CarProto.Car carProto) {
        this.vin = carProto.getVin();
        reportList = new ArrayList<>();
        for (CarProto.Report report : carProto.getReportsList()) {
            reportList.add(new Report(report));
        }
    }

    /**
     * Create a Car object given its binary representation of the protobuf.
     *
     * @param data binary representation of the protobuf
     * @throws InvalidProtocolBufferException
     */
    public Car(byte[] data) throws InvalidProtocolBufferException {
        this(CarProto.Car.parseFrom(data));
    }

    /**
     * Create a Car object given an instance.
     *
     * @param inst the instance that holds the KeyValueData
     * @throws InvalidProtocolBufferException
     */
    public Car(Instance inst) throws InvalidProtocolBufferException {
        this(inst.getData());
    }

    /**
     * Returns a copy of the Report list.
     *
     * @return a copy of the Report list.
     */
    public List<Report> getReportsList() {
        List<Report> repCopy = new ArrayList<>();
        for (Report rep : reportList) {
            repCopy.add(rep);
        }
        return repCopy;
    }

    /**
     * Returns the vin.
     *
     * @return vin.
     */
    public String getVin() {
        return vin;
    }


    /**
     * Converts this object to the protobuf representation.
     * @return The protobuf representation.
     */
    public CarProto.Car toProto() {
        CarProto.Car.Builder b = CarProto.Car.newBuilder();
        b.setVin(this.vin);
        List<CarProto.Report> reports = new ArrayList<>();
        for (int i=0; i<this.reportList.size();i++){
            reports.add(reportList.get(i).toProto());
        }
        b.addAllReports(reports);
        return b.build();
    }

    /**
     * @return an argument representing the car.
     */
    //TODO encode the car into []byte
    public Argument toArgument() {
      return new Argument("car", this.toProto().toByteArray());
    }

}
