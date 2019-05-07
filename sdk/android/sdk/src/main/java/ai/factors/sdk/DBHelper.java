package ai.factors.sdk;

import android.content.ContentValues;
import android.content.Context;
import android.database.Cursor;
import android.database.SQLException;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteException;
import android.database.sqlite.SQLiteOpenHelper;

class DBHelper extends SQLiteOpenHelper {

    private static final String DB_NAME = "factors_db";
    private static final int DB_VERSION = 1;

    private static final FactorsLogger logger = FactorsLogger.getLogger();
    private static final String TAG = "ai.factors.sdk.DBHelper";

    private static final String TABLE_REQUESTS = "REQUESTS";
    private static final String TABLE_STORE = "STORE";

    private static final String PROP_USER_UUID = "USER_UUID";

    private static final String COL_VALUE = "VALUE";
    private static final String COL_PROPERTY = "PROPERTY";

    private static final String COL_ID = "ID";
    private static final String COL_PAYLOAD = "PAYLOAD";
    private static final String COL_TYPE = "TYPE";

    DBHelper(Context context) {
        super(context, DB_NAME, null, DB_VERSION);
    }

    @Override
    public void onCreate(SQLiteDatabase db) {
        String createStore = String.format("CREATE TABLE %s (%s TEXT PRIMARY KEY NOT NULL, %s TEXT);", TABLE_STORE, COL_PROPERTY, COL_VALUE);
        db.execSQL(createStore);

        String createReqCache = String.format("CREATE TABLE %s (%s INTEGER PRIMARY KEY AUTOINCREMENT, %s TEXT, %s TEXT);", TABLE_REQUESTS, COL_ID, COL_PAYLOAD, COL_TYPE);
        db.execSQL(createReqCache);
    }

    @Override
    public void onUpgrade(SQLiteDatabase db, int oldVersion, int newVersion) {
        if (newVersion <= 1) {
            return;
        }

        if (oldVersion > newVersion) {
            logger.e(TAG, "onUpgrade() with invalid oldVersion and newVersion");
        }

        resetDatabase(db);
    }

    private void resetDatabase(SQLiteDatabase db) {
        db.execSQL(String.format("DROP TABLE IF EXISTS %s", TABLE_STORE));
        db.execSQL(String.format("DROP TABLE IF EXISTS %s", TABLE_REQUESTS));
        onCreate(db);
    }

    private void storeProperty(final String propertyName, final String value) {
        ContentValues values = new ContentValues();
        values.put(COL_PROPERTY, propertyName);
        values.put(COL_VALUE, value);
        SQLiteDatabase db = this.getWritableDatabase();
        try {
            db.insertWithOnConflict(TABLE_STORE, null, values, SQLiteDatabase.CONFLICT_REPLACE);
        } catch (SQLiteException e) {
            logger.e(TAG, String.format("Failed to insert property: %s, value: %s", propertyName, value), e);
        }
    }

    private String getProperty(final String propertyName) {

        String selectSQL = String.format("SELECT * FROM %s WHERE %s='%s';", TABLE_STORE, COL_PROPERTY, propertyName);
        SQLiteDatabase db = this.getReadableDatabase();
        try (
                Cursor c = db.rawQuery(selectSQL, null)
        ) {
            if (c.moveToFirst()) {
                return c.getString(c.getColumnIndex(COL_VALUE));
            }
        } catch (SQLiteException e) {
            logger.e(TAG, String.format("Failed to fetch property: %s", propertyName), e);
        }
        return null;
    }

    void storeUserUUID(final String userUUID) {
        if (Utils.isEmptyString(userUUID)) {
            return;
        }
        storeProperty(PROP_USER_UUID, userUUID);
    }

    String getUserUUID() {
        return getProperty(PROP_USER_UUID);
    }

    void storeRequest(final String type, final String payload) {
        ContentValues values = new ContentValues();
        values.put(COL_PAYLOAD, payload);
        values.put(COL_TYPE, type);
        SQLiteDatabase db = this.getWritableDatabase();
        try {
            db.insert(TABLE_REQUESTS, null, values);
        } catch (SQLiteException e) {
            logger.e(TAG, String.format("Failed to insert request Type: %s, Payload: %s", type, payload), e);
        }
    }

    void deleteRequest(final long id) {
        String deleteQuery = String.format("DELETE FROM %s WHERE %s = %d;", TABLE_REQUESTS, COL_ID, id);
        SQLiteDatabase db = this.getWritableDatabase();
        try {
            db.execSQL(deleteQuery);
        } catch (SQLiteException e) {
            logger.e(TAG, "Failed to delete request", e);
        }
    }

    RequestModel getNextRequest() {
        String selectSQL = String.format("SELECT * FROM %s ORDER BY %s ASC LIMIT 1;", TABLE_REQUESTS, COL_ID);
        RequestModel requestModel = null;

        SQLiteDatabase db = this.getReadableDatabase();
        try (Cursor c = db.rawQuery(selectSQL, null);) {
            if (c.moveToFirst()) {
                requestModel = new RequestModel(c.getLong(c.getColumnIndex(COL_ID)), c.getString(c.getColumnIndex(COL_PAYLOAD)), c.getString(c.getColumnIndex(COL_TYPE)));
            }
        } catch (SQLException e) {
            logger.e(TAG, "Failed to fetch request", e);
        }
        return requestModel;
    }

    long getRequestsCount() {
        String countQuery = String.format("SELECT COUNT(*) FROM %s;", TABLE_REQUESTS);
        SQLiteDatabase db = this.getReadableDatabase();
        long count = 0;
        try (Cursor cursor = db.rawQuery(countQuery, null)) {
            if (cursor != null && cursor.moveToFirst()) {
                count = cursor.getInt(0);
            }
        } catch (SQLException e) {
            logger.e(TAG, "Failed to fetch requestsCount", e);
        }
        return count;
    }
}
