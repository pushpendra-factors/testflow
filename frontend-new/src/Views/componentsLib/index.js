import React from 'react';
import { Layout, Breadcrumb, Row, Col, Divider,Skeleton  } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery'; 
import {Text} from 'factorsComponents';
import { Link } from 'react-router-dom';

function componentsLib() {
    const { Content } = Layout;
  return ( 
        <Layout>
        <Sidebar />
                <Layout className="fa-content-container">
                <Content> 
                    <div className="px-10 py-5 bg-white min-h-screen"> 


                    <Divider orientation="left">
                        <Breadcrumb> 
                            <Breadcrumb.Item>
                            <Link to="/">Home</Link>
                            </Breadcrumb.Item>
                            <Breadcrumb.Item> Components Lib </Breadcrumb.Item> 
                        </Breadcrumb> 
                    </Divider>
                    

                    <Row>
                        <Col span={12}>
                            <Text type={'title'} level={1} >Heading Style - Title1/40</Text> 
                            <Text type={'title'} level={2} weight={'bold'}>Heading Style - Title2/32</Text> 
                            <Text type={'title'} level={3} >Heading Style - Title3/24</Text>
                            <Text type={'title'} level={4} >Heading Style - Title4/20</Text>
                            <Text type={'title'} level={5} >Heading Style - Title5/18</Text>
                            <Text type={'title'} level={6} >Heading Style - Title6/16</Text>

                            <Text type={'title'} level={6} >Use for Headings - subtitle1/16</Text>
                            <Text type={'title'} level={7} >Use for Headings - subtitle2/14</Text> 


                            <Text type={'paragraph'}>Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.</Text>

                            <Text type={'paragraph'} size={'7'} mini >Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.</Text>
                            
                        </Col> 
                        <Col span={12}>
                            <Text type={'title'} level={1} >Heading Style - Title1/40</Text>
                            <Text type={'title'} level={1} weight={'thin'} >Heading Style - Title1/40</Text>
                            <Text type={'title'} level={1} weight={'bold'} >Heading Style - Title1/40</Text>
                            <Text type={'title'} level={2} >Heading Style - Title2/32</Text> 
                            <Text type={'title'} level={2} weight={'bold'}>Heading Style - Title2/32</Text> 
                            <Text type={'title'} level={3} >Heading Style - Title3/24</Text>
                            <Text type={'title'} level={4} >Heading Style - Title4/20</Text>
                            <Text type={'title'} level={5} >Heading Style - Title5/18</Text>
                            <Text type={'title'} level={6} >Heading Style - Title6/16</Text>

                            <Text type={'title'} level={6} >Use for Headings - subtitle1/16</Text>
                            <Text type={'title'} level={7} >Use for Headings - subtitle2/14</Text> 


                            <Text type={'paragraph'}>Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.</Text>

                            <Text type={'paragraph'} size={'7'} mini >Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.</Text>
                            
                        </Col> 
                    </Row>
                

                    
                    <div className="my-6">
                       
                    </div>


                    </div> 
                </Content>
            </Layout>
        </Layout>

  );
}

export default componentsLib;
