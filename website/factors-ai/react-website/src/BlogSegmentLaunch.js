import React, { Component } from 'react'
import Disqus from "disqus-react"
import './assets/css/blog.css'
import aravindAuthorPNG from './assets/img/blog/aravind_author.jpg';
import segmentFactorsAISVG from './assets/img/blog/blog-segment-launch.svg';

class BlogSegmentLaunch extends Component {
    render() {
      const disqusShortname = 'factorsai';
      const disqusConfig = {
          url: 'www.factors.ai/blog/segment-integration-launch',
          identifier: 'segment-integration-launch',
      };

      return (
        <div id="blog-single" className="main-container">
        <div className="container">
          <div className="row justify-content-center">
            <div className="col-lg-12 col-md-16 col-xs-16">
              <div className="blog-post">
                <div className="post-content">
                  <h3>FactorsAI + Segment: Easy and instant analytics to drive growth</h3>
                  <p>
                  We are excited to announce our integration with Segment, further enabling companies to easily instrument user interactions across platforms and push different types of customer data, from any 3rd party source in realtime to FactorsAI.
                  </p>
                  <div className="col-md-12" style={{textAlign: "center", marginTop: "70px", marginBottom: "65px"}}>
                    <img src={segmentFactorsAISVG}/>
                  </div>
                  <br />
                  <p>
                  FactorsAI provides advanced and intuitive analytics for marketers and product managers, to help drive growth. With FactorsAI you get immediate insights to optimize marketing campaigns, improve conversions and understand user behaviours that drive feature adoption and retention.
                  </p>
                  <br />
                  <p>
                  A good analytics setup requires detailed tracking of user actions like page views, Signups, AddedToCart with different attributes. The quality of insights on user behaviour shown by FactorsAI is dependent on the level of detail in tracking. With Segment integration this is a one time setup and you could send the same events to other tools for marketing automation, CRM etc.        
                  </p>
                  <br />
                  <p>Further with Segment integration, you can send data from different data sources like email, livechat which will send events like Email Delivered, Email Clicked, Live Chat Started etc. These additional events are useful when analyzing user conversions and by using Segment it can be done without the need to write custom code to hit our API’s.</p>
                  <br />
                  <p>Segment can perform all data collection tasks for FactorsAI. It can capture all the data that FactorsAI needs and sends it directly to FactorsAI in the right format, all in real-time. So, if you are on segment, you can now start getting insights on how to grow your customer base in no time.</p>
                  <br />
                  <p>To integrate with Segment, follow the steps <a href="../integrations/segment">here</a>. Happy Analyzing!</p>
                </div>
              </div>
              <div className="blog-comment">
                <div className="the-comment">
                  <div className="avatar">
                    <img src={aravindAuthorPNG} />
                  </div>
                  <div className="comment-box">
                    <div className="comment-author">
                      <h5>Aravind Murthy</h5>
                      <span>Founder, factors.ai</span>
                    </div>
                  </div>
                </div>
              </div>       
              <div className="blog-comment">
              <div id="disqus_thread" />
                <Disqus.DiscussionEmbed shortname={disqusShortname} config={disqusConfig} />
              </div>
            </div>
          </div>
        </div>
      </div>
      );
    }
}

export default BlogSegmentLaunch;