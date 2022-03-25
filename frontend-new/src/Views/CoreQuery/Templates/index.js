import React, { useState } from 'react';
import {
  Row, Col, Modal, Button, Tag
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { fetchTemplateConfig } from 'Reducers/templates';
import factorsai from 'factorsai';

const TemplatesModal = ({
  templatesModalVisible,
  setTemplatesModalVisible,
  fetchTemplateConfig,
  activeProject
}) => {

  const history = useHistory();
  const routeChange = (url) => {
    history.push(url);
  };

  const ChooseTemplate = (templateID) => { 

    //Factors RUN_QUERY tracking
    factorsai.track('RUN-QUERY',{'query_type': 'template', 'templateID': templateID});

    fetchTemplateConfig(activeProject.id, templateID).then(() => {
      routeChange('/templates');
    }).catch((e) => console.log("fetch template config error", e));
  }


  const templatesList = [
    {
      name: 'Google Search Ads Anomaly',
      desc: 'Know what metrics have changed substantially for your Search Campaigns and Keywords and the reason behind the change in conversions.',
      img: 'https://s3.amazonaws.com/www.factors.ai/assets/img/product/template-thumbnail-1.png',
      tag: 'Paid Marketing',
      active: true,
    },
    {
      name: 'SEO Anomaly',
      desc: 'Know what metrics have changed substantially for your organic search terms and landing pages and the reason behind the change in conversions.',
      img: 'https://s3.amazonaws.com/www.factors.ai/assets/img/product/template-thumbnail-2.png',
      tag: 'SEO & Organic Marketing',
      active: false,
    },
    {
      name: 'B2B SaaS Marketing Planning',
      desc: 'Plan the leads, opportunities and pipeline in accordance to revenue targets.',
      img: 'https://s3.amazonaws.com/www.factors.ai/assets/img/product/template-thumbnail-3.png',
      tag: 'Marketing Planning',
      active: false,
    },
    {
      name: 'Marketing Sourced vs Marketing Influenced',
      desc: 'Know how many opportunities, deals, pipeline and revenue was marketing sourced vs marketing influenced.',
      img: 'https://s3.amazonaws.com/www.factors.ai/assets/img/product/template-thumbnail-4.png',
      tag: 'Marketing Effectiveness',
      active: false,
    },

  ]

  return (
    <>

      <Modal
        title={null}
        visible={templatesModalVisible}
        footer={null}
        centered={false}
        // zIndex={1005}
        mask={false}
        closable={false}
        className={'fa-modal--full-width'}
      >

        <div className={'fa-modal--header'}>
          <div className={'fa-container'}>
            <Row justify={'space-between'} className={'py-4 m-0 '}>
              <Col>
                <SVG name={'brand'} size={40} />
              </Col>
              <Col>
                <Button size={'large'} type="text" onClick={() => setTemplatesModalVisible(false)}><SVG name="times"></SVG></Button>
              </Col>
            </Row>
          </div>
        </div>

        <div className={'fa-container'}>
          <Row gutter={[24, 24]} justify={'center'}>
            <Col span={8}>
              <div className={'flex flex-col items-center mt-10 mb-10 mb-10'}>
                <img src="https://s3.amazonaws.com/www.factors.ai/assets/img/product/templates-bg.png" className={'mb-2'} style={{ maxHeight: '75px' }} />
                <Text type={'title'} align={'center'} level={4} weight={'bold'} extraClass={'m-0'}>Start with Quick Templates</Text>
                <Text type={'title'} align={'center'} level={7} color={'grey'} extraClass={'m-0'}>Browse the templates from our wide range of commonly used questions. Curated from top marketers in the industry.</Text>
              </div>
            </Col>
          </Row>
          <Row gutter={[24, 24]} justify={'center'}>
            <Col span={10}>
              {templatesList.map((item, index) => {
                return (
                  <div className={`relative  flex p-4 items-center justify-start border-radius--sm border--thin-2 ${item.active ? 'cursor-pointer': 'fa-template--card cursor-not-allowed'} mb-6`} onClick={item.active ? () => ChooseTemplate(index+1) : null}>
                    {!item.active && <Tag color='red' className={'fai--custom-card--badge'} > Coming Soon </Tag> }
                    <img src={item.img} className={'mb-2'} style={{ maxHeight: '120px' }} />
                    <div className={'flex flex-col items-start ml-4'}>
                      <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>{item.name}</Text>
                      <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 mb-2'}>{item.desc}</Text>
                      <Tag style={{fontSize: '10px'}}>{item.tag}</Tag>
                    </div>
                  </div>
                )
              })}
            </Col>
          </Row>
        </div>

      </Modal>

    </>

  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});


export default connect(mapStateToProps, { fetchTemplateConfig })(TemplatesModal);
