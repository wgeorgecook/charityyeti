import React from 'react'
import bigLogo from './img/biglogo.png'
import Header from './header';
import tweetSeal from './img/tweet_seal.png'
import tweet1 from './img/tweet1.png'
import tweet2 from './img/tweet2.png'
import tweet3 from './img/tweet3.png'
import tweet4 from './img/tweet4.png'

const Home = () => {
    return (
        <div id='home'>
            <Header image='about'/>
            <div className={'mother mobile'}>
                <h1>What is CharityYeti?</h1>
                <p>Have you every read a tweet that warmed your &hearts; so much you wish there were a special way to say thanks?</p>
                <p>CharityYeti is a Twitter bot that allows you to reward worthy tweets with a donation to charity. Here's how it works:</p>
                <ul>
                <li><strong>Step 1:</strong> Reply to the tweet you want to reward and tag @CharityYeti</li>
                </ul>
                <section class="tweet">
                    <span class="tweeticon">
                        <img src={tweetSeal} alt='An illustration of a disappointed seal'/>
                    </span>
                    <span class="tweetbody">
                        <strong>You</strong> <em>@yourusername</em>
                        <br/><br/>
                        Replying to @hankgreen
                        <br/><br/>
                        Hey @CharityYeti I loved this so much, help me say thank you!
                        <br/><br/>
                        <img class="tweetaction" src={tweet1} alt='A purple reply icon'/>
                        <img class="tweetaction" src={tweet2} alt='A purple retweet icon'/>
                        <img class="tweetaction" src={tweet3} alt='A purple heart icon'/>
                        <img class="tweetaction" src={tweet4} alt='A purple share icon'/>
                    </span>
                </section>
                <ul>
                    <li>
                        <strong>Step 2:</strong> You will get a unique link to click to make your donation.
                    </li>
                </ul>
                <section class="tweet">
                    <span class="tweeticon">
                        <img src={bigLogo} alt='An illustration of a Yeti holding a bird, which has a coin in its mouth'/>
                    </span>
                    <span class="tweetbody">
                        <strong>Charity Yeti</strong> <em>@charityyeti</em><br/><br/>No problem. Here is your link - https://charityyeti.com/donate
                        <br/>
                        <br/>
                        <img class="tweetaction" src={tweet1} alt='A purple reply icon'/>
                        <img class="tweetaction" src={tweet2} alt='A purple retweet icon'/>
                        <img class="tweetaction" src={tweet3} alt='A purple heart icon'/>
                        <img class="tweetaction" src={tweet4} alt='A purple share icon'/>
                    </span>
                </section>
                <ul>
                    <li>
                        <strong>Step 3:</strong> CharityYeti will take care of the rest &#128522;
                    </li>
                </ul>
            </div>
        </div>
    )
}

export default Home