<?php
session_start();
$idmsg = '';
$canDonate = false;
$uri = explode('/',$_SERVER['REQUEST_URI']);
if(count($uri > 2) && !empty($uri[2])){
	$id = $uri[2];
	if(!ctype_xdigit($id)){
		$idmsg = "The incoming ID is not formatted correctly";
	} else {
		$ch = curl_init();
		curl_setopt($ch, CURLOPT_URL, "https://charityyeti.herokuapp.com/get?id=$id");
		curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
		$output = json_decode(curl_exec($ch));
		curl_close($ch);
		if(empty($output)){
			$idmsg = "The provided ID appears to be incorrect";
		} else {
			if($output->donationValue){
				$idmsg = "This ID has already been donated towards.";
				//remove below once I've figured out how to add new entries
				$_SESSION['donationkey'] = $id;
				$canDonate = true; 
			} else {
				$canDonate = true;
				$_SESSION['donationkey'] = $id;
				$idmsg = "You are donating for ID# $id";
			}
		}
	}
} else {
	$idmsg = "No ID supplied";
}
?>

<?php if(!$canDonate): ?>

<?php
//not sure exactly what we want to do if someone is here in error
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
<?php //print($idmsg) ?>

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
		<!--
		<h2>Method</h2>
		<button type="button" id="paypal-btn" class="doPayment" rel-type="paypal">PayPal</button><br/>
		<button type="button" id="venmo-btn" class="doPayment" rel-type="venmo">Venmo</button><br/>
		<button type="button" id="google-pay-btn" class="doPayment" rel-type="google">Google Pay</button><br/>
		<button type="button" id="apple-pay-btn" class="doPayment" rel-type="apple">Apply Pay</button><br/>
		<button type="button" id="samsung-pay" class="doPayment" rel-type="samsung">Samsung Pay</button><br/>
		-->
		<h3>Credit Card</h3>
		
		<label for="card-number">Card Number</label>
		<div id="card-number" class="hosted-field"></div>
		
		<label for="cvv">CVV</label>
		<div id="cvv" class="hosted-field"></div>

		<label for="expiration-date">Expiration (MM / YY)</label>
		<div id="expiration-date" class="hosted-field"></div>
		
		<!-- do we even need these?
		<label for="card-name">Name as it Appears on Card</label>
		<div id="card-name"><input id="card-name" type="text" /></div>
		
		<label for="card-zip">Billing ZIP Code</label>
		<div id="card-zip"><input id="card-zip" type="text" /></div>
		-->

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
        authorization: 'sandbox_ykcw354q_wmv3nh8skrjwtqr5'
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
					  console.log('epic fail');
					doPaymentFail(tokenizeErr);
					return;
				  }

				  // If this was a real integration, this is where you would
				  // send the nonce to your server.
				  console.log('epic win');
				  doPaymentSuccess(payload.nonce);
				});	
			}
          }, false);
        });
      });
    </script>

<?php include('../inc/footer.inc.php') ?>

<?php endif ?>