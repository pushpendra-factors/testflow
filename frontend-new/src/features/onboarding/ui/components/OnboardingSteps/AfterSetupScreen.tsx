import { Button, Col, Divider, Row } from 'antd';
import useMobileView from 'hooks/useMobileView';
import React from 'react';
import IllustrationImage from '../../../../../assets/images/onboarding_step5.png';
import { SVG, Text } from 'Components/factorsComponents';
import OnboardingHeader from '../OnboardingHeader';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import style from './index.module.scss';

const Step5 = () => {
  const isMobileView = useMobileView();
  const history = useHistory();
  const renderCard = (
    title: string,
    subTitle: string,
    svgName: string,
    url: string,
    color: string
  ) => {
    return (
      <div
        className={`p-4 gap-6 cursor-pointer flex justify-between items-center ${style.outlineBorder} `}
        onClick={() => history.push(url)}
      >
        <div className='flex gap-4'>
          <div className={`p-2 h-full  ${style.outlineBorder}`}>
            <SVG name={svgName} size='40' color={color} />
          </div>
          <div>
            <Text
              type={'title'}
              color='character-title'
              level={6}
              weight={'bold'}
              extraClass='m-0 mb-1'
            >
              {title}
            </Text>
            <Text
              type={'title'}
              color='character-secondary'
              level={7}
              extraClass='m-0'
            >
              {subTitle}
            </Text>
          </div>
        </div>
        <div>
          <SVG name='ChevronRight' size='16' />
        </div>
      </div>
    );
  };
  return (
    <>
      <OnboardingHeader totalSteps={5} currentStep={5} />
      <div style={{ padding: isMobileView ? '32px 16px' : '60px 222px' }}>
        <Row>
          <Col xs={24} sm={24} md={8}>
            <div className='p-4 flex justify-center items-center'>
              <img
                src={IllustrationImage}
                alt='illustration'
                className='h-full w-full'
                style={{ width: 217, height: 212 }}
              />
            </div>
          </Col>
          <Col xs={24} sm={24} md={16}>
            <Text
              type={'title'}
              level={3}
              color={'character-primary'}
              weight={'bold'}
            >
              Congratulations, Your Project is ready{' '}
              <span role='img' aria-label='congratulations'>
                🎉
              </span>
            </Text>
            <Text
              type={'title'}
              level={6}
              color='character-disabled-placeholder'
            >
              Yeah! Your project setup is now complete. We have started pulling
              data into your project and you can soon expect to see accounts we
              have identified for you.
              <br />
              <br />
              Meanwhile, feel free to set up additional integrations or invite
              your teammates while we set up the product for you liking.
            </Text>
            <Button
              type='primary'
              className={'m-0'}
              onClick={() => history.push(PathUrls.ProfileAccounts)}
            >
              Continue to project
            </Button>
          </Col>
          <Divider />
          <Col span={24} className='mb-5'>
            <Text type={'title'} level={6} color='character-secondary'>
              What do you want to get started with first?
            </Text>
          </Col>
          <Row gutter={[32, 40]}>
            <Col xs={24} sm={24} md={12}>
              {renderCard(
                'Connect with other apps',
                'Sync data from ad platforms like LinkedIn or connect to messaging apps like Slack',
                'ManageDb',
                PathUrls.SettingsIntegration,
                '#597EF7'
              )}
            </Col>
            <Col xs={24} sm={24} md={12}>
              {renderCard(
                'Invite your teammates',
                'Invite people from your team to achieve more using Factors',
                'UserPlusRegular',
                PathUrls.SettingsUser,
                '#13C2C2'
              )}
            </Col>
            <Col xs={24} sm={24} md={12}>
              {renderCard(
                'Create a segment',
                'Identify the set of accounts that matter most to your organisation by creating a custom segment',
                'ChartPie',
                PathUrls.ProfileAccounts,
                '#FA541C'
              )}
            </Col>
            <Col xs={24} sm={24} md={12}>
              {renderCard(
                'Explore Web Analytics',
                'Build KPI, Events or Funnel reports to find answers using data',
                'Analise',
                PathUrls.Dashboard,
                '#52C41A'
              )}
            </Col>
          </Row>
        </Row>
      </div>
    </>
  );
};

export default Step5;
