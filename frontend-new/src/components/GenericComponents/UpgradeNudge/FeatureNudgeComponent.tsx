import React, { useState } from 'react';
import { Button, Modal } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import usePlanUpgrade from 'hooks/usePlanUpgrade';
import { PhoneFilled, PlayCircleOutlined } from '@ant-design/icons';
import YouTube from 'react-youtube';
import { PRICING_HELP_LINK } from 'Views/Settings/ProjectSettings/Pricing/utils';
import { UpgradeNudgeConfig } from './config';
import styles from './index.module.scss';
// import YouTubePlayer from '../YoutubePlayer';

const FeatureNudgeComponent = ({ config }: FeatureNudgeComponentProps) => {
  const {
    image,
    backgroundColor,
    description,
    title,
    featureName,
    videoId,
    name
  } = config;
  const [showVideoModal, setShowVideoModal] = useState(false);
  const { handlePlanUpgradeClick } = usePlanUpgrade();
  const handlePlayClick = () => {
    setShowVideoModal(true);
  };
  const youtubeOpts = {
    height: '405',
    width: '100%',
    playerVars: {
      autoplay: 1,
      rel: 0
    }
  };

  return (
    <>
      <div
        className='px-6 py-4 flex justify-between items-center'
        style={{ backgroundColor, borderRadius: 12, height: 90 }}
      >
        <div className='flex gap-3 items-center' style={{ width: '80%' }}>
          <img
            src={image}
            alt='illustration'
            style={{ height: 60, width: 106 }}
          />
          <div className='flex flex-col justify-center'>
            <Text
              type='title'
              level={6}
              weight='bold'
              color='white'
              extraClass='m-0'
            >
              {title}
            </Text>
            <Text type='title' level={7} extraClass='m-0' color='white'>
              {description}
            </Text>
          </div>
        </div>
        <div className='flex items-center gap-2 mr-2'>
          {videoId && (
            <div>
              <Button
                className={styles['play-video-button']}
                onClick={handlePlayClick}
                type='text'
                icon={<PlayCircleOutlined />}
              >
                Learn
              </Button>
            </div>
          )}
          <Button
            icon={<SVG name='ArrowBottomUp' color='#595959' />}
            onClick={() => handlePlanUpgradeClick(featureName)}
          >
            Upgrade
          </Button>
        </div>
      </div>
      {videoId && showVideoModal && (
        <Modal
          className={styles['upgrade-modal']}
          visible={showVideoModal}
          onCancel={() => setShowVideoModal(false)}
          width={800}
          footer={null}
        >
          <div>
            <div className={styles['background-images']}>
              <div className={styles['background-image-top-left-account']} />
              <div
                className={styles['background-image-bottom-right-account']}
              />
            </div>
            <Text
              type='title'
              level={3}
              weight='bold'
              color='mono-5'
              extraClass='m-0 relative z-50'
            >
              Upgrade to access this feature
            </Text>

            <Text
              type='title'
              level={6}
              color='character-secondary'
              extraClass='m-0 mt-1'
            >
              Looks like your current plan doesn't include {name} 😢. Upgrade
              now or talk to your product admin if you wish to use this feature.
            </Text>

            <div className='mt-6 z-50 relative'>
              <YouTube
                iframeClassName={styles.player}
                opts={youtubeOpts}
                videoId={videoId}
              />
            </div>
            <div className='mt-6 flex gap-2 justify-end'>
              <a href={PRICING_HELP_LINK} target='_blank' rel='noreferrer'>
                <Button
                  icon={<PhoneFilled style={{ transform: 'rotate(90deg)' }} />}
                >
                  Get a Demo
                </Button>
              </a>
              <Button
                type='primary'
                icon={<SVG name='ArrowBottomUp' color='#fff' />}
                onClick={() => handlePlanUpgradeClick(featureName)}
              >
                Upgrade
              </Button>
            </div>
          </div>
        </Modal>
      )}
    </>
  );
};

interface FeatureNudgeComponentProps {
  config: UpgradeNudgeConfig;
}

export default FeatureNudgeComponent;
