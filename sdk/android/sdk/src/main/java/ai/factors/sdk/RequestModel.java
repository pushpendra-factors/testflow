package ai.factors.sdk;

public class RequestModel {
    long ID;
    String payload;
    String type;

    public RequestModel(long ID, String payload, String type) {
        this.ID = ID;
        this.payload = payload;
        this.type = type;
    }

}