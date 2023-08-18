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

    // parse user agent
    $dd->parse();

    // Check for bot
    $isBot = $dd->isBot();
    if ($isBot) {
        LogExecutionTime($time_start, $userAgent);

        return json_encode(["is_bot" => $isBot]);
    } else {

        $clientInfo = $dd->getClient();

        $osInfo = $dd->getOs();
        if (count($osInfo) == 0) {
            $osInfo = null;
        }

        $device = $dd->getDeviceName();
        $brand = $dd->getBrandName();
        $model = $dd->getModel();

        // check for invalid user agent
        if (is_null($clientInfo) && $device == "" && $brand == "" && is_null($osInfo)) {
            http_response_code(404); // Set HTTP response code to 404 (Not found)
            echo "User Agent not found";
            exit;
        }

        LogExecutionTime($time_start, $userAgent);

        return json_encode([
            "is_bot" => $isBot,
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
