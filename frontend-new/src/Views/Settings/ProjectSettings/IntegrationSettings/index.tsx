import {
  FaErrorComp,
  FaErrorLog,
  SVG,
  Text
} from 'Components/factorsComponents';
import { Input, Select, Skeleton } from 'antd';
import React, { useContext, useState } from 'react';

import { ErrorBoundary } from 'react-error-boundary';
import useIntegrationCheck from 'hooks/useIntegrationCheck';
import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import styles from './index.module.scss';
import {
  IntegrationPageCategories,
  IntegrationProviderData
} from './integrations.constants';
import { getIntegrationCategoryNameFromId } from './util';
import { IntegrationConfig } from './types';
import { IntegrationContext } from './IntegrationContext';
import IntegrationCard from './IntegrationCard';

const IntegrationSettings = () => {
  const defaultCategory = 'all';
  const categories = [
    { label: 'All Integrations', value: defaultCategory },
    ...IntegrationPageCategories.map((cat) => ({
      label: cat.name,
      value: cat.id
    }))
  ];
  const [selectedCategory, setSelectedCategory] = useState(defaultCategory);
  const [searchText, setSearchText] = useState('');
  const { integrationInfo } = useIntegrationCheck();

  const { integrationStatusLoading } = useContext(IntegrationContext);

  const handleCategoryChange = (value: string) => {
    setSelectedCategory(value);
  };

  const handleSearchTextChange = (e) => {
    setSearchText(e.target.value);
  };

  const renderIntegrationBody = () => {
    let Items = IntegrationProviderData;
    // filtering categories
    if (selectedCategory !== defaultCategory && searchText === '') {
      Items = Items.filter((item) => item.categoryId === selectedCategory);
    }
    if (searchText) {
      Items = Items.filter((item) => {
        const searchSmallCase = searchText.toLowerCase();
        return (
          item.name.toLowerCase().includes(searchSmallCase) ||
          item.desc.toLowerCase().includes(searchSmallCase)
        );
      });
    }
    // categorising data
    const categorizedData: { [key: string]: IntegrationConfig[] } = {};
    Items.forEach((item) => {
      if (categorizedData?.[item.categoryId]) {
        categorizedData[item.categoryId].push(item);
      } else {
        categorizedData[item.categoryId] = [item];
      }
    });

    return IntegrationPageCategories.sort(
      (categoryA, categoryB) => categoryA.sortOrder - categoryB.sortOrder
    ).map((category, index) => {
      const categoryMap = categorizedData[category.id];
      if (!categoryMap || !categoryMap.length) {
        return null;
      }
      return (
        <div className={`${index === 0 ? '' : 'mb-10'}`}>
          <Text
            type='title'
            level={6}
            extraClass='m-0'
            color='character-primary'
            weight='bold'
          >
            {getIntegrationCategoryNameFromId(category.id)}
          </Text>
          <div className='mt-4'>
            {categoryMap.map((c: IntegrationConfig) => (
              <IntegrationCard
                integrationConfig={c}
                integrationInfo={integrationInfo}
              />
            ))}
          </div>
        </div>
      );
    });
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp
          size='medium'
          title='Integrations Error'
          subtitle='We are facing some issues with the integrations. Drop us a message on the in-app chat.'
        />
      }
      onError={FaErrorLog}
    >
      {integrationStatusLoading && (
        <>
          <Skeleton />
          <Skeleton />
          <Skeleton />
          <Skeleton />
        </>
      )}
      {!integrationStatusLoading && (
        <>
          <CommonSettingsHeader
            title='Integrations'
            description='Seamlessly connect Factors with your existing tools to enhance productivity and enable powerful workflows.'
          />

          <div className={styles.integrationHeader}>
            <div className=' flex items-center justify-between w-full'>
              <div>
                <Select
                  style={{ width: 300 }}
                  onChange={handleCategoryChange}
                  options={categories}
                  value={selectedCategory}
                />
              </div>
              <div className='flex items-center justify-between'>
                <Input
                  onChange={handleSearchTextChange}
                  placeholder='Search Integration'
                  style={{ width: '220px', borderRadius: 5 }}
                  prefix={<SVG name='search' size={16} color='grey' />}
                />
              </div>
            </div>
          </div>
          <div className='mb-6'>{renderIntegrationBody()}</div>
        </>
      )}
    </ErrorBoundary>
  );
};

export default IntegrationSettings;
