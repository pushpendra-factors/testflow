import React from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import {
  FaErrorComp,
  FaErrorLog,
  SVG,
  Text
} from 'Components/factorsComponents';
import { Button, Col, Divider, Row } from 'antd';
import { useProductFruitsApi } from 'react-product-fruits';
import { PathUrls } from 'Routes/pathUrls';
import styles from './index.module.scss';
import Card from './Card';

function Checklist(): JSX.Element {
  const productFruitRef = React.createRef();
  const checklistId = 2288;

  useProductFruitsApi(
    (api) => {
      if (productFruitRef) {
        api.checklists.injectToElement(checklistId, productFruitRef.current);
      }
    },
    [productFruitRef]
  );

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp
          size='medium'
          title='Bundle Error'
          subtitle='We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
          className={undefined}
          type={undefined}
        />
      }
      onError={FaErrorLog}
    >
      <div className={`fa-container ${styles.container}`}>
        <Row gutter={[24, 24]} justify='center'>
          <Col span={20}>
            <Row gutter={[24, 24]}>
              <Col span={20}>
                <div className='flex justify-between items-center'>
                  <div className='flex flex-col'>
                    <Row className='flex'>
                      <div
                        className='cursor-pointer mt-3 mr-3'
                        onClick={() =>
                          window.open(PathUrls.ProfileAccounts, '_self')
                        }
                      >
                        <SVG name='ArrowLeft' />
                      </div>
                      <Text
                        type='title'
                        level={3}
                        weight='bold'
                        extraClass='m-0'
                        id='fa-at-text--page-title'
                      >
                        Setup Assist
                      </Text>
                    </Row>
                    <Text
                      type='title'
                      level={6}
                      extraClass='m-0 ml-8'
                      color='grey'
                    >
                      Curious to get started? Here are some resources to help.
                    </Text>
                  </div>
                </div>
              </Col>
            </Row>
            <Divider />
            <Row gutter={[24, 24]} className='sticky h-screen'>
              <Col span={16}>
                <div
                  className={`${styles.productFruits}`}
                  ref={productFruitRef}
                />
                <div>
                  <Text
                    type='title'
                    level={6}
                    extraClass='m-0 mt-6 mb-1'
                    color='grey'
                  >
                    Remove setup assist from the top navigation bar
                  </Text>
                  <Button type='default'>Remove Forever</Button>
                </div>
              </Col>
              <Col span={8}>
                <Row gutter={[24, 24]}>
                  <Col span={20}>
                    <div className='flex justify-between items-center'>
                      <div className='flex flex-col'>
                        <Text
                          type='title'
                          level={4}
                          weight='bold'
                          extraClass='m-0'
                          id='fa-at-text--page-title'
                        >
                          Helpful resources
                        </Text>
                        <Text
                          type='title'
                          level={7}
                          extraClass='m-0'
                          color='grey'
                        >
                          Browse through a collection of help docs, product
                          videos and also playbooks on how our customers
                          leverage Factors.
                        </Text>
                      </div>
                    </div>
                  </Col>
                </Row>
                <Divider />
                <Row>
                  <Col span={20}>
                    <div className='flex justify-between items-center'>
                      <div className='flex flex-col'>
                        <Text
                          type='title'
                          level={6}
                          extraClass='m-0 mb-2'
                          color='grey'
                        >
                          Help docs
                        </Text>
                      </div>
                    </div>
                  </Col>
                </Row>
                <Card
                  bgColor='rgba(250, 250, 250, 1)'
                  title='Integrations'
                  description='Connect Factors to the sales and marketing tools you use everyday.'
                  learnMoreUrl='https://help.factors.ai/en/collections/3954157-integrations'
                  imgUrl='assets/images/checklist/integration.svg'
                  category={1}
                />
                <Card
                  bgColor='rgba(255, 241, 240, 1)'
                  title='How to’s'
                  description='Not sure how to use a certain feature? There’s a help guide for that.'
                  learnMoreUrl='#'
                  imgUrl=''
                  category={1}
                />
                <Divider />
                <Row>
                  <Col span={20}>
                    <div className='flex justify-between items-center'>
                      <div className='flex flex-col'>
                        <Text
                          type='title'
                          level={6}
                          extraClass='m-0 mb-2'
                          color='grey'
                        >
                          Resources
                        </Text>
                      </div>
                    </div>
                  </Col>
                </Row>
                <Card
                  bgColor='rgba(255, 247, 230, 1)'
                  title='Customer Stories'
                  description='Learn how other customers use Factors to increase their pipeline and get better ROI on their spends.'
                  learnMoreUrl='#'
                  imgUrl='assets/images/checklist/customerStories.svg'
                  category={2}
                />
                <Card
                  bgColor='rgba(250, 250, 250, 1)'
                  title='Factors.ai Library'
                  description='Check out a host of videos we have put together to help you get started with Factors.'
                  learnMoreUrl='#'
                  imgUrl='assets/images/checklist/factorsLibrary.svg'
                  category={2}
                />
              </Col>
            </Row>
          </Col>
        </Row>
      </div>
    </ErrorBoundary>
  );
}
export default Checklist;
