import React from "react";
import { Row, Col, Modal} from "antd";
import {Text } from '../../../../components/factorsComponents';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';

import {createDashboardFromTemplate} from '../../../../reducers/dashboard_templates/services'


function CopyDashboardModal({showCopyDashBoardModal,setShowCopyDashBoardModal}){
    const history = useHistory();
    const { active_project } = useSelector((state) => state.global);
    const { activeTemplate } = useSelector((state)=>state.dashboardTemplates);
    const handleOk = async()=>{
        try{
            const res = await createDashboardFromTemplate(active_project.id,activeTemplate.id);
            history.push('/');

        }catch (err){
            console.log(err.response);
        }
        setShowCopyDashBoardModal(false);
    }
    const handleCancel=()=>{
        setShowCopyDashBoardModal(false);
    }
    return(
            <Modal        
                centered={true}
                width={'30%'}
                onCancel={handleCancel}
                onOk={handleOk}
                className={"fa-modal--regular p-4 fa-modal--slideInDown"}
                closable={true}
                okText={"Create Copy"}
                cancelText={"Cancel"}
                okButtonProps={{ size: "large"}}
                cancelButtonProps={{ size: "large" }}
                transitionName=""
                maskTransitionName=""
                visible={showCopyDashBoardModal}>
                <Row className={'pt-4'} >
                    <Col >
                        <Text type='title' level={4} weight={'bold'}>Do you want to create a copy?</Text>
                    </Col>
                    <Col >
                        <Text type='paragraph' level={7} color={'grey'} weight={'bold'}>Creating a copy will replicate this dashboard into your Project</Text>
                    </Col>
                </Row>
            </Modal>
    );
}

export default CopyDashboardModal;