package ai.factors.sdk;

import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;

import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

enum SDKRoute {
    TRACK("/sdk/event/track"), IDENTIFY("/sdk/user/identify"), ADD_USER_PROPERTIES("/sdk/user/add_properties");
    private String route;

    SDKRoute(String route) {
        this.route = route;
    }

    public String getRoute() {
        return this.route;
    }
}

class HTTPHelper {

    private static final FactorsLogger logger = FactorsLogger.getLogger();

    private static final String TAG = "ai.factors.sdk.HTTPHelper";

    private static final MediaType JSON = MediaType.parse("application/json; charset=utf-8");
    private static final String AUTHORIZATION = "Authorization";

    // https://stackoverflow.com/questions/48532860/is-it-thread-safe-to-make-calls-to-okhttpclient-in-parallel
    private static OkHttpClient httpClient = new OkHttpClient();

    protected JSONObject executeRequest(String url, String projectToken, String payload) {
        RequestBody requestBody = RequestBody.create(JSON, payload);
        Request request = new Request.Builder()
                .url(url)
                .addHeader(AUTHORIZATION, projectToken)
                .post(requestBody)
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            String respStr = response.body().string();
            JSONObject jsonResp = new JSONObject(respStr);
            return jsonResp;
        }catch (java.net.ConnectException e ){
            logger.e(TAG, "Failed to execute Request ConnectException", e);
        }
        catch (IOException e) {
            logger.e(TAG, "Failed to execute Request IOException", e);
        } catch (JSONException e) {
            logger.e(TAG, "Failed to unmarshal response", e);
        }
        return null;
    }
}
