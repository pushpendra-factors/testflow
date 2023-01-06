import React, { useState } from 'react';
import { Button, Tag } from 'antd';
import {
  Text,
  SVG,
  FaErrorComp,
  FaErrorLog
} from '../../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import TemplateCard from '../TemplateCard';
import CopyDashboardModal from './CopyDashboardModal';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { find, isNull } from 'lodash';

function TemplateDetails({
  setShowCardDetails,
  templates,
  setTemplateSelected,
  setShowTemplates
}) {
  const [showCopyDashBoardModal, setShowCopyDashBoardModal] = useState(false);
  const { activeTemplate } = useSelector((state) => state.dashboardTemplates);
  const renderTags = () => {
    let tags = [];
    if (activeTemplate && !isNull(activeTemplate.tags))
      tags = Object.values(activeTemplate.tags);
    return (
      <div className='flex flex-row mt-16'>
        <Text
          type={'title'}
          level={6}
          weight={'bold'}
          color={'grey-2'}
          extraClass={'m-2'}
        >
          Tags:
        </Text>
        {tags.map((tag) => {
          return <Tag className='p-12'>{tag}</Tag>;
        })}
      </div>
    );
  };
  const renderReports = () => {
    let units = [];
    if (activeTemplate && !isNull(activeTemplate.units)) {
      units = Object.values(activeTemplate.units);
    }
    return (
      <div className='flex flex-col mt-6'>
        <Text
          type={'title'}
          level={6}
          weight={'bold'}
          color={'grey-2'}
          extraClass={'m-2'}
        >
          Included Reports
        </Text>
        {units.map((unit, id) => {
          return (
            <Text
              type={'title'}
              level={7}
              color={'grey'}
              weight={'bold'}
              extraClass={'m-2'}
            >
              {id + 1}. {unit.title}
            </Text>
          );
        })}
      </div>
    );
  };
  const rendersimilarDashboards = () => {
    let similarDashboardIds = [];
    if (activeTemplate && !isNull(activeTemplate.similar_template_ids)) {
      similarDashboardIds = Object.values(activeTemplate.similar_template_ids);
    }
    const similarDashboards = similarDashboardIds.map((d) => {
      return templates.data.find((template) => template.id === d);
    });
    return (
      <div>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          color={'grey-2'}
          extraClass={'mt-8 mb-4'}
        >
          Similar Dashboards
        </Text>
        <div className='justify-evenly grid grid-cols-3 gap-4'>
          {similarDashboards.map((id) => {
            return (
              <TemplateCard
                id={id}
                title={templates.data[id - 1].title}
                description={templates.data[id - 1].description}
                setTemplateSelected={setTemplateSelected}
                setShowCardDetails={setShowCardDetails}
              />
            );
          })}
        </div>
      </div>
    );
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Analyse LP Error'}
            subtitle={
              'We are facing trouble loading Analyse landing page. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        <div className='flex flex-col'>
          <Button
            size={'large'}
            type='text'
            icon={<SVG size={20} name={'times'} />}
            onClick={() => {
              setShowCardDetails(false);
              setShowTemplates(false);
              setTemplateSelected(-1);
            }}
            className={styles.close}
          />
          <Button
            size={'large'}
            type='text'
            icon={<SVG size={20} name={'arrowLeft'} />}
            onClick={() => {
              setShowCardDetails(false);
              setTemplateSelected(-1);
            }}
            className={styles.arrow}
          />
          <div className='flex flex-row'>
            <div className='flex flex-col w-3/5 mr-2 ml-24 mt-24 p-8'>
              <div className='flex flex-row'>
                <img
                  alt='template'
                  src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/template-icon-1.png'
                  width='100%'
                  className=''
                />
              </div>
              <div className='flex flex-row'>{rendersimilarDashboards()}</div>
            </div>
            <div className='flex flex-col w-2/5 justify-start mr-12 mt-24 p-8'>
              <div className='flex flex-col'>
                <Text
                  type={'title'}
                  level={4}
                  weight={'bold'}
                  extraClass={'mx-2 mb-2'}
                >
                  {activeTemplate.title}
                </Text>
                <Text
                  type={'paragraph'}
                  level={7}
                  color={'grey'}
                  weight={'bold'}
                  extraClass={'m-2'}
                >
                  {activeTemplate.description}
                </Text>
                <Button
                  type='primary'
                  size={'large'}
                  icon={<SVG name='copy' size={16} color={'white'} />}
                  className='m-2 w-3/5'
                  onClick={() => setShowCopyDashBoardModal(true)}
                >
                  Copy this dashboard
                </Button>
                {renderTags()}
                {renderReports()}
              </div>
            </div>
          </div>
        </div>
        <CopyDashboardModal
          showCopyDashBoardModal={showCopyDashBoardModal}
          setShowCopyDashBoardModal={setShowCopyDashBoardModal}
          setShowTemplates={setShowTemplates}
        />
      </ErrorBoundary>
    </>
  );
}
export default TemplateDetails;
