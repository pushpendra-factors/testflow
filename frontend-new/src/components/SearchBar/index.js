import React from 'react';
import { Input } from 'antd';
import styles from './index.module.scss';

function SearchBar() {
    return (
        <Input
            size="large"
            placeholder="Lookup factors.ai"
            prefix={(
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path fillRule="evenodd" clipRule="evenodd" d="M15.7397 12.5964C15.0736 14.575 13.2032 16 11 16C8.23858 16 6 13.7614 6 11C6 8.23858 8.23858 6 11 6C12.2329 6 13.3615 6.44621 14.2333 7.18593C15.3141 8.10308 16 9.47144 16 11C16 11.5582 15.9085 12.0951 15.7397 12.5964ZM17.6661 13.1424C17.4267 13.8879 17.0656 14.579 16.6064 15.1922L20.7071 19.2929C21.0976 19.6835 21.0976 20.3166 20.7071 20.7072C20.3166 21.0977 19.6834 21.0977 19.2929 20.7072L15.1921 16.6064C14.0236 17.4816 12.5723 18 11 18C7.13401 18 4 14.866 4 11C4 7.13401 7.13401 4 11 4C12.6281 4 14.1264 4.55582 15.3154 5.48807C16.9498 6.76948 18 8.7621 18 11C18 11.7473 17.8829 12.4672 17.6661 13.1424Z" fill="black" />
                </svg>
            )}
            className={styles.searchBarBox}
        />
    );
}

export default SearchBar;