import React from 'react';
import { Row, Col, Button } from 'antd';
import SavedGoals from './savedList';
import { Text } from 'factorsComponents';
import { useHistory } from 'react-router-dom';

const PathAnalysisLP = ({ SetfetchingIngishts, setShowReport }) => {
  const history = useHistory();
  const createNewPathQuery = () => {
    history.push('/path-analysis/insights');
  };
  return (
    <>
      <div className={'fa-container'}>
        <Row gutter={[24, 24]} justify='center'>
          <Col span={20}>
            <Row gutter={[24, 24]}>
              <Col span={20}>
                <div className='flex justify-between items-center'>
                  <div className='flex flex-col'>
                    <Text
                      type={'title'}
                      level={3}
                      weight={'bold'}
                      extraClass={'m-0'}
                      id={'fa-at-text--page-title'}
                    >
                      Path Analysis
                    </Text> 
                    <Text
                      type={'title'}
                      level={6}
                      extraClass={'m-0 mt-4 mr-2'}
                      color={'grey'}
                    > 

                      Uncover the most effective paths that lead to conversions, helping you maximize customer engagement and drive business growth.&nbsp;
                      <a href='https://help.factors.ai/en/articles/7302103-path-analysis' target='_blank'>
                        Learn more
                      </a>
                    </Text>
                  </div>
                  
                </div>
              </Col>
              <Col span={4}>
              <div className='flex justify-end items-center'>
              <Button
                    type='primary'
                    size='large'
                    onClick={() => {
                      createNewPathQuery();
                    }}
                  > 
                    {`Create New`}
                  </Button>

              </div>
              </Col>
            </Row>
            <Row gutter={[24, 24]}>
              <Col span={24}>
                <SavedGoals
                  SetfetchingIngishts={SetfetchingIngishts}
                  setShowReport={setShowReport}
                />
              </Col>
            </Row>
          </Col>
        </Row>
      </div>
    </>
  );
};

export default PathAnalysisLP;
