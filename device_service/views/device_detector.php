<?php

require_once __DIR__ . '/../vendor/autoload.php';

use DeviceDetector\DeviceDetector;
use DeviceDetector\Parser\Device\AbstractDeviceParser;

// function  GetDevice() returns device info
function GetDevice($userAgent)
{

    AbstractDeviceParser::setVersionTruncation(AbstractDeviceParser::VERSION_TRUNCATION_NONE);
    $dd = new DeviceDetector($userAgent);
    $dd->parse();

    // Check if user agent is a bot
    if ($dd->isBot()) {
        $botInfo = $dd->getBot();
        echo json_encode(["is_bot" => $botInfo]);
    } else {

        $osInfo = $dd->getOs();
        $device = $dd->getDeviceName();
        $brand = $dd->getBrandName();
        $model = $dd->getModel();

        return json_encode([
            "is_bot" => $dd->isBot(),
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
