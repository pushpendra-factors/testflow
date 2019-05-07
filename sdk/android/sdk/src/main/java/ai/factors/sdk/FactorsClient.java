package ai.factors.sdk;

import android.content.Context;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.concurrent.atomic.AtomicBoolean;

public class FactorsClient implements Factors {

    /**
     * The class identifier used in logging.
     */
    private static final String TAG = "ai.factors.sdk.FactorsClient";

    private static final FactorsLogger logger = FactorsLogger.getLogger();
    private static FactorsClient instance = new FactorsClient();

    private AtomicBoolean uploading = new AtomicBoolean(false);


    /**
     * The background http worker thread instance.
     */
    private final WorkerThread httpThread = new WorkerThread("http-worker");

    /**
     * The background worker thread instance.
     */
    private final WorkerThread bgThread = new WorkerThread("bg-worker");

    /**
     * The token for factors project.
     */
    private String projectToken;

    private boolean initialized = false;

    /**
     * The Android App Context
     */
    private Context context;
    private HTTPHelper httpHelper;
    private DBHelper dbHelper;
    private DeviceInfo deviceInfo;

    private FactorsClient() {
    }

    public static Factors getInstance() {
        return instance;
    }

    public Factors init(final Context context, final String projectToken) {
        if (context == null) {
            return this;
        }

        if (Utils.isEmptyString(projectToken)) {
            return this;
        }

        if (initialized) {
            return this;
        }

        this.initialized = true;
        this.context = context;
        this.projectToken = projectToken;
        this.dbHelper = new DBHelper(this.context);
        this.httpHelper = new HTTPHelper();
        this.deviceInfo = new DeviceInfo(context);
        this.bgThread.start();
        this.httpThread.start();

        return this;
    }

    /**
     * Toggle message logging by the SDK.
     *
     * @param enableLogging whether to enable message logging by the SDK.
     * @return the FactorsClient
     */
    public Factors enableLogging(final boolean enableLogging) {
        logger.setEnableLogging(enableLogging);
        return this;
    }

    /**
     * Sets logging level. Log messages only if they are the same severity or higher than logLevel.
     *
     * @param logLevel the log level
     * @return the FactorsClient
     */
    public Factors setLogLevel(final int logLevel) {
        logger.setLogLevel(logLevel);
        return this;
    }

    @Override
    public void addUserProperties(final JSONObject properties) {

        if (!initialized) {
            return;
        }

        try {
            JSONObject payload = new JSONObject();

            String factorsUserID = dbHelper.getUserUUID();

            if (!Utils.isEmptyString(factorsUserID)) {
                payload.put(Constants.USER_ID, factorsUserID);
            }

            payload.put(Constants.PROPERTIES, Utils.mergeJSONObjects(properties, getDefaultUserProperties()));

            saveRequest(SDKRoute.ADD_USER_PROPERTIES, payload.toString());

        } catch (JSONException e) {
            logger.e(TAG, "Failed to create addUserProperties json payload", e);
        }
    }

    @Override
    public void identify(final String customerUserID) {

        if (!initialized) {
            return;
        }

        if (Utils.isEmptyString(customerUserID)) {
            return;
        }
        String factorsUserID = dbHelper.getUserUUID();
        final JSONObject payload = new JSONObject();
        try {
            if (!Utils.isEmptyString(factorsUserID)) {
                payload.put(Constants.USER_ID, factorsUserID);
            }
            saveRequest(SDKRoute.IDENTIFY, payload.toString());
        } catch (JSONException e) {
            logger.e(TAG, "Failed to create identify json payload", e);
        }
    }

    @Override
    public void track(final String eventName, final JSONObject eventProperties) {

        if (!initialized) {
            return;
        }

        if (Utils.isEmptyString(eventName)) {
            return;
        }

        String factorsUserID = dbHelper.getUserUUID();
        final JSONObject payload = new JSONObject();

        try {
            payload.put(Constants.EVENT_NAME, eventName);
            payload.put(Constants.EVENT_PROPERTIES, eventProperties);
            payload.put(Constants.USER_PROPERTIES, getDefaultUserProperties());

            if (!Utils.isEmptyString(factorsUserID)) {
                payload.put(Constants.USER_ID, factorsUserID);
            }

            saveRequest(SDKRoute.TRACK, payload.toString());

        } catch (JSONException e) {
            logger.e(TAG, "Failed to create track json payload", e);
        }
    }

