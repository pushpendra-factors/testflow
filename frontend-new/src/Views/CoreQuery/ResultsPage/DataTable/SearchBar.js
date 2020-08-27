import React, { useState } from 'react';
import styles from './index.module.scss';
import { Input } from 'antd';
import { tree } from 'd3';

function SearchBar() {

    const [searchBar, showSearchBar] = useState(false);

    const downloadCSV = () => {
        console.log("download csv");
    }

    const downloadBtn = (
        <div onClick={downloadCSV} className="flex flex-1 items-center justify-end cursor-pointer">
            <div className="mr-2">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path fillRule="evenodd" clipRule="evenodd" d="M12 3C12.5523 3 13 3.44772 13 4V11.5858L15.2929 9.29289C15.6834 8.90237 16.3166 8.90237 16.7071 9.29289C17.0976 9.68342 17.0976 10.3166 16.7071 10.7071L12 15.4142L7.29289 10.7071C6.90237 10.3166 6.90237 9.68342 7.29289 9.29289C7.68342 8.90237 8.31658 8.90237 8.70711 9.29289L11 11.5858V4C11 3.44772 11.4477 3 12 3ZM20.0255 15.1049C20.0255 14.5527 19.5778 14.1049 19.0255 14.1049C18.4732 14.1049 18.0255 14.5527 18.0255 15.1049V18H6.03497V15.105C6.03497 14.5527 5.58726 14.105 5.03497 14.105C4.48269 14.105 4.03497 14.5527 4.03497 15.105V20H20.0255V15.1049Z" fill="#0E2647" />
                </svg>
            </div>
            <div className={styles.downloadCSVHeading}>
                Download CSV
            </div>
        </div>
    );


    return (
        <div className={`${styles.searchBar}`}>
            {!searchBar ? (
                <div className="flex p-4 w-full">
                    <div onClick={showSearchBar.bind(this, true)} className={`flex items-center w-3/4 cursor-pointer`}>
                        <div className="mr-2">
                            <svg width="17" height="18" viewBox="0 0 17 18" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path fillRule="evenodd" clipRule="evenodd" d="M13.6661 9.22917C13.4267 9.97472 13.0656 10.6658 12.6064 11.279L16.7071 15.3797C17.0976 15.7703 17.0976 16.4034 16.7071 16.794C16.3166 17.1845 15.6834 17.1845 15.2929 16.794L11.1921 12.6932C10.0236 13.5684 8.57232 14.0868 7 14.0868C3.13401 14.0868 0 10.9528 0 7.08679C0 3.2208 3.13401 0.086792 7 0.086792C8.6281 0.086792 10.1264 0.642616 11.3154 1.57486C12.9498 2.85628 14 4.84889 14 7.08679C14 7.83405 13.8829 8.55396 13.6661 9.22917ZM11.7397 8.68317C11.0736 10.6618 9.20321 12.0868 7 12.0868C4.23858 12.0868 2 9.84822 2 7.08679C2 4.32537 4.23858 2.08679 7 2.08679C8.23286 2.08679 9.36151 2.533 10.2333 3.27273C11.3141 4.18987 12 5.55823 12 7.08679C12 7.64501 11.9085 8.18186 11.7397 8.68317Z" fill="black" />
                            </svg>
                        </div>
                        <div className={styles.breakupHeading}>
                            Break-up
                        </div>
                    </div>
                    {downloadBtn}
                </div>
            ) : (
                    <Input
                        className={`${styles.inputSearchBar}`}
                        size="large"
                        placeholder="Search for"
                        prefix={(
                            <svg width="17" height="18" viewBox="0 0 17 18" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path fillRule="evenodd" clipRule="evenodd" d="M13.6661 9.22917C13.4267 9.97472 13.0656 10.6658 12.6064 11.279L16.7071 15.3797C17.0976 15.7703 17.0976 16.4034 16.7071 16.794C16.3166 17.1845 15.6834 17.1845 15.2929 16.794L11.1921 12.6932C10.0236 13.5684 8.57232 14.0868 7 14.0868C3.13401 14.0868 0 10.9528 0 7.08679C0 3.2208 3.13401 0.086792 7 0.086792C8.6281 0.086792 10.1264 0.642616 11.3154 1.57486C12.9498 2.85628 14 4.84889 14 7.08679C14 7.83405 13.8829 8.55396 13.6661 9.22917ZM11.7397 8.68317C11.0736 10.6618 9.20321 12.0868 7 12.0868C4.23858 12.0868 2 9.84822 2 7.08679C2 4.32537 4.23858 2.08679 7 2.08679C8.23286 2.08679 9.36151 2.533 10.2333 3.27273C11.3141 4.18987 12 5.55823 12 7.08679C12 7.64501 11.9085 8.18186 11.7397 8.68317Z" fill="black" />
                            </svg>
                        )}
                        suffix={downloadBtn}
                    />
                )}
        </div>
    )
}

export default SearchBar;