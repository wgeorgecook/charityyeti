<?php
if(isset($_GET['id']) && ctype_xdigit($_GET['id'])){
	header('Location: /donate/' . $_GET['id']);
}
?>

<?php $pagetitle = "Charity Yeti &ndash; a Twitter Bot for Rewarding Tweets" ?>
<?php $headerimg = "about" ?>
<?php include('inc/header.inc.php') ?>

<h1>What is CharityYeti?</h1>
<p>Have you every read a tweet that warmed your &hearts; so much you wish there were a special way to say thanks?</p>
<p>CharityYeti is a Twitter bot that allows you to reward worthy tweets with a donation to charity. Here's how it works:</p>
<ul>
<li><strong>Step 1:</strong> Reply to the tweet you want to reward and tag @CharityYeti</li>
</ul>
<section class="tweet">
<span class="tweeticon"><img src="/assets/imgs/tweet_seal.png" /></span>
<span class="tweetbody"><strong>You</strong> <em>@yourusername</em><br/><br/>Replying to @hankgreen<br/><br/>Hey @CharityYeti I loved this so much, help me say thank you!<br/><br/><img class="tweetaction" src="/assets/imgs/tweet1.png" /><img class="tweetaction" src="/assets/imgs/tweet2.png" /><img class="tweetaction" src="/assets/imgs/tweet3.png" /><img class="tweetaction" src="/assets/imgs/tweet4.png" /></span>
</section>
<ul>
<li><strong>Step 2:</strong> You will get a unique link to click to make your donation.</li>
</ul>
<section class="tweet">
<span class="tweeticon"><img src="/assets/imgs/biglogo.png" /></span>
<span class="tweetbody"><strong>Charity Yeti</strong> <em>@charityyeti</em><br/><br/>No problem. Here is your link - https://charityyeti.com/donate<br/><br/><img class="tweetaction" src="/assets/imgs/tweet1.png" /><img class="tweetaction" src="/assets/imgs/tweet2.png" /><img class="tweetaction" src="/assets/imgs/tweet3.png" /><img class="tweetaction" src="/assets/imgs/tweet4.png" /></span>
</section>
<ul>
<li><strong>Step 3:</strong> CharityYeti will take care of the rest &#128522;</li>
</ul>
<p></p>

<?php include('inc/footer.inc.php') ?>