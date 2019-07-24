import React from 'react'
import './assets/css/blog.css'
import bigData1PNG from './assets/img/blog/big_data_analytics_1.jpg'
import bigData2PNG from './assets/img/blog/big_data_analytics_2.jpg'
import segmentIntegrationPNG from './assets/img/blog/blog_segment_integration.png'


class Blog extends React.Component {
    render() {
      return (
        <section id="blog" className="section-padding main-container">
          <div className="container">
            <div className="section-header text-center">
              <h2 className="section-title">Insights, Solutions and Updates from our Team.</h2>
            </div>
            <div className="row">
            <div className="col-lg-4 col-md-6 col-sm-12 col-xs-12 blog-item">
                <div className="blog-item-wrapper">
                  <div className="blog-item-img">
                    <a href="/blog/segment-integration-launch">
                      <img src={segmentIntegrationPNG} alt />
                    </a>
                  </div>
                  <div className="blog-item-text">
                    <h3>
                      <a href="/blog/segment-integration-launch">Announcing our partnership with Segment</a>
                    </h3>
                    <a href="/blog/segment-integration-launch" className="btn btn-common btn-rm">Read More</a>
                  </div>
                </div>
              </div>
              <div className="col-lg-4 col-md-6 col-sm-12 col-xs-12 blog-item">
                <div className="blog-item-wrapper">
                  <div className="blog-item-img">
                    <a href="/blog/big-data-analytics-next-1">
                      <img src={bigData1PNG} alt />
                    </a>                
                  </div>
                  <div className="blog-item-text"> 
                    <h3>
                      <a href="/blog/big-data-analytics-next-1">What's next in Big Data and Analytics? (Part 1)</a>
                    </h3>
                    <a href="/blog/big-data-analytics-next-1" className="btn btn-common btn-rm">Read More</a>
                  </div>
                </div>
              </div>
              <div className="col-lg-4 col-md-6 col-sm-12 col-xs-12 blog-item">
                <div className="blog-item-wrapper">
                  <div className="blog-item-img">
                    <a href="/blog/big-data-analytics-next-2">
                      <img src={bigData2PNG} alt />
                    </a>                
                  </div>
                  <div className="blog-item-text"> 
                    <h3>
                      <a href="/blog/big-data-analytics-next-2">What's next in Big Data and Analytics? (Part 2)</a>
                    </h3>
                    <a href="/blog/big-data-analytics-next-2" className="btn btn-common btn-rm">Read More</a>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      );
    }
}

export default Blog;
