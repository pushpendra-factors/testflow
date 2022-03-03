import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Progress, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import InviteMembers from './InviteMembers';
import styles from './index.module.scss';


function Congrates({handleCancel}) {
    const [showInvite, setShowInvite] = useState(false);
    const history = useHistory();

    const handleContinue = () => {
        handleCancel();
        history.push('/project-setup');
    }

    const handleInvite = () => {
        setShowInvite(true);
    }

  return (
    <>
    {!showInvite &&
      <div className={'fa-container'}>
            <Row justify={'center'} className={`${styles.start}`}>
                <Col span={7} >
                    <div className={'flex flex-col justify-center mt-16'}>
                        <Row className={'mb-1'}>
                            <Col span={24}>
                                <img src='assets/images/Illustration=pop gift.png' style={{width: '100%',maxWidth: '80px', marginLeft:'11vw'}}/>
                                <Text type={'title'} level={3} color={'grey-2'} align={'center'} weight={'bold'}>Project created succesfully</Text>
                                {/* <Progress percent={100} status={'normal'} strokeWidth={3} showInfo={false} /> */}
                            </Col>
                        </Row>
                        <Row className={'mt-1'}>
                            <Col>
                                <Text type={'paragraph'} mini  extraClass={'m-0'} color={'grey'} weight={'bold'} style={{textAlign:'center'}}>Congratulations! Your project has been created.It’s now time to connect to your data sources</Text>
                                {/* <Text type={'paragraph'} mini  extraClass={'m-0'} color={'grey'} weight={'bold'} style={{ textAlign:'center'}}></Text> */}
                            </Col>
                        </Row>
                        <Row className={'mt-6'}>
                            <Col>
                                <Button size={'large'} type={'primary'} style={{width:'27vw', height:'36px'}} className={'m-0'} onClick={handleContinue}>Continue to project</Button>
                            </Col>
                        </Row>
                        <Row className={'mt-4'}>
                            <Col>
                                <Button size={'large'} type={'link'} style={{width:'27vw', height:'36px', backgroundColor:'white'}} className={'m-0'} onClick={handleInvite}>Invite my teammates</Button>
                            </Col>
                        </Row>
                    </div>
                </Col>
            </Row>
            <div className={`${styles.hideSVG}`}>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
            </div>
      </div>
    } {
        showInvite && <InviteMembers handleCancel = {handleCancel} />
    }
    </>

  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});

export default connect(mapStateToProps, { })(Congrates);
