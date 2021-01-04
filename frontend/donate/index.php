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
		$idmsg = "The incoming ID is not formatted correctly";
	} else {
		$ch = curl_init();
		curl_setopt($ch, CURLOPT_URL, DB_BASE . "/get?id=$id");
		curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
		$output = json_decode(curl_exec($ch));
		curl_close($ch);
		if(empty($output)){
			$idmsg = "We were unable to find the tweet correlating to this ID";
		} else {
			if($output->donationValue){
				$idmsg = "This ID has already been donated towards.";
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
		$idmsg = "There was no ID supplied.";
	}
}
?>

<?php if(!$canDonate): ?>

<?php
//jrp: not sure exactly what we want to do if someone is here in error
http_response_code(403);
?>
<!DOCTYPE html>
<html><head><title>403 &ndash; Forbidden</title></head>
<body>Sorry, we were unable to process your request for the following reason:<pre><?php print($idmsg) ?></pre></body>
</html>

<?php else: ?>

<?php $pagetitle = "Charity Yeti &ndash; Donation Page" ?>
<?php $headerimg = "donate" ?>
<?php include('../inc/header.inc.php') ?>

<div class="nojs" style="min-height:250px">
Please note, JavaScript is required to process a donation. Please enable JavaScript and reload this page to continue with your donation.
</div>

<div class="js">

	<form action="/" id="payment-form" method="post">
		<div><h3>Donation Info</h3>
		<input id="radio-5" type="radio" name="amount" value="5" checked="checked" /><label for="radio-5">$5</label>
		<input id="radio-10" type="radio" name="amount" value="10" /><label for="radio-10">$10</label>
		<input id="radio-25" type="radio" name="amount" value="25" /><label for="radio-25">$25</label>
		<span class="wrapper">
			<input id="radio-other" type="radio" name="amount" value="0" onchange="$('input[name=amnt]').prop('disabled', false);$('input[name=amnt]').focus()" /><label for="radio-other">Other <em>(specify)</em>:<label>
			$<input name="amnt" type="number" disabled="true" />
		</span>
		<div style="margin-top: 10px">
			<strong>Charity:</strong> Partners in Health <em class="wrapper">*more options coming soon</em>
		</div>
		</div>
		
		<div>

		<h3>Credit Card</h3>
		
		<label for="card-number">Card Number</label>
		<div id="card-number" class="hosted-field"></div>
		
		<label for="cvv">CVV</label>
		<div id="cvv" class="hosted-field"></div>

		<label for="expiration-date">Expiration (MM / YY)</label>
		<div id="expiration-date" class="hosted-field"></div>
		
		<?php //jrp: do we need to collect zip code or name on card? ?>

		<button type="submit" id="credit-card-btn" class="doPayment" rel-type="credit" disabled="disabled">DONATE</button>
		</div>
	</form>
</div>
	
	<script src="https://js.braintreegateway.com/web/3.63.0/js/client.min.js"></script>
	<script src="https://js.braintreegateway.com/web/3.63.0/js/hosted-fields.min.js"></script>
	<script src="https://js.braintreegateway.com/web/3.63.0/js/data-collector.min.js"></script>
	<script src="https://js.braintreegateway.com/web/3.63.0/js/google-payment.min.js"></script>
    <script>
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
              'font-size': '16px'
            },
            'input.invalid': {
              'color': 'red'
            },
            'input.valid': {
              'color': 'green'
            },
			'input:focus': {
				'background-color': '#ffffff'
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

<?php include('../inc/footer.inc.php') ?>

<?php endif ?>
