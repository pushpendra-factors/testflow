import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Progress, Button, Upload, message
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';


function Brand() {
    const [imageUrl, setImageUrl] = useState('');
    const [loading, setLoading] = useState(false);
    const history = useHistory();

    const handleCreate = (e) => {
        e.preventDefault();
        history.push('/project-setup');
    }

    function getBase64(img, callback) {
        const reader = new FileReader();
        reader.addEventListener('load', () => callback(reader.result));
        reader.readAsDataURL(img);
      }
      
      function beforeUpload(file) {
        const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
        if (!isJpgOrPng) {
          message.error('You can only upload JPG/PNG file!');
        }
        const isLt2M = file.size / 1024 / 1024 < 2;
        if (!isLt2M) {
          message.error('Image must smaller than 2MB!');
        }
        return isJpgOrPng && isLt2M;
      }

      const handleChange = info => {
        if (info.file.status === 'uploading') {
          setLoading(true);
          return;
        }
        if (info.file.status === 'done') {
          // Get this url from response in real world.
          getBase64(info.file.originFileObj, imageUrl => {
            setImageUrl(imageUrl);
            setLoading(false);
          });
        }
      };

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
                                <Upload
                                    name="avatar"
                                    accept={''}
                                    showUploadList={false}
                                    action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
                                    beforeUpload={beforeUpload}
                                    onChange={handleChange}
                                >
                                    {imageUrl ? <img src={imageUrl} alt="avatar" style={{width:'105px',marginLeft:'150px'}} /> : <img src='../../../../assets/avatar/ModalAvatar.png' style={{marginLeft:'150px'}}></img>}
                                </Upload>
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
