import React, { useState } from 'react';
import Header from './header';
import PageContent from './PageContent';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';

function CoreQuery() {
    const [drawerVisible, setDrawerVisible] = useState(false);
    const [showResult, setShowResult] = useState(false);
    const [queries, setQueries] = useState([]);

    const queryChange = (newEvent, index, changeType = 'add') => {
        const queryupdated = [...queries];
        if(queryupdated[index]) {
            if(changeType === 'add') {
                queryupdated[index] = newEvent;
            } else {
                queryupdated.splice(index, 1);
            }
            
        } else {
            queryupdated.push(newEvent);
        }
        setQueries(queryupdated);
    }

    const runQuery = () => {
        setShowResult(true);
        closeDrawer();
    }

    const closeDrawer = () => {
        setDrawerVisible(false);
    }

    return (
        <>

            <QueryComposer
                drawerVisible={drawerVisible}
                queries={queries}
                onClose={closeDrawer}
                runQuery={runQuery}
                eventChange={queryChange}
            >

            </QueryComposer>

            {
                showResult ? (<PageContent setDrawerVisible={setDrawerVisible} queries={queries.map(elem => elem.label)} />) : (<CoreQueryHome setDrawerVisible={setDrawerVisible} />)
            }

        </>
    )
}

export default CoreQuery;