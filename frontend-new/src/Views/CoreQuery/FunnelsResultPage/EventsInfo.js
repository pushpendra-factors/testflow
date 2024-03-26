import React from 'react';
import styles from './index.module.scss';
import { QUERY_TYPE_EVENT } from '../../../utils/constants';

function EventsInfo({ queries, setDrawerVisible, queryType }) {
  const charArr = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];

  if (queryType === QUERY_TYPE_EVENT) {
    return (
            <div onClick={setDrawerVisible.bind(this, true)} className={`whitespace-nowrap pb-1 flex items-center cursor-pointer leading-6 overflow-hidden ${styles.eventsText}`}>
                {queries.map((q, index) => {
                  if (index < queries.length - 1) {
                    return (
                            <div className="flex items-center" key={index}>
                                <div style={{ backgroundColor: '#3E516C' }} className="text-white w-6 h-6 flex justify-center items-center mr-1 rounded-full font-semibold leading-5 text-xs">{charArr[index]}</div>
                                <span style={{ color: '#0E2647' }} className="text-xl font-semibold">{q}</span>
                                <span style={{ color: '#8692A3' }} className="text-xl font-normal">&nbsp;and&nbsp;</span>
                            </div>
                    );
                  } else {
                    return (
                            <div className="flex items-center" key={index}>
                                <div style={{ backgroundColor: '#3E516C' }} className="text-white w-6 h-6 flex justify-center items-center mr-1 rounded-full font-semibold leading-5 text-xs">{charArr[index]}</div>
                                <span style={{ color: '#0E2647' }} className="text-xl font-semibold">{q}</span>
                            </div>
                    );
                  }
                })}
            </div>
    );
  }

  return (
        <div className="flex justify-between items-center">
            <div className="flex items-center leading-6">
                <span className="mr-2">
                    <svg width="20" height="24" viewBox="0 0 20 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M5.10893 5.52488L6.46842 7.43263C4.17018 7.77059 2.73045 8.38528 2.42888 8.79916C2.8642 9.39671 5.66343 10.4106 9.99237 10.413C14.3213 10.4155 17.123 9.39671 17.5559 8.79916C17.2543 8.38528 15.8146 7.77059 13.5188 7.43263L14.8782 5.52488C17.444 6.02937 19.6839 7.01876 19.6839 8.79916C19.6841 9.11929 19.6064 9.43461 19.4577 9.71752C19.3098 9.95527 19.1474 10.1835 18.9713 10.4008L12.3174 18.4824C11.8655 19.0273 11.6176 19.7145 11.6169 20.4244V23.2334C11.6169 23.3349 11.5969 23.4354 11.5581 23.529C11.5192 23.6227 11.4623 23.7076 11.3906 23.7789C11.3188 23.8502 11.2338 23.9065 11.1403 23.9444C11.0468 23.9824 10.9468 24.0012 10.846 23.9999H9.14361C9.04364 23.9999 8.94466 23.9801 8.8523 23.9416C8.75995 23.9031 8.67603 23.8466 8.60535 23.7754C8.53466 23.7042 8.47859 23.6197 8.44034 23.5267C8.40208 23.4337 8.38239 23.3341 8.38239 23.2334V20.4342C8.38172 19.7243 8.13387 19.0371 7.68198 18.4922L1.02805 10.4106C0.851937 10.1933 0.689488 9.96506 0.541654 9.72732C0.392884 9.4444 0.315228 9.12909 0.315479 8.80896C0.303319 7.01631 2.54318 6.02692 5.10893 5.52488ZM6.50976 3.42856C6.43177 3.58319 6.39776 3.75653 6.41149 3.92941C6.42523 4.10229 6.48617 4.26797 6.58758 4.40815L9.22873 8.14039C9.31559 8.26273 9.43019 8.36244 9.563 8.43122C9.69582 8.50001 9.84302 8.5359 9.99237 8.5359C10.1417 8.5359 10.2889 8.50001 10.4217 8.43122C10.5546 8.36244 10.6692 8.26273 10.756 8.14039L13.3972 4.40815C13.4978 4.26683 13.5578 4.10028 13.5706 3.92682C13.5833 3.75337 13.5483 3.57974 13.4693 3.42505C13.3904 3.27035 13.2706 3.14059 13.1231 3.05004C12.9756 2.95949 12.8062 2.91167 12.6335 2.91183H11.5513V0H8.43589V2.92162H7.34393C7.17202 2.92062 7.00318 2.96742 6.85599 3.05687C6.70881 3.14632 6.589 3.27494 6.50976 3.42856Z" fill="url(#paint0_linear)" />
                        <defs>
                            <linearGradient id="paint0_linear" x1="3.47049" y1="20.4694" x2="17.177" y2="8.1849" gradientUnits="userSpaceOnUse">
                                <stop stopColor="#7A6DC9" />
                                <stop offset="1" stopColor="#5DB9C8" />
                            </linearGradient>
                        </defs>
                    </svg>
                </span>
                <div onClick={setDrawerVisible.bind(this, true)} className={`cursor-pointer ${styles.eventsText}`}>
                    {queries.map((q, index) => {
                      if (index < queries.length - 1) {
                        return (
                                <React.Fragment key={index}>
                                    <span style={{ color: '#0E2647' }} className="text-xl font-semibold">{q}</span>
                                    <span style={{ color: '#8692A3' }} className="text-xl font-normal">&nbsp;and then&nbsp;</span>
                                </React.Fragment>
                        );
                      } else {
                        return (
                                <React.Fragment key={index}>
                                    <span style={{ color: '#0E2647' }} className="text-xl font-semibold">{q}</span>
                                </React.Fragment>
                        );
                      }
                    })}
                </div>
            </div>
        </div>
  );
}

export default EventsInfo;
