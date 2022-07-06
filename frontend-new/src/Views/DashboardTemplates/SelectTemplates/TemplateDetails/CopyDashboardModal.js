import React, { useState, useCallback, useEffect } from "react";
import { Row, Col, Tabs, Modal, notification, Input, Checkbox} from "antd";
import {Text, SVG} from '../../../../components/factorsComponents';
import { useSelector, useDispatch } from 'react-redux';

import {createDashboardFromTemplate} from '../../../../reducers/dashboard_templates/services'


function CopyDashboardModal({showCopyDashBoardModal,setShowCopyDashBoardModal}){
    const { active_project } = useSelector((state) => state.global);
    const {activeTemplate} = useSelector((state)=>state.dashboardTemplates);
    const handleOk = async()=>{
        try{
            const res = await createDashboardFromTemplate(active_project.id,activeTemplate.id);
            alert('Copy of Dasboard Created!');

        }catch (err){
            console.log('raghav',err.response);
        }
        setShowCopyDashBoardModal(false);
    }
    const handleCancel=()=>{
        setShowCopyDashBoardModal(false);
    }
    return(
            <Modal        
                centered={true}
                zIndex={1005}
                width={'30%'}
                onCancel={handleCancel}
                onOk={handleOk}
                className={"fa-modal--regular p-4 fa-modal--slideInDown"}
                // confirmLoading={apisCalled}
                closable={true}
                okText={"Create Copy"}
                cancelText={"Cancel"}
                okButtonProps={{ size: "large"}}
                cancelButtonProps={{ size: "large" }}
                visible={showCopyDashBoardModal}>
                <Row className={'pt-4'} >
                    <Col >
                        <Text type='title' level={4} weight={'bold'}>Do you want to create a copy?</Text>
                    </Col>
                    <Col >
                        <Text type='paragraph' level={7} color={'grey'} weight={'bold'}>Creating a copy will replicate the dashboard into your Project</Text>
                    </Col>
                </Row>
            </Modal>
    );
}

export default CopyDashboardModal;