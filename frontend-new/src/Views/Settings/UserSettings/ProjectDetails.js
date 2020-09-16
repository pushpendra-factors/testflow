import React, {useState, useEffect} from 'react';
import { Row, Col, Modal, Button, Menu, Avatar, Input, Skeleton  } from 'antd';  
import {Text, SVG} from 'factorsComponents';   
import { UserOutlined } from '@ant-design/icons';

function ProjectDetails (props){  

    const [dataLoading, setDataLoading] = useState(true);   
      
    useEffect(() =>{
        setTimeout(() => { 
          setDataLoading(false);
        }, 2000);  
    });

    return (
      <> 
          <div className={`mb-10 pl-4`}>
                <Row>
                  <Col>
                    <Text type={'title'} level={3} weight={'bold'} extraClass={`m-0`}>Your Projects</Text>   
                  </Col>
                </Row>  
                <Row className={`mt-2`}> 
                  <Col span={24}>  
                    { dataLoading ? <Skeleton avatar active paragraph={{ rows: 4 }}/> :
                    <>
                    {[1,2].map((item)=>{
                        return (
                            <div className="flex justify-between items-center border-bottom--thin-2 py-5" >
                                <div className="flex justify-start items-center" >
                                    <Avatar size={60} shape={'square'} />
                                    <div className="flex justify-start flex-col ml-4" >
                                        <Text type={'title'} level={6} weight={`bold`} extraClass={`m-0`}>Project Name-{item}</Text> 
                                        <Text type={'title'} level={7} weight={`regular`} extraClass={`m-0 mt-1`}>Owner</Text> 
                                    </div>
                                </div>
                                <div>
                                    <Button type="text">Leave Project</Button> 
                                </div>
                            </div> 
                        )
                    })}
                    </>
                    }
                  </Col>
                </Row> 
              </div> 

      </>
      
    );
  
}

export default ProjectDetails