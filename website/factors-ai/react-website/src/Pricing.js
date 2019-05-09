import React, { Component } from 'react';

import freeSVG from './assets/img/pricing/free.svg';
import startupSVG from './assets/img/pricing/startup.svg';
import enterpriseSVG from './assets/img/pricing/enterprise.svg';
import config from './config';

class Pricing extends Component {
  constructor(props) {
    super(props);
  }

  selectPlan = (plan) => {
    let url = config.signup_url;
    url = (!plan || plan == "") ? url : url + '?plan=' + plan;
    window.location.replace(url);
  }
  
  render() {
    return (
      <div className="container nav-exclude-margin container-bottom">
        <div className="row">
          <div className="col-md-12" style={{ marginBottom: '10px' }}>
            <h3 className="centered-title">Pricing</h3>
          </div>
        </div>
        <div className="row">
          <div className="col-md-4">
            <div className="pricing-card">
              <h4>Free</h4>
              <div className="image">
                <img src={freeSVG} />
              </div>
              <div className="no-price"></div>
              <div className="pricing-desc">
                <p>Upto 500K user actions / month</p>
                <p>3 months data retention</p>
                <p>3 member license</p>
                <p>Query (Core analytics)</p>
                <p>Explain (Advanced analytics)</p>
                <p>Dashboards</p>
                <p>Smart Reports</p>
              </div>
              <button onClick={() => { this.selectPlan("free") }} className="no-cta-msg">Get started</button>
            </div>
          </div>
          <div className="col-md-4">
            <div className="pricing-card"> 
              <h4>Startup</h4>
              <div className="image">
                <img src={startupSVG} />
              </div>
              <div className="price">
                <p>$49 <span>/ month</span></p>
                <span>Base price for first 1M user actions</span>
              </div>
              <div className="pricing-desc">
                <p>Upto 5M user actions / month</p>
                <p>1 year data retention</p>
                <p>20 member license</p>
                <p>Query (Core analytics)</p>
                <p>Explain (Advanced analytics)</p>
                <p>Dashboards</p>
                <p>Smart Reports</p>
                <p>Dedicated Customer Success</p>
              </div>
              <div className="cta-msg">
                <p>$10 / 100K user actions for above 1M</p>
              </div>
              <button  onClick={() => { this.selectPlan("enterprise") }}>Start for free</button>
            </div>
          </div>
          <div className="col-md-4">
            <div className="pricing-card">
              <h4>Enterprise</h4>
              <div className="image">
                <img src={enterpriseSVG} />
              </div>
              <div className="no-price"></div>
              <div className="pricing-desc">
                <p>Above 5M user actions per month</p>
                <p style={{ fontWeight: "700", letterSpacing: "0.05rem" }}>Tailored Solutions</p>
              </div>
              <button className="no-cta-msg">Contact us</button>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default Pricing;