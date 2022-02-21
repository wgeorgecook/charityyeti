var donationAmt;
var deviceData;

var fadeLoader = function(){
	document.getElementById('splash').style.opacity = '0'
	window.setTimeout(function(){
		document.getElementById('splash').style.display = 'none'
	}, 300)
}
var showLoader = function(){
	document.getElementById('splash').style.display = 'block'
	document.getElementById('splash').style.opacity = '0.9'
}
var tryAgain = function(){
	document.getElementById('donate-fail').style.inset = '50% 0'
	document.getElementById('donate-fail').style.opacity = '0'
}
var writeMsg = function(x){
	fadeLoader()
	document.getElementById('donate-fail').style.inset = '9vh 0 0'
	document.getElementById('donate-fail').style.opacity = '1'
	document.getElementById('error-msg').innerHTML = x
}

var preflight = function(){
	showLoader()
	var re = /[-+]?\d*\.?\d+([eE][-+]?\d+)?/gm
	donationAmt = (document.forms[0].amnt.value).match(re)
	donationAmt = Math.round(100 * donationAmt) / 100
	if(donationAmt > 0){
		if(donationAmt >= 5){
			return true
		} else {
			writeMsg("Donation amount must be at least $5.");
		}
	} else {
		writeMsg("You need to specify a donation amount.");
		return false;
	}
}
var doPaymentFail = function(err){
	writeMsg('Unable to make payment. Please ensure all payment information has been entered correctly.');
}
var doPaymentSuccess = function(nonce){
	var xhttp;
	xhttp = new XMLHttpRequest();
	xhttp.onreadystatechange = function() {
		if (this.readyState == 4 && this.status == 200) {
			fadeLoader()
			document.getElementById('donate-success').style.inset = '9vh 0 0'
			document.getElementById('donate-success').style.opacity = '1'
		}
		if (this.readyState == 4 && this.status == 403) {
			writeMsg(this.responseText);
		}
	};
	xhttp.open('POST', '/donate/success/', true);
	xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
	xhttp.send("nonce="+nonce+"&amt="+donationAmt+"&device="+deviceData);
}