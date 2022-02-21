<?php
session_start();
require_once $_SERVER['DOCUMENT_ROOT'] . '/settings.php';

if('sandbox' == BRAINTREE_ENV){
	ini_set('display_errors', 1);
	ini_set('display_startup_errors', 1);
	error_reporting(E_ALL);
}

$idmsg = '';
$canDonate = false;
$uri = explode('/',$_SERVER['REQUEST_URI']);
if((count($uri) > 2) && !empty($uri[2])){
	$id = $uri[2];
	if(!ctype_xdigit($id)){
		$idmsg = "The donation link appears to be invalid.";
	} else {
		$ch = curl_init();
		curl_setopt($ch, CURLOPT_URL, DB_BASE . "/get?id=$id");
		curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
		$output = json_decode(curl_exec($ch));
		curl_close($ch);
		if(empty($output)){
			$idmsg = "There doesn't appear to be a tweet this donation is for.";
		} else {
			if(!empty($output->donationValue)){
				// TODO: maybe we should add the previous donation value to this one
				// so we can return how much this person collectively donated?
				$idmsg = "We've already recieved a donation using this link.";
				if('sandbox' == BRAINTREE_ENV){
					$_SESSION['donationkey'] = $id;
					$canDonate = true; 
				}
			} else {
				$canDonate = true;
				$_SESSION['donationkey'] = $id;
			}
		}
	}
} else {
	if('sandbox' == BRAINTREE_ENV){
		$id = '5e330cf31398d78ea074d32c';
		$_SESSION['donationkey'] = $id;
		$canDonate = true;
	} else {
		$idmsg = "This is not a valid donation link.";
	}
}

//$canDonate = false;
//$idmsg = "I'm forcing an error page on you.";
?>

<?php if(!$canDonate): ?>

<?php
//jrp: not sure exactly what we want to do if someone is here in error
http_response_code(403);
?>
<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link rel="icon" type="image/png" sizes="192x192" href="/assets/imgs/favicon.png">
	<link rel="icon" type="image/png" sizes="128x128" href="/assets/imgs/favicon.png">
	<link rel="apple-touch-icon" type="image/png" sizes="128x128" href="/assets/imgs/favicon.png">
	<link rel="apple-touch-icon-precomposed" type="image/png" sizes="128x128" href="/assets/imgs/favicon.png">
	<title>403 &ndash; Forbidden</title>
	<link href="https://fonts.googleapis.com/css?family=Roboto:500&display=swap" rel="stylesheet">
	<link href="/assets/css/donate.css" rel="stylesheet">
	<link href="/assets/css/loading.css" rel="stylesheet">
</head>
<body>
<div id="logo"></div>
<section class="popup" id="donate-fail">
	<div class="modal">
		<div class="text">
			<p>We're encountered some difficulties:<br/><span class="goldtext"><?php print($idmsg) ?></span><br/><br/><button id="fail-btn" type="button" onclick="window.location = '../'">BACK</button></p>
		</div>
	</div>
</section>
</body>
<script>
document.getElementById('donate-fail').style.inset = '9vh 0 0'
document.getElementById('donate-fail').style.opacity = '1'
var ratio =  Math.round((document.getElementById("donate-fail").offsetHeight / document.getElementById("donate-fail").offsetWidth) * 100) / 100;
if(ratio > 1.55){
	document.getElementById("donate-fail").style.backgroundColor = '#ffffff'
}
</script>
</html>

<?php else: ?>

<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link rel="icon" type="image/png" sizes="192x192" href="/assets/imgs/favicon.png">
	<link rel="icon" type="image/png" sizes="128x128" href="/assets/imgs/favicon.png">
	<link rel="apple-touch-icon" type="image/png" sizes="128x128" href="/assets/imgs/favicon.png">
	<link rel="apple-touch-icon-precomposed" type="image/png" sizes="128x128" href="/assets/imgs/favicon.png">
	<title>Charity Yeti &ndash; a Twitter Bot for Rewarding Tweet</title>
	<link href="https://fonts.googleapis.com/css?family=Roboto:500,600&display=swap" rel="stylesheet">
	<link href="/assets/css/donate.css" rel="stylesheet">
	<link href="/assets/css/loading.css" rel="stylesheet">
</head>
<body>

<section id="splash"><div id="biglogo"></div><div id="loading"><span class="ball ball1"></span><span class="ball ball2"></span><span class="ball ball3"></span></div></section>

<div id="logo"></div>

<section id="nojs">
	<section class="popup" id="js-fail" style="inset: 9vh 0 0; opacity: 1">
		<div class="modal">
			<div class="text">
				<p>JavaScript is required to process a donation.<br/><span class="goldtext">Please enable JavaScript and reload this page to continue</span></p>
			</div>
		</div>
	</section>
</section>

