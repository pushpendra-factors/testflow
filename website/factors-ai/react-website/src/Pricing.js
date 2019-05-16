import React, { Component } from 'react';

import freeSVG from './assets/img/pricing/free.svg';
import startupSVG from './assets/img/pricing/startup.svg';
import enterpriseSVG from './assets/img/pricing/enterprise.svg';

import featQuerySVG from './assets/img/features/query.svg';
import featFactorSVG from './assets/img/features/factor.svg';
import featDashboardSVG from './assets/img/features/dashboard.svg';
import featJSSDKSVG from './assets/img/features/js_sdk.svg';
import featAndroidSDK from './assets/img/features/android_sdk.svg';
import featIOSSDKSVG from './assets/img/features/ios_sdk.svg';
import featAutotrackSVG from './assets/img/features/autotrack.svg';
import featSegmentSVG from './assets/img/features/segment.svg';

import config from './config';

const FeatureCard = (props) => {
  let customImgHeight = props.imgHeight ? props.imgHeight : null;
  let tag = props.tag ? <span className="tag"> coming soon </span> : null;

  return (
    <div className="feature-card">
      <div style={{ height: customImgHeight }} className="image">
        <img src={props.image}/>
      </div>
      <h5>{props.title}</h5>
      <div>
        <p style={{ marginBottom: props.tag ? '2px' : null }}>{props.children}</p>
      </div>
      { tag }
    </div>
  );
}

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
        <div className="row" style={{ marginBottom: '70px' }}>
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
                <p>Query (Basic analytics)</p>
                <p>Explain (Advanced analytics)</p>
                <p>Dashboards</p>
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
                <p>Query (Basic analytics)</p>
                <p>Explain (Advanced analytics)</p>
                <p>Dashboards</p>
                <p>Smart Reports</p>
                <p>Dedicated Customer Success</p>
              </div>
              <div className="cta-msg">
                <p>$10 / 100K user actions for above 1M</p>
              </div>
              <button  onClick={() => { this.selectPlan("startup") }}>Start for free</button>
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
              <a href="mailto:team@factors.ai"><button className="no-cta-msg">Contact us</button></a>
            </div>
          </div>
        </div>

        <div className="row" style={{ marginTop: '30px' }}>
          <div className="col-md-12" style={{ marginBottom: '10px' }}>
            <h3 className="centered-title">Features</h3>
          </div>
        </div>
        <div className="row">
          <div className="col-md-4">
            <FeatureCard image={featQuerySVG} imgHeight="40px" title="Query">
              Basic analytics to get metrics and charts of users and their actions.
            </FeatureCard>
          </div>
          <div className="col-md-4">
            <FeatureCard image={featFactorSVG} imgHeight="40px" title="Explain">
              Automatic and instant explanations of factors that positively and negatively affect goals.
            </FeatureCard>
          </div>
          <div className="col-md-4">
            <FeatureCard image={featDashboardSVG} imgHeight="40px" title="Dashboard">
              Individual and Team Dashboards to stay on top of all metrics of interest.
            </FeatureCard>
          </div>
          <div className="col-md-4">
            <FeatureCard image={featJSSDKSVG} title="Javascript SDK">
              Simple SDK to identify and track custom user actions on web pages.
            </FeatureCard>
          </div>
          <div className="col-md-4">
            <FeatureCard image={featAutotrackSVG} title="Autotrack">
              Automatically track all web pageviews. Define rules to map them to virtual actions. 
            </FeatureCard>
          </div>
          <div className="col-md-4">
            <FeatureCard image={featAndroidSDK} title="Android SDK">
              Identify and track user actions on android applications.
            </FeatureCard>
          </div>
          <div className="col-md-4">
            <FeatureCard image={featSegmentSVG} title="Segment Integration">
              Hook with Segment to have data flowing in and to start analyzing in no time.
            </FeatureCard>
          </div>
          <div className="col-md-4">
            <FeatureCard image={featIOSSDKSVG} title="IOS SDK" tag="coming soon">
            </FeatureCard>
          </div>
        </div>
      </div>
    );
  }
}

export default Pricing;