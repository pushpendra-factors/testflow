package ai.factors.sdk;

import android.content.Context;

import org.json.JSONObject;

public interface Factors {

    Factors init(final Context context, final String projectToken);

    void addUserProperties(final JSONObject properties);

    void identify(final String customerUserID);

    void track(final String eventName, final JSONObject eventProperties);

    Factors enableLogging(final boolean enableLogging);

    Factors setLogLevel(final int logLevel);
}