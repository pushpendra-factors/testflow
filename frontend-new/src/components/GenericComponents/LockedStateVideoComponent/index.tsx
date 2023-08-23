import { Text } from 'Components/factorsComponents';
import { Button } from 'antd';
import React from 'react';
import { Link, useHistory } from 'react-router-dom';
import style from './index.module.scss';
import YouTubePlayer from '../YoutubePlayer';
import { PathUrls } from 'Routes/pathUrls';

const LockedStateComponent = ({
  title,
  description,
  learnMoreLink,
  upgradeText = 'To use this feature, you should upgrade from your current plan to Professional',
  upgradeButtonText = 'Upgrade plan',
  upgradeLink = PathUrls.SettingsPricing,
  embeddedLink = 'https://global-uploads.webflow.com/5f28f6242b5cee6e96d76336/649505b5b4c5c322af5ec115_RA%20In%20Feture%202.webp'
}: LockedStateComponentProps) => {
  const history = useHistory();
  return (
    <div className={style.container}>
      <div
        className='flex gap-10 items-center justify-start'
        style={{ height: 420 }}
      >
        {/* iframe video */}
        <div className='w-1/2 h-full' style={{ borderRadius: 15 }}>
          {/* Todo: uncommnet below player once videos are available */}
          {/* <YouTubePlayer
            embeddedLink={embeddedLink}
            title={title}
            extraClass={style.videoPlayer}
          /> */}
          <img src={embeddedLink} alt='feature' />
        </div>
        {/* description */}
        <div className='w-1/2 h-full flex items-center'>
          <div className='w-full'>
            <Text type={'title'} level={3} weight={'bold'}>
              {title}
            </Text>
            <div className='flex items-center flex-wrap gap-1 mt-1'>
              <Text type={'paragraph'} mini extraClass={'m-0'} color='grey'>
                {description}
              </Text>
              {learnMoreLink && (
                <Link
                  className='flex items-center font-semibold gap-2'
                  style={{ color: `#1d89ff` }}
                  target='_blank'
                  to={{
                    pathname: learnMoreLink
                  }}
                >
                  <Text
                    type={'paragraph'}
                    level={7}
                    weight={'bold'}
                    color='brand-color-6'
                  >
                    Learn more
                  </Text>
                </Link>
              )}
            </div>
            <Text type={'paragraph'} mini color='grey' extraClass={'m-0 mt-2'}>
              {upgradeText}
            </Text>

            <div className={style.upgradeButton}>
              <Button type='primary' onClick={() => history.push(upgradeLink)}>
                {upgradeButtonText}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

type LockedStateComponentProps = {
  title: string;
  description: string;
  learnMoreLink?: string;
  upgradeText?: string;
  upgradeButtonText?: string;
  upgradeLink?: string;
  embeddedLink: string;
};

export default LockedStateComponent;
