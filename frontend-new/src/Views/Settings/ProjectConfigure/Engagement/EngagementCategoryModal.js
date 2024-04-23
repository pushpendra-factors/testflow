import { Text } from 'Components/factorsComponents';
import {
  Button,
  InputNumber,
  Modal,
  Skeleton,
  Tooltip,
  message,
  notification
} from 'antd';
import React, { useEffect, useState } from 'react';
import { InfoCircleFilled } from '@ant-design/icons';
import { EngagementTag } from 'Components/Profile/constants.ts';
import {
  getEngagementCategoryRanges,
  updateEngagementCategoryRanges
} from 'Reducers/timelines';
import { useSelector } from 'react-redux';
import _ from 'lodash';
import styles from './index.module.scss';

function EngagementPill({ type }) {
  return (
    <div
      className={`${styles['category-pill']} flex items-center`}
      style={{ backgroundColor: EngagementTag[type]?.bgColor }}
    >
      <img
        src={`../../../assets/icons/${EngagementTag[type]?.icon}.svg`}
        alt=''
      />

      <Text extraClass='m-0' type='title' level={7}>
        {type}
      </Text>
    </div>
  );
}
const EngagementFormOrder = ['Hot', 'Warm', 'Cool', 'Ice'];
function EngagementCategoryModal({
  visible,
  onOk,
  onCancel,
  getRanges,
  ...props
}) {
  const [status, setStatus] = useState({
    isFormLoading: false,
    isFormSubmitting: false
  });
  const [savedRanges, setSavedRanges] = useState([]);
  const activeProject = useSelector((state) => state.global.active_project);
  const [categoryRange, setCategoryRange] = useState([
    [90, 100],
    [70, 90],
    [50, 70],
    [0, 50]
  ]);
  const onInputChange = (value, indexes) => {
    const x = indexes[0];
    const y = indexes[1];

    setCategoryRange((prev) => {
      const arr = [];

      for (const e of prev) {
        arr.push([...e]);
      }
      if (x === 0 && y === 0) {
        arr[0][0] = value < 3 ? 3 : value > 99 ? 99 : value; // HOT - Low
        arr[1][1] = arr[0][0]; // Warm - High
        arr[1][0] = Math.min(arr[1][0], arr[1][1] - 1); // Warm - Low
        arr[2][1] = Math.min(arr[2][1], arr[1][0]); // Cool - High
        arr[2][0] = Math.min(arr[2][0], arr[2][1] - 1); // Cool - Low
        arr[3][1] = Math.min(arr[3][1], arr[2][0]); // Ice - High
      } else if (x === 1 && y === 0) {
        arr[1][0] = value < 2 ? 2 : value >= arr[1][1] ? arr[1][1] - 1 : value; // Warm - Low
        arr[2][1] = arr[1][0]; // Cool - High
        arr[2][0] = Math.min(arr[2][0], arr[2][1] - 1); // Cool - Low
        arr[3][1] = Math.min(arr[3][1], arr[2][0]); // Ice - High
      } else if (x === 2 && y === 0) {
        arr[2][0] = value < 1 ? 1 : value >= arr[2][1] ? arr[2][1] - 1 : value; // Cool - Low
        arr[3][1] = arr[2][0]; // Ice - High
      }
      return arr;
    });
  };
  const handleResetButton = () => {
    const tmp = [
      [90, 100],
      [70, 90],
      [30, 70],
      [0, 30]
    ];
    setCategoryRange(tmp);
  };
  const fetchCategories = () => {
    setStatus((prev) => ({ ...prev, isFormLoading: true }));
    getEngagementCategoryRanges(activeProject.id)
      .then((data) => {
        try {
          const tmpRng = data?.data?.bck;
          const HotRng = tmpRng.find((e) => e.nm === 'Hot');
          const WarmRng = tmpRng.find((e) => e.nm === 'Warm');
          const CoolRng = tmpRng.find((e) => e.nm === 'Cool');
          const IceRng = tmpRng.find((e) => e.nm === 'Ice');

          const tmpCategory = [];
          tmpCategory.push([HotRng.low, HotRng.high]);
          tmpCategory.push([WarmRng.low, WarmRng.high]);
          tmpCategory.push([CoolRng.low, CoolRng.high]);
          tmpCategory.push([IceRng.low, IceRng.high]);
          setCategoryRange(tmpCategory);
          setSavedRanges(tmpCategory);
        } catch (err) {
          handleResetButton();
          // message.error('Failed to Load Category Ranges, Reset to Default');
        }
      })
      .catch((err) => {
        handleResetButton();
        // eslint-disable-next-line no-console
        console.error(err);
        message.error('Failed to Load Category Ranges');
      })
      .finally(() => {
        setStatus((prev) => ({ ...prev, isFormLoading: false }));
      });
  };

  const handleApply = () => {
    const payload = {
      date: String(new Date().getTime() / 1000),
      bck: categoryRange.map((eachCategoryRange, eachIndex) => ({
        nm:
          eachIndex === 0
            ? 'Hot'
            : eachIndex === 1
              ? 'Warm'
              : eachIndex === 2
                ? 'Cool'
                : eachIndex === 3
                  ? 'Ice'
                  : '',
        high: eachCategoryRange[1],
        low: eachCategoryRange[0]
      }))
    };
    setStatus((prev) => ({ ...prev, isFormSubmitting: true }));
    updateEngagementCategoryRanges(activeProject.id, payload)
      .then(() => {
        notification.success({
          message: 'Engagement category rules updated.',
          description:
            'All accounts will be re-assigned categories based on new rules within 24 hours.'
        });
      })
      .catch((err) => {
        console.error(err);
        message.error('Error Updating Category Ranges');
      })
      .finally(() => {
        if (onOk) onOk();
        // Finally
        setStatus((prev) => ({ ...prev, isFormSubmitting: false }));
      });
  };
  useEffect(() => {
    fetchCategories();
  }, []);
  return (
    <Modal
      title={null}
      width={574}
      visible={visible}
      className='fa-modal--regular p-6'
      onCancel={onCancel}
      centered
      {...props}
      footer={
        <div
          className='inline-flex justify-between pb-4 pl-4 pr-4'
          style={{ width: '100%' }}
        >
          <div>
            <Button type='text' onClick={handleResetButton}>
              Reset to Default
            </Button>
          </div>
          <div className='inline-flex justify-between' style={{ gap: '10px' }}>
            <Button type='text' className='dropdown-btn' onClick={onCancel}>
              Cancel
            </Button>
            <Button
              loading={status.isFormSubmitting}
              type='primary'
              onClick={handleApply}
              disabled={
                _.isEqual(savedRanges, categoryRange) ||
                savedRanges.length === 0
              }
            >
              Apply Changes
            </Button>
          </div>
        </div>
      }
    >
      <div className='p-2'>
        <div className='pb-4'>
          <Text extraClass='m-0' type='title' level={4} weight='bold'>
            Engagement Category
          </Text>
          <Text extraClass='m-0' type='title' level={7} color='grey'>
            Fine tune how to grade the engagement of accounts based on their
            engagement score
          </Text>
        </div>
        <div>
          <table className={`${styles.engagement_category_fill} p-2`}>
            <thead>
              <td>Category</td>
              <td>From</td>
              <td>To</td>
              <td />
            </thead>
            <tbody>
              {EngagementFormOrder.map((eachType, eachIndex) => (
                <tr key={`row-${eachType}`}>
                  <td>
                    <EngagementPill type={eachType} />
                  </td>
                  <td>
                    <div
                      className={`${styles['category-pill']} ${styles['input-cells']}`}
                    >
                      {status.isFormLoading ? (
                        <Skeleton.Input
                          active
                          size='default'
                          block
                          style={{ width: 48 }}
                        />
                      ) : (
                        <InputNumber
                          disabled={eachIndex === 3 || status.isFormLoading}
                          controls={false}
                          //   min={categoryRange[1][1] + 1}
                          value={categoryRange[eachIndex][0]}
                          onKeyDown={(e) => {
                            if (e.key === 'ArrowDown') {
                              const value = Number(e.currentTarget.value);
                              onInputChange(value - 1, [eachIndex, 0]);
                            }
                            if (e.key === 'ArrowUp') {
                              const value = Number(e.currentTarget.value);
                              onInputChange(value + 1, [eachIndex, 0]);
                            }
                            if (e.key === 'Enter') {
                              const value = Number(e.currentTarget.value);
                              onInputChange(value, [eachIndex, 0]);
                            }
                          }}
                          onBlur={(e) => {
                            const value = Number(e.currentTarget.value);
                            onInputChange(value, [eachIndex, 0]);
                          }}
                          style={{ width: '50px' }}
                        />
                      )}
                      <span>%</span>
                    </div>
                  </td>
                  <td>
                    <div className={styles['category-pill']}>
                      {status.isFormLoading ? (
                        <Skeleton.Input
                          active
                          size='default'
                          block
                          style={{ width: 48 }}
                        />
                      ) : (
                        <InputNumber
                          controls={false}
                          value={categoryRange[eachIndex][1]}
                          onChange={(e) => onInputChange(e, [eachIndex, 1])}
                          disabled
                          style={{ width: '50px' }}
                        />
                      )}
                      <span>%</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div id='engagement-modal-tooltip' style={{ width: 'fit-content' }}>
            <Tooltip
              title='Categories are assigned based on relative scores of accounts.'
              getTooltipContainer={() =>
                document.querySelector('#engagement-modal-tooltip')
              }
            >
              <div className='flex items-center gap-2 mx-auto my-0'>
                {' '}
                <InfoCircleFilled /> <div>Percentile Based</div>
              </div>
            </Tooltip>
          </div>
        </div>
      </div>
    </Modal>
  );
}
export default EngagementCategoryModal;
