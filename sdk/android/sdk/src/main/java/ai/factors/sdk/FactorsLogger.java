package ai.factors.sdk;

import android.util.Log;

public class FactorsLogger {

    protected static final FactorsLogger instance = new FactorsLogger();
    private static final int defaultLogLevel = Log.INFO;
    private volatile boolean enabled = true;
    private volatile int logLevel = defaultLogLevel;

    private FactorsLogger() {
    }

    public static FactorsLogger getLogger() {
        return instance;
    }

    FactorsLogger setEnableLogging(boolean enable) {
        FactorsLogger factorsLogger = this;
        factorsLogger.enabled = enable;
        return factorsLogger;
    }

    FactorsLogger setLogLevel(int logLevel) {
        FactorsLogger factorsLogger = this;
        factorsLogger.logLevel = logLevel;
        return factorsLogger;
    }

    private boolean shouldLog(int logAtLevel) {
        return enabled && logLevel <= logAtLevel;
    }

    int d(String tag, String msg) {
        if (!shouldLog(Log.DEBUG)) {
            return 0;
        }
        return Log.d(tag, msg);
    }

    int d(String tag, String msg, Throwable tr) {
        if (!shouldLog(Log.DEBUG)) {
            return 0;
        }
        return Log.d(tag, msg, tr);
    }

    int v(String tag, String msg) {
        if (!shouldLog(Log.VERBOSE)) {
            return 0;
        }
        return Log.v(tag, msg);

    }

    int v(String tag, String msg, Throwable tr) {
        if (!shouldLog(Log.VERBOSE)) {
            return 0;
        }
        return Log.v(tag, msg, tr);
    }

    int i(String tag, String msg) {
        if (!shouldLog(Log.INFO)) {
            return 0;
        }
        return Log.i(tag, msg);
    }

    int i(String tag, String msg, Throwable tr) {
        if (!shouldLog(Log.INFO)) {
            return 0;
        }
        return Log.i(tag, msg, tr);
    }

    int w(String tag, String msg) {
        if (!shouldLog(Log.WARN)) {
            return 0;
        }
        ;
        return Log.w(tag, msg);
    }

    int w(String tag, Throwable tr) {
        if (!shouldLog(Log.WARN)) {
            return 0;
        }
        return Log.w(tag, tr);
    }

    int w(String tag, String msg, Throwable tr) {
        if (!shouldLog(Log.WARN)) {
            return 0;
        }
        return Log.w(tag, msg, tr);
    }

    int e(String tag, String msg) {
        if (!shouldLog(Log.ERROR)) {
            return 0;
        }
        return Log.e(tag, msg);
    }

    int e(String tag, String msg, Throwable tr) {
        if (!shouldLog(Log.ERROR)) {
            return 0;
        }
        return Log.e(tag, msg, tr);
    }

    // wtf = What a Terrible Failure, logged at level ASSERT
    int wtf(String tag, String msg) {
        if (!shouldLog(Log.ASSERT)) {
            return 0;
        }
        return Log.wtf(tag, msg);
    }

    int wtf(String tag, Throwable tr) {
        if (!shouldLog(Log.ASSERT)) {
            return 0;
        }
        return Log.wtf(tag, tr);
    }

    int wtf(String tag, String msg, Throwable tr) {
        if (!shouldLog(Log.ASSERT)) {
            return 0;
        }
        return Log.wtf(tag, msg, tr);
    }
}