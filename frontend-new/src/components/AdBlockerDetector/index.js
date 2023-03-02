import { InfoCircleFilled, InfoCircleOutlined } from '@ant-design/icons';
import { notification } from 'antd';
import React, { useEffect, useState } from 'react';
const showAdBlocker = () => {
  notification.info({
    message: `Hmm, it appears that you're using one or more ad blockers`,
    description:
      'We understand why! But unfortunately, this makes it tougher to build perfect reports.',
    placement: 'topRight',
    style: {
      height: '155px',
      width: '452px',
      display: 'grid',
      placeContent: 'center'
    },
    icon: <InfoCircleFilled width={'24px'} style={{ color: 'red' }} />
  });
};
const AdBlockerDetector = (props) => {
  let [isAdBlocker, setIsAdBlocker] = useState(false);
  useEffect(() => {
    // Detects Crystal Adblocker
    // Method 3
    (function detectAdblockWithInvalidURL(callback) {
      let controller = new AbortController();
      const signal = controller.signal;

      var flaggedURL =
        'https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js';

      if (window.fetch) {
        var request = new Request(flaggedURL, {
          method: 'HEAD',
          mode: 'no-cors',
          signal: signal
        });
        fetch(request)
          .then(function (response) {
            if (response.status === 404) {
              controller.abort();
              callback(false, response);
            }
          })
          .catch(function (error) {
            controller.abort();
            callback(true, error);
          });
      } else {
        var http = new XMLHttpRequest();
        http.open('HEAD', flaggedURL, false);

        try {
          http.send();
        } catch (err) {
          controller.abort();
          callback(true, err);
        }

        if (http.status === 404) {
          controller.abort();
          callback(false, http);
        }
      }
    })(function (usingAdblock, res) {
      // returns if adblocker is there or not

      if (isAdBlocker === false) {
        setIsAdBlocker(isAdBlocker || usingAdblock);
      }
    });
  }, []);

  useEffect(() => {
    if (isAdBlocker) {
      let lastTime = localStorage.getItem('lastAdBlockerTrigger');
      if (!lastTime) {
        showAdBlocker();
        localStorage.setItem(
          'lastAdBlockerTrigger',
          String(new Date().getTime())
        );
      } else {
        let curr = new Date().getTime();
        let diff = curr - Number(lastTime);
        let threshold = 1000 * 60 * 60 * 24;
        if (diff > threshold) {
          showAdBlocker();
          localStorage.setItem(
            'lastAdBlockerTrigger',
            String(new Date().getTime())
          );
        }
      }
    }
  }, [isAdBlocker]);
  return <>{props.children}</>;
};
export default AdBlockerDetector;
