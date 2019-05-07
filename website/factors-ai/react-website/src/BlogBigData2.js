import React from 'react'
import Disqus from "disqus-react"
import './assets/css/blog.css'
import aravindAuthorPNG from './assets/img/blog/aravind_author.jpg'
import bigDataTechnologiesPNG from './assets/img/blog/big_data_technologies.png'

class BlogBigData2 extends React.Component {
    render() {
      const disqusShortname = 'factorsai';
      const disqusConfig = {
        url: 'www.factors.ai/blog/big-data-analytics-next-2',
        identifier: 'big-data-analytics-next-2',
      };

      return (
        <div id="blog-single">
        <div className="container">
          <div className="row justify-content-center">
            <div className="col-lg-12 col-md-16 col-xs-16">
              <div className="blog-post">
                <div className="post-content">
                  <h3>Big Data and Analytics - What's next?  (Part 2)</h3>
                  <p>
                    In the <a href="big-data-analytics-next-1">previous blog</a>, we very briefly went over the history of Big Data Technologies. We saw how databases evolved from relational databases to NoSQL databases like Bigtable, Cassandra, DynamoDB etc with the rise of internet along with development of technologies like GFS, MapReduce etc for distributed file storage and computation. These technologies were first developed by companies like Google, Amazon etc and later picked up in a big way by the open source community.
                  </p>
                  <br />
                  <div className="post-thumb">
                    <img src={bigDataTechnologiesPNG} alt />
                  </div>
                  <br />
                  <h5>Big Data and Enterprises</h5>
                  <p>
                    Soon enough commercial versions of these open source technologies were being distributed by companies like Cloudera, Hortonworks etc. Traditional enterprises started adopting these technologies for their analytics and reporting needs.
                  </p>
                  <br />
                  <p>
                    Prior to this enterprises built data warehouses which were actually large relational databases. It involved combining data from multiple databases of ERP, CRM etc and build an unified and relatively denormalized database. Designing the data warehouse was complex and required careful thought. Data was updated periodically. Updation involved a three stage process of extracting data from various sources, combining and transforming these to the denormalized format and loading it into the data warehouse. This came to known as ETL (Extract, Transform and Load).
                  </p>
                  <br />
                  <p>
                    With adoption of Hadoop, enterprises could now just periodically dump all their data into a cluster of machines and run ad-hoc run map reduces to pull out any report of interest. Visualization tools like Tableau, PowerBI, Qlik etc could connect directly to this ecosystem, making it seamless to plot graphs from a simple interface, but actually done by crunching large volumes of data in the background.
                  </p>
                  <br />
                  <h5>Customer Centric View of Data</h5>
                  <p>
                    Databases are a final system of record and analytics on databases only gives information on the current state of customers and not how they reached here.  With the rise of internet a lot of businesses are now online, or have multiple digital touchpoints with customers. Now it's easier to instrument and collect customer data as a series of actions, be it clickstream or online transactions. This customer centric model of data enables richer analytics and insights. Additionally the data is incremental, and can be made available immediately in reports, instead of being updated only periodically. More enterprises are moving to this model and datastores and technologies that cater specifically to these kind of use cases are actively being developed like TimescaleDB, Druid, Snowplow etc.
                  </p>
                  <br />
                  <h5>So what’s next?</h5>
                  <p>
                    To summarize, the bulk of the big data revolution, that has happened in the last 15 years, is to build systems capable of storing and querying large amounts of data. The queries are raw i.e if X and Y are variables in the data and x1 and y2 are two corresponding values of interest, then the system can return all data points where in the variable X matches x1 and Y matches y2. Or some post processed result on all the matching data points. Along the way, we also have systems that can compute on large amounts of data in a distributed fashion.
                  </p>
                  <br />
                  <p>
                    So what’s next in analytics from here? Is it building machine learning models? Certainly, the availability of all these data, enables organizations to build predictive models for specific use cases. In fact, the recent surge of interest in machine learning has actually been because of the better results we get by running the old ML algorithms at larger scale in a distributed way. While most ML techniques can be used to build offline models to power predictive features, it is not useful in the context of online or interactive analytics. Most techniques are particularly designed for high dimensional unstructured data like language or images, where the challenge is not only to build models that fit well on seen data points, but also generalizes well to hitherto unseen data points.
                  </p>
                  <br />
                  <h5>Datastores that make sense of data</h5>
                  <p>
                    The next logical step would be datastores and systems that can make sense of data. Making sense of data would mean that instead of blindly pulling out data points such that variable X is x1 and Y to y2, it should also be able to interactively answer different class of queries like
                  </p><ul>
                    <li>
                      Give the best value for variable Y,  that maximizes the chance that X is x1.
                    </li>
                    <li>
                      Find all the variables or combination of variables, that influence X most when X is x1.
                    </li>
                  </ul>
                  <p />
                  <br />
                  <p>
                    Such a system would continuously build a complete statistical or probabilistic model as and when data gets added or updated. Models would be descriptive and queryable. The time taken to infer or answer the different class of queries should also be tractable.  But just like there are a host of databases each tuned differently for 
                  </p><ul>
                    <li>Data Model</li>
                    <li>Scale</li>
                    <li>Read and Write Latencies</li>
                    <li>Transaction guarantees</li>
                    <li>Consistency, etc </li>
                  </ul>
                  <p />
                  <br />
                  <p>
                    We could possibly have different systems here tuned for
                  </p><ul>
                    <li>Assumptions on Data Model</li>
                    <li>Accuracy</li>
                    <li>Ability to Generalize</li>
                    <li>Scale of the data</li>
                    <li>Size of the models</li>
                    <li>Time taken to evaluate different types of queries.</li>
                  </ul>
                  <p />
                  <br />
                  <p>
                    Autometa - is one such, first of it’s kind, system that we are building at factors.ai. It continuously makes sense of customer data to reduce the work involved in inferring from data. Drop in a mail to hello@factors.ai to know more or to give it a try. 
                  </p>
                </div>
              </div>
              <div className="blog-comment">
                <div className="the-comment">
                  <div className="avatar">
                    <img src={aravindAuthorPNG} alt />
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

export default BlogBigData2;