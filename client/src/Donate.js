import React, { Component } from 'react';

export default class Donate extends Component {

  componentDidMount = () => {
    const getParams = new URLSearchParams(window.location.search)

    const honorary = getParams.get('honorary')
    const invoker = getParams.get('invoker')
    const originalTweetID = getParams.get('originalTweetID')
    const invokerTweetID = getParams.get('invokerTweetID')

    this.setState({
      honorary,
      invoker,
      originalTweetID,
      invokerTweetID
    })

  }

  updateDonation = (e) => {
    this.setState({
      donationValue: e.target.value
    })
  }

  // TODO: This needs to be more robust, and have proper notifications
  // TODO: We can include all of the different wallet and cash transfer apis
  sendData = (e) => {
    e.preventDefault()
    fetch(`http://localhost:8080?invoker=${this.state.invoker}&honorary=${this.state.honorary}&invokerTweetID=${this.state.invokerTweetID}&originalTweetID=${this.state.originalTweetID}&donationValue=${this.state.donationValue}`)
    .then( r => {
      console.log(r)
      if (r.status === 200) {
        this.setState({ 'success': true})
      }
    })
  }

  render() {
    // TODO: Obviously this doesn't actually do anything. We need to include the various ways to handle donations
    return (
      <div className="donate">
        <h1>Donate Money</h1>
        <div className={"donationForm"}>
        {
          (this.state && this.state.success)
          ?
            <div>Thank you for your ${this.state.donationValue} donation!</div>
          :
            <form>
              <label>How much?
                <input
                  type="number"
                  name="donate"
                  onChange={this.updateDonation}
                />
              </label>
              <input
                type="submit"
                value="Confirm donation"
                onClick={this.sendData}
              />
            </form>
        }
        </div>
      </div>
    )
  }
}
