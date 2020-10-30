import React from 'react'
import Header from './header'

const donate = () => {
    return (
        <div>
            <Header image='donate' />
                <form action="/" id="payment-form" className='mother' method="post">
                    <div>
                        <h3>Donation Info</h3>
                        <input id="radio-5" type="radio" name="amount" value="5" />
                        <label for="radio-5">$5</label>
                        <input id="radio-10" type="radio" name="amount" value="10" />
                        <label for="radio-10">$10</label>
                        <input id="radio-25" type="radio" name="amount" value="25" />
                        <label for="radio-25">$25</label>
                        <span class="wrapper">
                            <input id="radio-other" type="radio" name="amount" value="0" />
                                <label for="radio-other">Other <em>(specify)</em>:</label>
                            $<input name="amnt" type="number" />
                        </span>
                        <div style={{marginTop: "10px"}}>
                            <strong>Supporting:</strong> Partners in Health
                        </div>
            
                        <div>
                            <h3>Credit Card</h3>
                    
                            <label for="card-number">Card Number</label>
                            <div id="card-number" class="hosted-field"></div>
                        
                            <label for="cvv">CVV</label>
                            <div id="cvv" class="hosted-field"></div>

                            <label for="expiration-date">Expiration (MM / YY)</label>
                            <div id="expiration-date" class="hosted-field"></div>
                        
                            <label for="card-name">Name as it Appears on Card</label>
                            <div id="card-name"><input id="card-name" type="text" /></div>
                            
                            <label for="card-zip">Billing ZIP Code</label>
                            <div id="card-zip"><input id="card-zip" type="text" /></div>

                            <button type="submit" id="credit-card-btn" class="doPayment" rel-type="credit" disabled="disabled">DONATE</button>
                        </div>
                    </div>
                </form>
            </div>
    )
}

export default donate 