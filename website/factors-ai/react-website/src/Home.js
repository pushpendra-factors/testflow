// Home.js
import React from 'react';
import factorsPNG from './assets/img/factors_app.png'
import tabMeasureSVG from './assets/img/tab_measure.svg'
import tabTrackSVG from './assets/img/tab_track.svg'
import tabDiscoverSVG from './assets/img/tab_discover1.svg'
import heroSVG from './assets/img/hero.svg'

class Home extends React.Component {
  render() {
    return (
        <div>
          {/* hero area */}
          <div className="container nav-exclude-margin hero" id="home">
            <div className="row">
              <div className="col-md-5 content">
                <h2 className="heading">Does Analytics feel like a lot of work?</h2>
                <div className="break" />
                <p className="sub-text">Automatically discover what’s working and not working for your users.</p>
                <span className="error" id="invalid-email" style={{display: 'none'}}>Please enter a valid email</span>
                <span className="success" id="valid-email" style={{display: 'none'}}>Thanks, we will get back to you soon</span>
                <input id="work-email" placeholder="Your work email" />
                <button id="get-invite-button">Get Invite</button>
              </div>
              <div className="col-md-7 content">
                <img src={heroSVG} />
              </div>
            </div>
          </div>
          {/* hero banner */}
          <div className="container-fluid">
            <div className="row">
              <div className="col-md-12 banner top-banner">
                <h3>Goal Driven Analytics</h3>
                <p>Introducing the new way to do analytics</p>
              </div>
            </div>
          </div>
          {/* goal driven analytics description */}
          <div className="container-fluid shaded-background">
            <div className="row">
              <div className="container gda">
                <div className="row">
                  <div className="col-md-12">
                    <p>Getting insights out of user data can’t be easier than this. Enter your goal and instantly get factors affecting it. Avoid the labor of guessing multiple hypothesis and firing multiple queries, plotting multiple charts or spending hours viewing multiple individual user sessions to validate your theories. Get answers for a completely different class of questions, compared to traditional analytics products.</p>
                  </div>
                </div>
                <div className="row tabular justify-content-center">
                  <div className="col-md-4 content">
                    <h5>Traditional Products</h5>
                    <span>How many signups?</span><br />
                    <span>What’s the revenue?</span><br />
                    <span>What’s the week one retention?</span><br />
                    <span>What’s the number of orders?</span><br />
                  </div>
                  <div className="col-md-4 content">
                    <h5>FactorsAI</h5>                
                    <span>How to increase signups?</span><br />
                    <span>How to increase revenue?</span><br />
                    <span>How to increase week one retention?</span><br />
                    <span>How to increase the number of orders?</span><br />
                  </div>
                </div>
              </div>
            </div>
          </div>
          {/* features */}
          <div className="container tabs">
            <div className="row header no-mobile">
              <div className="col-md-4">
                <h4 id="feat_track" className="active-tab">Capture</h4>
              </div>
              <div className="col-md-4">
                <h4 id="feat_measure">Measure</h4>            
              </div>
              <div className="col-md-4">
                <h4 id="feat_discover">Discover</h4>            
              </div>
            </div>
            <div className="row content" id="tab_track">
              <div className="col-md-5 txt">
                <h4>Capture your customer data</h4>
                <p>Track and define all relavant user actions data using simple to add SDKs.</p>
                <p>Pull data from multiple sources using rich set of integrations.</p>
              </div>
              <div className="col-md-7 image">
                <img src={tabTrackSVG} style={{width: '100%'}} />
                <div id="gda"> </div> {/* anchor for gda */}
              </div>
            </div>
            <div className="row content" id="tab_measure" style={{display: 'none'}}>
              <div className="col-md-5 txt">
                <h4>Measure your Product's performance</h4>
                <p>Define and measure any metric of interest - Visits, Signups, Funnel Conversions, Feature usage, Retention, Revenue or Profits. </p>
                <p>Setup dashboards to keep real time tab on metrics.</p>
              </div>
              <div className="col-md-7 image">
                <img src={tabMeasureSVG} style={{width: '100%'}} />
                <div id="gda"> </div> {/* anchor for gda */}
              </div>
            </div>
            <div className="row content" id="tab_discover" style={{display: 'none'}}>
              <div className="col-md-5 txt">
                <h4>Discover what is working and not working</h4>
                <p>Automatically discover user behaviors that positively and negatively affect your goals. Further, instantly drill down goals to sub-goals and discover influencing sub-factors.</p>
                <p>Get insights in understandable natural language, aided with relevant charts.</p>
              </div>
              <div className="col-md-7 image">
                <img src={tabDiscoverSVG} style={{width: '85%'}} />
                <div id="gda"> </div> {/* anchor for gda */}
              </div>
            </div>
          </div>
          {/* tech blog */}
          <div className="container-fluid shaded-background">
            <div className="row">
              <div className="container blog">
                <div className="row">
                  <div className="col-md-12">
                    <h4 className="centered-title">AI that makes sense of data</h4>
                    <p>Take analytics beyond raw storage and retreival of data - using our inhouse datastore that uses AI and advanced statistical modeling techinques to make sense of your data. This next generation, first of it's kind datastore, pulls out probabilistic inferences from data for a given query - rather than just pulling out the data points that match the query. Built to seamlessly scale to billions of data points.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          {/* screenshot */}
          <div className="container" style={{paddingTop: '50px', paddingBottom: '60px'}}>
            <div className="container screenshot">
              <div className="row">
                <div className="col-md-12">
                  <h4 className="centered-title">Enable Data Driven Decision Making</h4>
                </div>
              </div>
              <div className="row">
                <div className="col-md-12" align="center">
                  <img src={factorsPNG} />
                </div>
              </div>
            </div>
          </div>
          {/* signup banner */}
          <div className="container-fluid bottom-banner">
            <div className="row">
              <div className="col-md-12 banner">
                <h3>FactorsAI is now available on request</h3>
                <button id="bottom-signup">Get early access</button>
              </div>
            </div>
          </div>
        </div>
      );
  }
}

export default Home;