import React, { useRef } from 'react';
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
import { updateChecklistStatus } from 'Reducers/global';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { bindActionCreators } from 'redux';
import logger from 'Utils/logger';
import { fetchProjectAgents } from 'Reducers/agentActions';
import { meetLink } from 'Utils/meetLink';
import Card from './Card';
import styles from './index.module.scss';

function Checklist({
  updateChecklistStatus,
  fetchProjectAgents
}: ChecklistComponentProps): JSX.Element {
  const { active_project } = useSelector((state) => state.global);
  const history = useHistory();
  const productFruitRef = useRef<HTMLDivElement>(null);
  const checklistId = 2288;

  useProductFruitsApi(
    (api) => {
      if (productFruitRef) {
        api.checklists.injectToElement(checklistId, productFruitRef.current);
      }
    },
    [productFruitRef]
  );

  const handleRemoveForever = () => {
    updateChecklistStatus(active_project?.id, true)
      .then((res: any) => {
        fetchProjectAgents(active_project?.id);
        history.push(PathUrls.Dashboard);
      })
      .catch((err: any) => {
        logger.error(err);
      });
  };

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
                      Get started with these recommended actions or go through
                      some helpful resources to unlock the full value of
                      Factors. If you ever need help, feel free to book a free
                      consultation call on how to utilise Factors data with your
                      team.
                      <a
                        href={meetLink()}
                        target='_blank'
                        className='ml-1'
                        rel='noreferrer'
                      >
                        Book a call
                      </a>
                    </Text>
                  </div>
                </div>
              </Col>
            </Row>
            <Divider />
            <Row gutter={[24, 24]} className='sticky pb-2'>
              <Col>
                <div
                  className={`${styles.productFruits}`}
                  ref={productFruitRef}
                />
              </Col>
            </Row>
            <Divider />
            <Row>
              <Col>
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
                          A collection of help docs, product videos and
                          playbooks on how customers leverage Factors.
                        </Text>
                      </div>
                    </div>
                  </Col>
                </Row>
                <Row className='flex justify-between items-center mt-4'>
                  <Col span={12} className='pr-6'>
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
                    <div className='mb-8'>
                      <Card
                        title='Integrations'
                        description='Have you brought in all the data you care about? Explore data integrations that Factors has to offer.'
                        learnMoreUrl='https://www.factors.ai/integrations'
                        imgUrl='assets/images/checklist/integration.svg'
                      />
                    </div>
                    <div>
                      <Card
                        title='Help guides'
                        description='Need help on how to use a certain feature? Check out our host of help guides crafted to get you started quickly.'
                        learnMoreUrl='https://help.factors.ai/en'
                        imgUrl='assets/images/checklist/helpGuides.svg'
                      />
                    </div>
                  </Col>
                  <Col span={12} className='pl-6'>
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
                    <div className='mb-8'>
                      <Card
                        title='Customer Stories'
                        description='Learn how other customers use Factors to increase their pipeline, get better ROI on their spends and close more revenue.'
                        learnMoreUrl='https://www.factors.ai/customers'
                        imgUrl='assets/images/checklist/customerStories.svg'
                      />
                    </div>
                    <div>
                      <Card
                        title='Video Library'
                        description='Check out a host of videos like product walkthroughs, feature videos, webinars, podcasts and much much more.'
                        learnMoreUrl='https://www.youtube.com/@factors-ai'
                        imgUrl='assets/images/checklist/videoLibrary.svg'
                      />
                    </div>
                  </Col>
                </Row>
                <Divider />
                <Row>
                  <div>
                    <Text
                      type='title'
                      level={6}
                      extraClass='m-0 mb-1'
                      color='grey'
                    >
                      Want to remove{' '}
                      <span className='italic'>Finish Setup</span> button from
                      the top bar? You can still access this screen using the
                      project menu.
                    </Text>
                    <Button
                      type='default'
                      onClick={() => handleRemoveForever()}
                    >
                      Remove Forever
                    </Button>
                  </div>
                </Row>
              </Col>
            </Row>
          </Col>
        </Row>
      </div>
    </ErrorBoundary>
  );
}

const mapDispatchToProps = (dispatch: any) =>
  bindActionCreators(
    {
      updateChecklistStatus,
      fetchProjectAgents
    },
    dispatch
  );
const connector = connect(null, mapDispatchToProps);
type ChecklistComponentProps = ConnectedProps<typeof connector>;

export default connector(Checklist);
