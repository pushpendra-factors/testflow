import React, { Component } from 'react'
import './assets/css/blog.css'
import enableSegmentOnFactorsPNG from './assets/img/integrations/enable-segment-on-factors.png'
import enableFactorsDestination1PNG from './assets/img/integrations/enable-factors-destination-1.png'
import enableFactorsDestination2PNG from './assets/img/integrations/enable-factors-destination-2.png'



class IntegrationsSegment extends Component {
    render() {

      return (
        <div id="blog-single">
        <div className="container">
          <div className="row justify-content-center">
            <div className="col-lg-12 col-md-16 col-xs-16">
              <div className="blog-post">
                <div className="post-content">
                  <h3>Integrating with Segment</h3>
                  <p>
                    FactorsAI automatically connects to your user data in Segment. Segment allows you to easily manage integrations with multiple analytics services. By tracking events and users via Segment’s API and libraries, you can send your product’s data to all of your analytics/marketing platforms, with minimal instrumentation code. We collect all your Segment identifies, pages, screens, and track events and is immediately available on FactorsAI for analysis.
                  </p>
                  <br />
                  <br />
                  <h5>Step 1: Setup Segment Integration on FactorsAI</h5>
                  <p><ul>
                      <li>
                        Sign up on FactorsAI.
                      </li>
                      <li>
                        Create a project for your data on FactorsAI.
                      </li>
                      <li>
                        Go to Settings > Segment for the created project.
                      </li>
                      <li>
                        Enable the Segment Integration.
                      </li>
                      <li>
                        Note down the API Key for the integration.
                      </li>
                  </ul></p>
                  <div className="container">
                    <img style={{width: '95%', padding: '1px', border: '1px solid #021a40'}}
                         src={enableSegmentOnFactorsPNG}/>
                  </div>
                  <br />
                  <h5>Step 2: Setup Destination FactorsAI on Segment</h5>
                  <p><ul>
                      <li>
                        Login to Segment. Choose the appropriate Workspace.
                      </li>
                      <li>
                        Click on Destinations > Add Destination.
                      </li>
                      <li>
                        Search for factorsai in catalog and click on FactorsAI search result.  
                      </li>
                      <li>
                        Click on "Configure FactorsAI".
                      </li>
                      <li>
                        Select the appropriate Source. (Ex: Javascript). Click on Confirm Source.
                      </li>
                      <li>
                        Click on API Key. Enter the API Key noted down from FactorsAI account.
                      </li>
                  </ul></p>
                  <div className="container">
                    <img style={{width: '75%', padding: '1px'}}
                         src={enableFactorsDestination1PNG}/>
                    <img style={{width: '75%', padding: '1px',}}
                         src={enableFactorsDestination2PNG}/>
                  </div>
                  <br />
                  <p>
                        After you've created your Segment + FactorsAI integration, we'll immediately start ingesting your user data. We'll need a minimum of 1 week of data to build descriptive models and give deeper insights for your goals.
                  </p>
                  <br />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      );
    }
}

export default IntegrationsSegment;