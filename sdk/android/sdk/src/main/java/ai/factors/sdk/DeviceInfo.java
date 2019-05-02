package ai.factors.sdk;

import android.content.Context;
import android.os.Build;
import android.telephony.TelephonyManager;

class DeviceInfo {

    private Context context;

    DeviceInfo(Context context) {
        this.context = context;
    }

    static final String OS_NAME = "android";

    static String getOsName() {
        return OS_NAME;
    }

    static String getOsVersion() {
        return Build.VERSION.RELEASE;
    }

    static String getBrand() {
        return Build.BRAND;
    }

    static String getManufacturer() {
        return Build.MANUFACTURER;
    }

    static String getModel() {
        return Build.MODEL;
    }

    String getNetworkCarrier() {
        try {
            TelephonyManager manager = (TelephonyManager) context
                    .getSystemService(Context.TELEPHONY_SERVICE);
            return manager.getNetworkOperatorName();
        } catch (Exception e) {
            // Failed to get network operator name from network

        }
        return null;
    }

}
