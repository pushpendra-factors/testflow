import React from 'react';
import { BrowserRouter, Switch, Route } from 'react-router-dom';
import Home from './Home';
import Blog from './Blog';
import Pricing from './Pricing';
import BlogBigData1 from './BlogBigData1'
import BlogBigData2 from './BlogBigData2'
import ResourcesDropdown from './ResourcesDropdown'
import facebookSVG from './assets/img/facebook.svg'
import linkedinSVG from './assets/img/linkedin.svg'
import logoFactorsPNG from './assets/img/logo_factors.svg'
import twitterSVG from './assets/img/twitter.svg'
import './App.css';
import IntegrationsSegment from './IntegrationsSegment';

function App() {
  return (
    <BrowserRouter>
      <div className="App">
      <nav className="navbar navbar-expand-md navbar-light" style={{ marginTop: '10px' }}>
        <div className="container">
          <a href="/" className="navbar-brand" style={{paddingLeft: '30px'}}>
            <img src={logoFactorsPNG} alt="true" />
          </a>       
          <button className="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarCollapse" aria-controls="navbarCollapse" aria-expanded="false" aria-label="Toggle navigation">
            <span className="navbar-toggler-icon" />
          </button>
          <div className="collapse navbar-collapse" id="navbarCollapse">
            <ul className="navbar-nav mr-auto w-100 justify-content-end clearfix">
              <li className="nav-item">
                <a id="home-nav-link" className="nav-link" href="/">
                  Home
                </a>
              </li>
              <li className="nav-item">
                <a id="blog-nav-link" className="nav-link" href="/pricing">
                  Pricing
                </a>
              </li>
              <li className="nav-item">
               <ResourcesDropdown />
              </li>
            </ul>
          </div>
        </div>
      </nav>

      <Switch>
        <Route exact path="/" component={Home} />
        <Route path="/blog/big-data-analytics-next-1" component={BlogBigData1} />
+       <Route path="/blog/big-data-analytics-next-2" component={BlogBigData2} />
+       <Route path="/blog" component={Blog} />
        <Route path="/pricing" component={Pricing} />
        <Route path="/integrations/segment" component={IntegrationsSegment} />
+     </Switch>

      <div className="container-fluid footer" id="contact">
        <div className="row">
          <div className="container">
            <div className="row">
              <div className="col-md-12">
                <p style={{fontSize: '16px'}}>
                  <i className="fa fa-envelope-o" />  hello@factors.ai
                </p>
                <a id="footer-linkedin" className="linkedin" href="https://www.linkedin.com/company/factors-ai" target="_blank"><img src={linkedinSVG} alt="linkedin" /></a>
                <a id="footer-facebook" className="facebook" href="https://www.facebook.com/factorsai" target="_blank"><img src={facebookSVG} alt="facebook" /></a>
                <a id="footer-twitter" className="twitter" href="https://twitter.com/factorsai" target="_blank"><img src={twitterSVG} alt="twitter" /></a>
                <p className="copyright">© Slashbit Technologies Pvt Ltd</p>
              </div>
            </div>
          </div>
        </div>
      </div>
      </div>
    </BrowserRouter>
  );
}

export default App;
