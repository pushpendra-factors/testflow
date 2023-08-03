<?php

require_once __DIR__ . '/../vendor/autoload.php';

use DeviceDetector\ClientHints;
use DeviceDetector\DeviceDetector;
use DeviceDetector\Parser\Device\AbstractDeviceParser;

function LogExecutionTime($time_start, $userAgent)
{
    $time_end = microtime(true);
    $execution_time = ($time_end - $time_start) * 1000;
    error_log('User Agent: ' . $userAgent . ' Total Execution Time: ' . $execution_time . ' ms .');
}

// function  GetDevice() returns device info
function GetDevice($userAgent)
{

    $time_start = microtime(true);
    $clientHints = ClientHints::factory($_SERVER);
    AbstractDeviceParser::setVersionTruncation(AbstractDeviceParser::VERSION_TRUNCATION_NONE);
    $dd = new DeviceDetector($userAgent, $clientHints);

    // caching device info
    $cache = new \Symfony\Component\Cache\Adapter\ApcuAdapter('', 31622400, null, null);
    $dd->setCache(
        new \DeviceDetector\Cache\PSR6Bridge($cache)
    );

    // parse user agent
    $dd->parse();

    // Check for bot
    if ($dd->isBot()) {
        LogExecutionTime($time_start, $userAgent);

        return json_encode(["is_bot" => $dd->isBot()]);
    } else {

        $clientInfo = $dd->getClient();
        $osInfo = $dd->getOs();
        $device = $dd->getDeviceName();
        $brand = $dd->getBrandName();
        $model = $dd->getModel();

        LogExecutionTime($time_start, $userAgent);

        return json_encode([
            "is_bot" => $dd->isBot(),
            "client_info" => $clientInfo,
            "os_info" => $osInfo,
            "device_type" => $device,
            "device_brand" => $brand,
            "device_model" => $model,
        ], JSON_PRETTY_PRINT);
    }
}

// function  getUserAgent returns User-Agent Request Header value
function GetUserAgent()
{
    foreach ($_SERVER as $key => $value) {
        if (substr($key, 0, 5) <> 'HTTP_') {
            continue;
        }
        $header = str_replace(' ', '-', ucwords(str_replace('_', ' ', strtolower(substr($key, 5)))));
        if ($header == 'User-Agent') {
            return $value;
        }
    }
    return array();
}
