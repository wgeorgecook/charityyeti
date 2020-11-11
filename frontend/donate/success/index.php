<?php
session_start();
require_once $_SERVER['DOCUMENT_ROOT'] . '/settings.php';

if('sandbox' == BRAINTREE_ENV){
	ini_set('display_errors', 1);
	ini_set('display_startup_errors', 1);
	error_reporting(E_ALL);
}

require_once $_SERVER['DOCUMENT_ROOT'] . '/braintree/lib/Braintree.php';

if($_SERVER['REQUEST_METHOD'] != 'POST'){
	http_response_code(403);
	print('Request method was not POST.');
	exit;
}
if(!isset($_POST['nonce']) || empty($_POST['nonce'])){
	http_response_code(403);
	print('There was no payment nonce.');
	exit;
}
if(!isset($_POST['amt']) || empty($_POST['amt'])){
	http_response_code(403);
	print('There was no amount specified.');
	exit;
}
if(!isset($_POST['device']) || empty($_POST['device'])){
	http_response_code(403);
	print('Unable to determine device.');
	exit;
}
if(!isset($_SESSION['donationkey']) || empty($_SESSION['donationkey'])){
	http_response_code(403);
	print_r($_SESSION);
	exit;
}
$id = $_SESSION['donationkey'];
$nonce = $_POST['nonce'];
$device = $_POST['device'];
$amt = $_POST['amt'];

$gateway = new Braintree\Gateway([
	'environment' => BRAINTREE_ENV,
	'merchantId' => BRAINTREE_MID,
	'publicKey' => BRAINTREE_PUB,
	'privateKey' => BRAINTREE_PVT
]);
$transaction = $gateway->transaction();

// TODO: need to include token if it exists
$params = [
	'payment_amount' => $amt,
	'payment_method_nonce' => $nonce,
	'device_data' => $device,
	'options' => [
		'submitForSettlement' => true
	],
	'_id' => $id,
	'token' => '',
];

$result = $transaction->sale($params);


// TODO: I now handle updating the database directly from 
// the backend on a successful donation. Is there 
// anything else you need me to return to you? If not 
// I don't think we need this. I'll just return a 200 OK.
if ($result->success) {
	$post = new stdClass();
	$post->_id = $id;
	$post->donationValue = (int) $amt;
	$postval = json_encode($post);
	$ch = curl_init();
	curl_setopt($ch, CURLOPT_URL, DB_BASE . "/update");
	curl_setopt($ch, CURLOPT_POSTFIELDS, $postval);
	curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
	curl_setopt($ch, CURLOPT_HTTPHEADER, array(                                                                          
		'Content-Type: application/json',                                                                                
		'Content-Length: ' . strlen($postval))                                                                       
	);
	$output = curl_exec($ch);
	curl_close($ch);

	if('production' == BRAINTREE_ENV){
		$_SESSION['donationkey'] = '';
		//jrp: is there something specific we want to send back here to be included on the success page?
	} else {
		print_r($postval);
	}
} else {
	http_response_code(403);
	print('There was a transaction error.');
	exit;
}


?>