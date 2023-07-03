<?php

header('Access-Control-Allow-Origin: *');
header('Content-Type: application/json');
header('Access-Control-Allow-Method: POST');
header('Access-Control-Allow-Headers:Content-Type');

include('device_detector.php');

$requestMethod = $_SERVER["REQUEST_METHOD"];

$userAgent = GetUserAgent();

if ($requestMethod == 'POST') {
    // collect value of input field
    if (empty($userAgent)) {
        echo "User Agent is empty";
    } else {
        $device = GetDevice($userAgent);
        echo $device;
    }
}
