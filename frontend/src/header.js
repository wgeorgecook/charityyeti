import React from 'react'
import logo from './img/logo.png'
import headerAbout from './img/header_about.png'
import headerDonate from './img/header_donate.png'

const header = (props) => {
    const headerImg = props.image === 'about' ? headerAbout : headerDonate
    const alt = props.image === 'about' ? 
    'A Yeti holding a bird, which has a coin in its mouth. There are mountains to either side of the yeti.' 
    : 'A yeti doing a handstand, with a bird on its right foot, with footprints where the yeti was walking to the right'
    return (
        <div id='top'>
            <div id='top-bg'></div>
            <header style={{maxWidth: "480px", width: "100%", margin: "0 auto"}}>
                <img src={logo} alt='The Charity Yeti logo' class="logo" />
                <img src={headerImg} alt={alt} class="header" />
            </header>
        </div>
    )
}

export default header