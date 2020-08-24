import React from 'react';
// import SearchBar from '../../components/SearchBar';
import SaveQueryButton from './SaveQueryButton';

function Header() {
    return (
        <div className="flex items-center">
            <div className="w-1/3"></div>
            <div className="w-1/3 flex justify-center items-center">
                {/* <SearchBar /> */}
            </div>
            <div className="w-1/3 flex justify-end">
                <SaveQueryButton />
            </div>
        </div>
    )
}

export default Header;