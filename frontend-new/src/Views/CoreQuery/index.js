import React from 'react';
import Header from './header';
import PageContent from './PageContent';
import QueryComposer from '../../components/QueryComposer';

import {Text} from '../../components/factorsComponents';

function CoreQuery() {
    return (
        <>
            <Header />
            <div>
                <QueryComposer visible={true}></QueryComposer>
            </div>
            <PageContent />
        </>
    )
}

export default CoreQuery;