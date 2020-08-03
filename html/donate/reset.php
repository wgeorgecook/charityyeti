<pre>
<?php
/*
$post = new stdClass();
$post->_id = '5e330cf31398d78ea074d32c';
$post->donationValue = 0;
$postval = json_encode($post);
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, "https://charityyeti.herokuapp.com/update");
curl_setopt($ch, CURLOPT_POSTFIELDS, $postval);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
curl_setopt($ch, CURLOPT_HTTPHEADER, array(                                                                          
	'Content-Type: application/json',                                                                                
	'Content-Length: ' . strlen($postval))                                                                       
);
$output = curl_exec($ch);
curl_close($ch);
print_r($postval);
//*/
?>
</pre>

<pre>
<?php
//*
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, "https://charityyeti.herokuapp.com/get?id=5e330cf31398d78ea074d32c");
curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
$output = json_decode(curl_exec($ch));
curl_close($ch);
print_r($output);
//*/
?>
</pre>