import React, { useState, useCallback, useEffect } from "react";
import { Row, Col, Tabs, Modal, notification, Input, Checkbox} from "antd";
import { useSelector, useDispatch } from "react-redux";
import { Text ,SVG} from "../../../components/factorsComponents";
import styles from "./index.module.scss";


const {TextArea} = Input;

function AddDashboardDetails({AddDashboardDetailsVisible,setAddDashboardDetailsVisible,AddReportsVisible,setAddReportsVisible}){
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const handleOk = ()=>{
        setAddDashboardDetailsVisible(false);
        setAddReportsVisible(true);
    }
    const handleCancel=()=>{
        setAddDashboardDetailsVisible(false);
    }
    return(
            <Modal        
                title={"Add New Dashboard"}
                centered={true}
                zIndex={1005}
                width={700}
                onCancel={handleCancel}
                onOk={handleOk}
                className={"fa-modal--regular p-4 fa-modal--slideInDown"}
                // confirmLoading={apisCalled}
                closable={false}
                okText={"Create"}
                cancelText={"Close"}
                transitionName=""
                maskTransitionName=""
                okButtonProps={{ size: "large" }}
                cancelButtonProps={{ size: "large" }}
                visible={AddDashboardDetailsVisible}>
                <Row className={'pt-4'} gutter={[24, 24]}>
                    <Col span={24}>
                        <Input onChange={(e) => setTitle(e.target.value)} value={title} className={'fa-input'} size={'large'} placeholder="Dashboard Title" />

                    </Col>
                    <Col span={24}>
                        <TextArea rows={4} onChange={(e) => setDescription(e.target.value)} value={description} className={'fa-input'} size={'large'} placeholder="Description (Optional)" />
                    </Col>
                </Row>
            </Modal>
    );
}

export default AddDashboardDetails;