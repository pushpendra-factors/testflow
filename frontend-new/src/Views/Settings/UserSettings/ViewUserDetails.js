import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton
} from 'antd';
import { Text } from '../../../components/factorsComponents';
import { EditOutlined} from '@ant-design/icons';
import { fetchAgentInfo } from '../../../reducers/agentActions';
import { connect } from 'react-redux';
import {udpateProjectDetails} from 'Reducers/global'
function ViewUserDetails({
  fetchAgentInfo, editDetails, editPassword, agent,activeProject
}) {
  const [dataLoading, setDataLoading] = useState(true);

  const [selectedFile, setSelectedFile] = useState(null);
  const [avatarImage, setAvatarImage] = useState(null);

  const handleFileChange = (event) => {
    const file = event.target.files[0];
    
    const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
    const isLt2M = file.size / 1024 / 1024 < 2;

    if (!isJpgOrPng || !isLt2M) {
      return;
    } 
    setSelectedFile(file);
    const reader = new FileReader();
    reader.onload = (e) => {
      udpateProjectDetails(activeProject.id, { profile_picture: e.target.result })
      setAvatarImage(e.target.result);
    };
    reader.readAsDataURL(file);
  };

  const handleEditClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const fileInputRef  = React.createRef();


  useEffect(() => {
    fetchAgentInfo().then(() => {
      setDataLoading(false);
    });
  }, [fetchAgentInfo]);

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Profile</Text>
          </Col>
        </Row>

        <Row className={"mt-2"}>
          <Col>
            {!dataLoading ? (
              <div
                style={{
                  position: "relative",
                  display: "inline-block",
                  overflow: "hidden",
                }}
              >
                <Avatar
                  size={104}
                  style={{
                    color: "#f56a00",
                    backgroundColor: "#fde3cf",
                    fontSize: "42px",
                    textTransform: "uppercase",
                    fontWeight: "400",
                  }}
                  src={avatarImage} 
                >
                  {`${agent?.first_name?.charAt(0)}${agent?.last_name?.charAt(0)}`}
                </Avatar>
                <EditOutlined
                  style={{
                    position: "absolute",
                    bottom: 0,
                    right: 0,
                    backgroundColor: "white",
                    padding: "4px",
                    borderRadius: "50%",
                    cursor: "pointer",
                    zIndex: 1,
                  }}
                  onClick={handleEditClick}
                />
                <input
                  type="file"
                  ref={fileInputRef}
                  style={{ display: "none" }}
                  onChange={handleFileChange}
                />
                <div
                  style={{
                    content: '""',
                    position: "absolute",
                    width: "100%",
                    height: "30px",
                    bottom: "-15px",
                    left: "0",
                    background:
                      "linear-gradient(to bottom, transparent, rgba(0, 0, 0, 0.2))",
                  }}
                ></div>
              </div>
            ) : (
              <Skeleton.Avatar active={true} size={104} shape={"square"} />
            )}
            <Text type={"paragraph"} mini extraClass={"m-0 mt-1"} color={"grey"}>
              A photo helps personalize your account
            </Text>
          </Col>
        </Row>

        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Name</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{`${agent?.first_name} ${agent?.last_name}`}</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Email</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{agent?.email}</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Mobile</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{agent?.phone}</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Password</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>&#8226; &#8226; &#8226; &#8226; &#8226; &#8226;</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col className={'flex justify-start items-center'}>
            <Button size={'large'} disabled={dataLoading} onClick={editDetails}>Edit Details</Button>
            <Button size={'large'} disabled={dataLoading} className={'ml-4'} onClick={editPassword} >Change Password</Button>
          </Col>
        </Row>
      </div>

    </>

  );
}

const mapStatesToProps = (state) => {
  return {
  activeProject: state.global.active_project,
    agent: state.agent.agent_details,
  };
};
export default connect(mapStatesToProps, {udpateProjectDetails, fetchAgentInfo })(ViewUserDetails);