<section id="js">

	<!--<h1>DONATE</h1>-->
	<!--<h2>Details:</h2-->

	<div class="modal">
		<div class="text">
			<form action="/" id="payment-form" method="post">
				<div class="form-half lefthalf"><label id="testdata">Amount:</label><input name="amnt" /></div><div class="form-half righthalf"><label>Supporting:</label><input disabled="disabled" value="@PIH" title="Currently only accepting donations supporting Partners in Heath" /></div>
				<div class="form-whole"><label>Card number:</label><div id="card-number" class="hosted-field"></div></div>
				<div class="form-half lefthalf"><label>Expires:</label><div id="expiration-date" class="hosted-field"></div></div><div class="form-half righthalf"><label>CVV:</label><div id="cvv" class="hosted-field"></div></div>
				<button type="submit" id="credit-card-btn" class="doPayment" rel-type="credit" disabled="disabled">DONATE</button>
			</form>
		</div>
	</div>

	<!-- FAILURE -->
	<section class="popup" id="donate-fail">
		<div class="modal">
			<div class="text">
				<p><span id="error-msg">Error</span><br/><span class="goldtext">Seal of disapproval</span><br/><br/><button id="fail-btn" type="button" onclick="tryAgain()">TRY AGAIN</button></p>
			</div>
		</div>
	</section>
	<!-- -->

	<!-- SUCCESS -->
	<section class="popup" id="donate-success">
		<div class="modal">
			<div class="text">
				<p>Success!<br/><span class="goldtext">Seal of approval</span><br/><br/><button id="success-btn" type="button" onclick="window.location = '../'">DONE</button></p>
			</div>
		</div>
	</section>
	<!-- -->

</section>

</body>
	
<script src="https://js.braintreegateway.com/web/3.63.0/js/client.min.js"></script>
<script src="https://js.braintreegateway.com/web/3.63.0/js/hosted-fields.min.js"></script>
<script src="https://js.braintreegateway.com/web/3.63.0/js/data-collector.min.js"></script>
<script src="https://js.braintreegateway.com/web/3.63.0/js/google-payment.min.js"></script>
<script src="../assets/js/donate.js"></script>

<script>

window.setTimeout(fadeLoader, 1800)
document.getElementById("js").style.display = 'block'
document.getElementById("nojs").style.display = 'none'

var loadH = Math.round(((0.95 -  (document.getElementById("biglogo").offsetHeight / window.innerHeight)) / 2) * 1000) / 10
document.getElementById("biglogo").style.marginTop = loadH + 'vh'

//Braintree
  var form = document.querySelector('#payment-form');
  var submit = document.querySelector('#credit-card-btn');

  braintree.client.create({
	authorization: '<?php print(BRAINTREE_ATH) ?>'
  }, function (clientErr, clientInstance) {
	if (clientErr) {
	  console.error(clientErr);
	  return;
	}
	braintree.dataCollector.create({
		client: clientInstance,
		paypal: false
	  }, function (err, dataCollectorInstance) {
		if (err) {
		  // Handle error in creation of data collector
		  console.log(err);
		  return;
		}
		// At this point, you should access the dataCollectorInstance.deviceData value and provide it
		// to your server, e.g. by injecting it into your form as a hidden input.
		deviceData = dataCollectorInstance.deviceData;
		console.log(deviceData);
	});

	// This example shows Hosted Fields, but you can also use this
	// client instance to create additional components here, such as
	// PayPal or Data Collector.

	braintree.hostedFields.create({
	  client: clientInstance,
	  styles: {
		'input': {
		  'color': '#fcb73d',
		  'font-weight': '500',
		  'font-size': '16px'
		},
		'input.invalid': {
		  'color': '#f95900'
		},
		'input.valid': {
		  'color': '#fcb73d'
		},
		'input:focus': {
			'background-color': '#f7f7ff'
		}
	  },
	  fields: {
		number: {
		  selector: '#card-number'
		},
		cvv: {
		  selector: '#cvv'
		},
		expirationDate: {
		  selector: '#expiration-date'
		}
	  }
	}, function (hostedFieldsErr, hostedFieldsInstance) {
	  if (hostedFieldsErr) {
		console.error(hostedFieldsErr);
		return;
	  }

	  submit.removeAttribute('disabled');

	  form.addEventListener('submit', function (event) {
		event.preventDefault();
		if(preflight()){
			hostedFieldsInstance.tokenize(function (tokenizeErr, payload) {
			  if (tokenizeErr) {
				doPaymentFail(tokenizeErr);
				return;
			  }

			  // If this was a real integration, this is where you would
			  // send the nonce to your server.
			  doPaymentSuccess(payload.nonce, "<?php echo $id ?>", ""); // TODO: send a client token if we have one
			});	
		}
	  }, false);
	});
  });
</script>

</html>
<?php endif ?>