    private void updateUserIDIfPresentInResponse(final JSONObject resp) {

        if (resp == null) {
            return;
        }

        if (!resp.has(Constants.USER_ID)) {
            return;
        }

        try {
            String factorsUserID = resp.getString(Constants.USER_ID);
            if (Utils.isEmptyString(factorsUserID)) {
                return;
            }
            this.dbHelper.storeUserUUID(factorsUserID);
        } catch (JSONException e) {
            logger.e(TAG, "updateUserIDIfPresentInResponse Failed to read user_id from jsonResp", e);
        }
    }

    private boolean hasEventsToUpload() {
        return dbHelper.getRequestsCount() > 0;
    }

    private boolean isStoreEventsLimitReached() {
        return dbHelper.getRequestsCount() > Constants.MAX_NO_OF_REQS;
    }

    private void saveRequest(final SDKRoute route, final String payload) {
        final FactorsClient fc = this;
        runOnBGThread(new Runnable() {
            @Override
            public void run() {
                if (isStoreEventsLimitReached()) {
                    // Start dropping events
                    logger.e(TAG, "saveRequest Failed to store request, MAX_NO_OF_REQS limit reached");
                    return;
                }
                fc.dbHelper.storeRequest(route.toString(), payload);
                if (!fc.uploading.get()) {
                    fc.executeRequests();
                }
            }
        });
    }

    private boolean isNetworkAvailable() {
        ConnectivityManager cm =
                (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);

        NetworkInfo activeNetwork = cm.getActiveNetworkInfo();
        return activeNetwork != null && activeNetwork.isConnectedOrConnecting();
    }

    private void executeRequests() {
        final FactorsClient fc = this;
        runOnHTTPThread(new Runnable() {
            @Override
            public void run() {
                fc.uploading.set(true);
                while (isNetworkAvailable() && fc.hasEventsToUpload()) {
                    RequestModel req = fc.dbHelper.getNextRequest();
                    final String url = String.format("%s%s", Constants.SEVER, SDKRoute.valueOf(req.type).getRoute());
                    JSONObject resp = fc.httpHelper.executeRequest(url, fc.projectToken, req.payload);
                    fc.dbHelper.deleteRequest(req.ID);
                    updateUserIDIfPresentInResponse(resp);
                }
                fc.uploading.set(false);
            }
        });
    }

    private void runOnBGThread(Runnable r) {
        if (Thread.currentThread() != bgThread) {
            bgThread.post(r);
        } else {
            r.run();
        }
    }

    private void runOnHTTPThread(Runnable r) {
        if (Thread.currentThread() != httpThread) {
            httpThread.post(r);
        } else {
            r.run();
        }
    }

    private JSONObject getDefaultUserProperties() {
        JSONObject userProperties = new JSONObject();
        try {
            userProperties.put(Constants.PLATFORM, Constants.MOBILE);
            userProperties.put(Constants.OS, DeviceInfo.getOsName());
            userProperties.put(Constants.OS_VERSION, DeviceInfo.getOsVersion());
            userProperties.put(Constants.DEVICE_MODEL, DeviceInfo.getModel());
            userProperties.put(Constants.DEVICE_BRAND, DeviceInfo.getBrand());
            userProperties.put(Constants.DEVICE_MANUFACTURER, DeviceInfo.getManufacturer());
            userProperties.put(Constants.NETWORK_CARRIER, this.deviceInfo.getNetworkCarrier());

            return userProperties;
        } catch (JSONException e) {
            logger.e(TAG, "Failed to create defaultUserProperties json payload", e);
        }
        return new JSONObject();
    }
}
