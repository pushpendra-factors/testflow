package ai.factors.sdk;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

public class Utils {

    public static boolean isEmptyString(String input) {
        return input == null || input.length() == 0;
    }

    public static JSONObject mergeJSONObjects(JSONObject obj1, JSONObject obj2) {
        JSONObject merged = new JSONObject();

        copyJSONObjects(merged, obj1);
        copyJSONObjects(merged, obj2);

        return merged;
    }

    // Do a shallow copy
    public static void copyJSONObjects(JSONObject dest, JSONObject src) {

        if (dest == null || src == null || src.length() == 0) {
            return;
        }

        JSONArray nameArray = src.names();
        String[] names = new String[nameArray.length()];
        for (int i = 0; i < nameArray.length(); i++) {
            names[i] = nameArray.optString(i);
        }

        for (int i = 0 ; i < names.length; i++){
            try {
                dest.put(names[i], src.get(names[i]));
            } catch (JSONException e) {

            }
        }
    }
}
