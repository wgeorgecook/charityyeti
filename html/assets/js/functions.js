var doPaymentSuccess;
var doPaymentFail;
var preflight;
var donationAmt;
var deviceData;
var doClose;

jQuery(function ($){
	$(document).ready(function(){
		$(".js").show();
		$(".nojs").hide();
		
		preflight = function(){
			$(".blackout").show();
			$(".blackout-text").removeClass('success');
			$(".blackout-text").addClass('fail');
			donationAmt = document.forms[0].amount.value
			if(donationAmt == 0){
				var re = /[-+]?\d*\.?\d+([eE][-+]?\d+)?/gm
				donationAmt = (document.forms[0].amnt.value).match(re)
				donationAmt = Math.round(100 * donationAmt) / 100
			}
			if(donationAmt > 0){
				return true;
			} else {
				writeMsg("You need to specify a donation amount.");
				return false;
			}
			
		}
		
		doPaymentSuccess = function(nonce){
			$.ajax({
				url: "/donate/success/",
				method: 'POST',
				data: {
					nonce: nonce,
					amt: donationAmt,
					device: deviceData
				},
				success: function(data){
					$(".blackout-text").removeClass('fail');
					$(".blackout-text").addClass('success');
					writeMsg("$" + donationAmt + " was donated successfully.");
				},
				error: function(xhr){
					writeMsg('Something went wrong.');
				}
			});
		}
		
		doPaymentFail = function(err){
			writeMsg('Unable to make payment. Please ensure all payment information has been entered correctly.');
		}
		
		var writeMsg = function(x){
			$(".blackout-text").html('<div>' + x + '</div><button id="close-btn" onclick="doClose()">CLOSE</button>');
			$(".blackout-text").css('opacity', '1');
		}
		
		$(".blackout").click(function(){
			doClose();
		});
		
		doClose = function(){
			if($(".blackout-text").hasClass("success")){
				window.location.href = '/';
			} else {
				$(".blackout-text").html('');
				$(".blackout-text").css('opacity', '0.01');
				$(".blackout").hide();
			}
		}
	});
});