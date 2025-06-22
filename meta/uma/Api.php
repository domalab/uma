<?php

function uma_log($m) {
	global $plugin;
  shell_exec("/usr/bin/logger"." ".escapeshellarg($m)." -t uma");
}

$socket_path = "/var/run/uma-api.sock";
$action = $_POST['action'] ?? '';
$params = !empty($_POST['params']) ? json_decode($_POST['params'], true) : new stdClass();

$socket = fsockopen("unix://{$socket_path}", -1, $errno, $errstr, 30);
if (!$socket) {
  http_response_code(503);
  $reply = [
    'data' => null,
    'error' => $errstr,
  ];
  uma_log("service unavailable ${errstr}");
  echo json_encode($reply);
  exit;
}

// send data to the golang server
$data = [
    'action' => $action,
    'params' => $params,
];
$json_data = json_encode($data) . "\n";

fwrite($socket, $json_data);

// read the response from the golang server
$response = '';
while (!feof($socket)) {
    $line = fgets($socket, 1024);
    $response .= trim($line);
}

// Close the socket
fclose($socket);

// Send the response back to the mobile app (if needed)
echo json_encode($response);

// Set appropriate HTTP headers and status code
http_response_code(200);
header('Content-Type: application/json');
