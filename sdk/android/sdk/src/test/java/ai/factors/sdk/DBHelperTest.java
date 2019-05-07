package ai.factors.sdk;

import android.content.Context;

import org.junit.After;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.robolectric.RobolectricTestRunner;

import androidx.test.core.app.ApplicationProvider;

@RunWith(RobolectricTestRunner.class)
public class DBHelperTest {

    Context ctx;
    DBHelper db;

    @Before
    public void setup() {
        ctx = ApplicationProvider.getApplicationContext();
        db = new DBHelper(ctx);
    }

    @After
    public void tearDown() {
        db.close();
    }

    @Test
    public void testCreateDBHelper() {
        Assert.assertNotNull(db);
    }

    @Test
    public void testSetGetUserUUID() {

        // Case 1:
        String userUUID = "testUser13";
        db.storeUserUUID(userUUID);
        String resultUserUUID = db.getUserUUID();
        Assert.assertEquals(resultUserUUID, userUUID);


        // Case 2:
        String updatedUserUUID = "testUser14";
        db.storeUserUUID(updatedUserUUID);
        String resultUpdatedUserUUID = db.getUserUUID();
        Assert.assertEquals(resultUpdatedUserUUID, updatedUserUUID);
    }

    @Test
    public void testSetGetDeleteRequest() {
        String payload = "{\"event_name\":\"test_event_created\"}";
        String type = "event";
        db.storeRequest(type, payload);

        RequestModel req = db.getNextRequest();
        Assert.assertEquals(payload, req.payload);
        Assert.assertEquals(type, req.type);
        Assert.assertTrue(req.ID > 0);

        db.deleteRequest(req.ID);
        req = db.getNextRequest();
        Assert.assertNull(req);
    }

    @Test
    public void testGetNextRequest() {
        String typeIdentify = "identify";
        String typeEvent = "event";

        String payload1 = "{\"event_name\":\"test_event_created\"}";
        db.storeRequest(typeEvent, payload1);

        String payload2 = "{\"c_uid\":\"cust_user_id\"}";
        db.storeRequest(typeIdentify, payload2);

        String payload3 = "{\"event_name\":\"test_event_created\"}";
        db.storeRequest(typeEvent, payload3);

        RequestModel req = db.getNextRequest();
        Assert.assertEquals(payload1, req.payload);

        db.deleteRequest(req.ID);

        req = db.getNextRequest();
        Assert.assertEquals(payload2, req.payload);
        db.deleteRequest(req.ID);

        req = db.getNextRequest();
        Assert.assertEquals(payload3, req.payload);
        db.deleteRequest(req.ID);
    }

    @Test
    public void testCountRequests() {
        String typeIdentify = "identify";
        String typeEvent = "event";

        String payload1 = "{\"event_name\":\"test_event_created\"}";
        db.storeRequest(typeEvent, payload1);

        long count = db.getRequestsCount();
        Assert.assertEquals(1, count);

        String payload2 = "{\"event_name\":\"test_event_created\"}";
        db.storeRequest(typeEvent, payload2);

        count = db.getRequestsCount();
        Assert.assertEquals(2, count);
    }
}