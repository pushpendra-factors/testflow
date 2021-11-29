import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Progress, Skeleton, Avatar, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';


function Brand() {

    const history = useHistory();

    const handleCreate = (e) => {
        e.preventDefault();
        history.push('/project-setup');
    }

  return (
    <>
      <div className={'fa-container'}>
            <Row justify={'center'}>
                <Col span={7} >
                    <div className={'flex flex-col justify-center mt-20'}>
                        <Row className={'mb-20'}>
                            <Col span={24} >
                                <Text type={'title'} level={3} color={'grey-2'} weight={'bold'}>Brand your Project</Text>
                                <Progress percent={100} status={'normal'} strokeWidth={3} showInfo={false} />
                            </Col>
                        </Row>
                        <Row className={'mt-2'}>
                            <Col>
                                <Text type={'paragraph'} mini extraClass={'m-0 mt-1 mb-4'} color={'grey'} style={{marginLeft:'140px'}}>Project Thumbnail</Text>
                                <img src='../../../../assets/avatar/ModalAvatar.png' style={{marginLeft:'150px'}}></img>
                                <Text type={'paragraph'} mini  extraClass={'m-0 mt-4'} color={'grey'} style={{marginLeft:'80px'}}>A logo helps personalise your Project</Text>
                            </Col>
                        </Row>
                        <Row className={'mt-20'}>
                            <Col>
                                <Button size={'large'} type={'primary'} style={{width:'280px', height:'36px'}} className={'ml-16'} onClick={handleCreate}>Create</Button>
                            </Col>
                        </Row>
                    </div>
                </Col>
            </Row>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
      </div>

    </>

  );
}

export default connect(null, { })(Brand);
