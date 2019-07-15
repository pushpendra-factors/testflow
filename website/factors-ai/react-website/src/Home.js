// Home.js
import React from 'react';

import advancedAnalyticsPNG from './assets/img/advanced-analytics.png'
import basicAnalyticsPNG from './assets/img/basic-analytics.png'
import dashboardsPNG from './assets/img/dashboards.png'
import funnelPNG from './assets/img/funnel.png'

import heroSVG from './assets/img/hero.svg'

class Home extends React.Component {
  render() {
    return (
        <div>
          {/* hero area */}
          <div className="container nav-exclude-margin hero" id="home">
            <div className="row">
              <div className="col-md-5 content">
                <h2 className="heading">Intelligent user analytics to help grow your business</h2>
                <div className="break" />
                <p className="sub-text">Software that <b>explains</b> - how to optimise marketing campaigns, improve website conversions and drive user engagement.</p>
                <span className="error" id="invalid-email" style={{display: 'none'}}>Please enter a valid email</span>
                <span className="success" id="valid-email" style={{display: 'none'}}>Thanks, we will get back to you soon</span>
                <button id="signup-button" className='primary-cta'>SIGN UP FOR FREE</button>
              </div>
              <div className="col-md-7 content">
                <img src={heroSVG} />
              </div>
            </div>
          </div>
          {/* hero banner */}
          <div className="container-fluid">
            <div className="row">
              <div className="col-md-12 banner banner-dark top-banner">
                <h3>Goal Driven Analytics</h3>
                <p>Introducing the new way to do analytics</p>
              </div>
            </div>
          </div>
          <div className="container-fluid shaded-background">
           <div className="container gda">
            <div className="row">
              <div className="col-md-3 top-margin">
                <h5>Goal Driven Analytics</h5>
                <p>Enter your goal and instantly get factors affecting it.</p>
                <p>Avoid spending hours slicing and dicing the data or viewing multiple user sessions.</p>
              </div>
              <div className="col-md-9 top-margin image screenshot">
                <img src={advancedAnalyticsPNG} style={{width: '100%'}} />
              </div>
            </div>
          </div>
          </div>
          <div className="container-fluid">
           <div className="container gda">
           <div className="row">
              <div className="col-md-9 top-margin image screenshot">
                <img src={basicAnalyticsPNG} style={{width: '100%'}} />
              </div>
              <div className="col-md-3 top-margin">
                <h5>Basic Analytics</h5>
                <p>Flexible query interface and visualizations to allow for in-depth and granular analysis.</p>
              </div>
            </div>
            </div>
          </div>
          <div className="container-fluid shaded-background">
           <div className="container gda">
            <div className="row">
              <div className="col-md-3 top-margin">
                <h5>Custom Dashboards</h5>
                <p>Intuitive and realtime dashboards to stay on top of your metrics and KPI’s.</p>
              </div>
              <div className="col-md-9 top-margin image screenshot">
                <img src={dashboardsPNG} style={{width: '100%'}} />
              </div>
            </div>
          </div>
          </div>
          <div className="container-fluid">
           <div className="container gda">
           <div className="row">
              <div className="col-md-9 top-margin image screenshot">
                <img src={funnelPNG} style={{width: '100%'}} />
              </div>
              <div className="col-md-3 top-margin">
                <h5>Customer Journey Funnels</h5>
                <p>Track funnels and the conversion ratios at every step of the customer journey.</p>
              </div>
            </div>
            </div>
          </div>
          {/* tech blog */}
          <div className="container-fluid shaded-background">
            <div className="row">
              <div className="container blog">
                <div className="row">
                  <div className="col-md-12">
                    <h4 className="centered-title">AI to help make sense of data</h4>
                    <p>What makes us truly unique is that we are taking analytics beyond raw storage and retreival of data. Using our inhouse datastore, that uses advanced statistical modeling techinques, to help understand user behavioural data.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          {/* signup banner */}
          <div className="container-fluid">
            <div className="row">
              <div className="col-md-12 banner banner-light bottom-banner">
                <h3>Try FactorsAI today</h3>
                <button id="signup-button-footer" className="primary-cta">SIGN UP NOW</button>
              </div>
            </div>
          </div>
        </div>
      );
  }
}

export default Home;
