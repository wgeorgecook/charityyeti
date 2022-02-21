<?php
if(isset($_GET['id']) && ctype_xdigit($_GET['id'])){
	header('Location: /donate/' . $_GET['id']);
	exit;
}
?><!DOCTYPE html>
<html>

<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link rel="icon" type="image/png" sizes="192x192" href="assets/imgs/favicon.png">
	<link rel="icon" type="image/png" sizes="128x128" href="assets/imgs/favicon.png">
	<link rel="apple-touch-icon" type="image/png" sizes="128x128" href="assets/imgs/favicon.png">
	<link rel="apple-touch-icon-precomposed" type="image/png" sizes="128x128" href="assets/imgs/favicon.png">
	<title>Charity Yeti &ndash; a Twitter Bot for Rewarding Tweets</title>
	<link href="https://fonts.googleapis.com/css?family=Roboto:500&display=swap" rel="stylesheet">
	<link href="assets/flickity/flickity.css" rel="stylesheet">
	<link href="assets/flickity/flickity-fade.css" rel="stylesheet">
	<link href="assets/css/welcome.css" rel="stylesheet">
	<link href="assets/css/loading.css" rel="stylesheet">
</head>

<body>
	<section id="splash"><div id="biglogo"></div><div id="loading"><span class="ball ball1"></span><span class="ball ball2"></span><span class="ball ball3"></span></div></section>
	<div id="logo"></div>
	<section id="mothermold">
		<div class="modal welcome1">
			<div class="text"><p>Welcome to CharityYeti!<br/><br/>With my help, you'll be able to reward your favorite charities' tweets with donations.</p></div>
		</div>
		<div class="modal welcome2">
			<div class="text"><ol><li>Reply to the tweet you want to reward and tag @CharityYeti.</li><li>You will get a unique link to click on to make your donation.</li></ol></div>
		</div>
		<div class="modal welcome3">
			<div class="text"><p>That's it! (:<br/>We'll take care of the rest.</p></div>
		</div>
	</section>
</body>

<script src="assets/flickity/flickity.pkgd.min.js"></script>
<script src="assets/flickity/flickity-fade.js"></script>
<script>

var ratio =  Math.round((document.getElementById("mothermold").offsetHeight / document.getElementById("mothermold").offsetWidth) * 100) / 100;
if(ratio > 1.55){
	document.getElementById('mothermold').classList.add('narrow')
}

var loadH = Math.round(((0.95 -  (document.getElementById("biglogo").offsetHeight / window.innerHeight)) / 2) * 1000) / 10
document.getElementById("biglogo").style.marginTop = loadH + 'vh'

window.setTimeout(function(){
	document.getElementById('splash').style.opacity = '0'
	window.setTimeout(function(){
		document.getElementById('splash').style.display = 'none'
	}, 300)
}, 2000)
var flkty = new Flickity( '#mothermold', {
	cellAlign: 'left',
	contain: true,
	fade: true
})

if(document.querySelector(".flickity-prev-next-button.previous").getBoundingClientRect().left < 100){
	document.querySelector(".flickity-prev-next-button.previous").style.display = 'none'
	document.querySelector(".flickity-prev-next-button.next").style.display = 'none'
}

</script>

</html>