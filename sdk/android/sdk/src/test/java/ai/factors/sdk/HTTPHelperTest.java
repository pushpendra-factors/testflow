package ai.factors.sdk;

import org.junit.Assert;
import org.junit.Test;

public class HTTPHelperTest {
    @Test
    public void testTrack() {
        Assert.assertEquals("TRACK",SDKRoute.TRACK.toString());
        Assert.assertEquals(SDKRoute.TRACK.getRoute(), "/sdk/event/track");
    }
}
