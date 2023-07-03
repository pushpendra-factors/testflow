<?php

$request = $_SERVER['REQUEST_URI'];
$viewDir = '/views/';
switch ($request) {
    case '':
    case '/device_service':
        require __DIR__ . $viewDir . 'home.php';
        break;
    default:
        http_response_code(404);
        require __DIR__ . $viewDir . '404.php';
}
