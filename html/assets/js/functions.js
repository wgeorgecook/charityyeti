var doPaymentSuccess;
var doPaymentFail;
var preflight;
var donationAmt;
var deviceData;

jQuery(function ($){
	$(document).ready(function(){
		$(".js").show();
		$(".nojs").hide();
		
		preflight = function(){
			console.log('doing the preflight check');
			$(".blackout").show();
			donationAmt = document.forms[0].amount.value
			//var method = $(this).attr('rel-type');
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
			console.log('doing the success stuff');
			$.ajax({
				url: "/donate/success/",
				method: 'POST',
				data: {
					nonce: nonce,
					amt: donationAmt,
					device: deviceData
				},
				success: function(data){
					writeMsg("$" + donationAmt + " was donated successfully.");
				},
				error: function(xhr){
					writeMsg('Something went wrong.');
				}
			});
		}
		
		doPaymentFail = function(err){
			writeMsg('payment failure');
			console.log(err);
		}
		
		var writeMsg = function(x){
			$(".blackout-text").html(x);
			$(".blackout-text").css('opacity', '1');
		}
		
		$(".blackout").click(function(){
			$(".blackout-text").html('');
			$(".blackout-text").css('opacity', '0.01');
			$(this).hide();
		});
	});
});