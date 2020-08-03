<?php
session_start();
require_once $_SERVER['DOCUMENT_ROOT'] . '/braintree/lib/autoload.php';

$result = $gateway->transaction()->sale([
	'amount' => '25.00',
	'paymentMethodNonce' => 'tokencc_bc_d8zjt3_vstd5q_n8s2gc_5473d5_xp2',
	'deviceData' => {
		"correlation_id": "fee775907bba559715289b8992fecf47"
	},
	'options' => [
		'submitForSettlement' => true
	]
]);

print("<pre>");
print_r($result);
print("</pre>");

//let's send back an error message with these potential points of failure...
/*
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

$result = $gateway->transaction()->sale([
	'amount' => $amt,
	'paymentMethodNonce' => $nonce,
	'deviceData' => $device,
	'options' => [
		'submitForSettlement' => true
	]
]);

if ($result->success) {
	$post = new stdClass();
	$post->_id = $id;
	$post->donationValue = (int) $amt;
	$postval = json_encode($post);
	//$ch = curl_init();
	//curl_setopt($ch, CURLOPT_URL, "https://charityyeti.herokuapp.com/update");
	//curl_setopt($ch, CURLOPT_POSTFIELDS, $postval);
	//curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
	//curl_setopt($ch, CURLOPT_HTTPHEADER, array(                                                                          
	//	'Content-Type: application/json',                                                                                
	//	'Content-Length: ' . strlen($postval))                                                                       
	//);
	//$output = curl_exec($ch);
	//curl_close($ch);

	//$_SESSION['donationkey'] = '';

	//is there something specific we want to send back here to be included on the success page?
	print_r($postval);
} else {
	http_response_code(403);
	print('There was no amount specified.');
	exit;
}


?>