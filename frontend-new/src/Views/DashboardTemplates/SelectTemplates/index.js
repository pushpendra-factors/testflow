import React, { useState, useCallback, useEffect } from "react";
import {Spin,Input, Button} from "antd";
import { useSelector, useDispatch } from "react-redux";
import {
    Text,
    SVG,
    FaErrorComp,
    FaErrorLog,
  } from '../../../components/factorsComponents'

import TemplateCard from "./TemplateCard";
import { SearchOutlined } from '@ant-design/icons';
import { ErrorBoundary } from 'react-error-boundary';
import styles from './index.module.scss';
import TemplateDetails from "./TemplateDetails";
import { ACTIVE_TEMPLATE_CHANGE } from 'Reducers/types';


function SelectTemplates({setShowTemplates,templates}){
    const [searchVal, setSearchVal] = useState('');
    const [showCardDetails, setShowCardDetails] = useState(false);
    const [TemplateSelected, setTemplateSelected] = useState(-1);
    const dispatch = useDispatch();

    useEffect(() => {
        if(TemplateSelected != -1){
            const selectedTemplate = templates.data.find((d) => d.id === TemplateSelected);
            dispatch({
            type: ACTIVE_TEMPLATE_CHANGE,
            payload: selectedTemplate
            });
            setShowCardDetails(true);
        }
    }, [TemplateSelected]);

    const handleSearchChange = useCallback((e) => {
      setSearchVal(e.target.value);
    }, []);

    const renderTemplateCards=()=>{
        const filteredTemplates = templates.data.filter(
            (q) => q.title.toLowerCase().indexOf(searchVal.toLowerCase()) > -1
        );
        return filteredTemplates.map(t=><TemplateCard id = {t.id} title={t.title} description={t.description} setTemplateSelected={setTemplateSelected}/>)
    }
    if (templates.loading) {
        return (
          <div className='flex justify-center items-center w-full h-64'>
            <Spin size='large' />
          </div>
        );
      }
    
        return(
            <div>
                { showCardDetails && 
                    <div className="ant-modal-wrap bg-white">
                        < TemplateDetails 
                        setShowTemplates={setShowTemplates}
                        setShowCardDetails={setShowCardDetails}
                        TemplateSelected={TemplateSelected}
                        templates={templates}
                        setTemplateSelected={setTemplateSelected}
                        />
                    </div>
                }
                <ErrorBoundary
                    fallback={
                    <FaErrorComp
                        size={'medium'}
                        title={'Analyse LP Error'}
                        subtitle={
                        'We are facing trouble loading Analyse landing page. Drop us a message on the in-app chat.'
                        }
                    />
                    }
                    onError={FaErrorLog}>
                    
                    <div className='flex flex-col h-full'>
                    {/* <Header> */}
                        <Button
                            size={'large'}
                            type='text'
                            icon={<SVG size={20} name={'times'} />}
                            onClick={()=>setShowTemplates(false)} 
                            className={styles.close}
                        />
                        <div className='w-full h-full py-4 flex flex-col justify-center items-center mt-24'>
                            <div className="w-full h-full flex flex-col justify-center items-center mb-3">
                            <Text type='title' level={4} weight={'bold'} >Pick From Dashboard Templates</Text>
                            <Text type='paragraph' >Browse the templates from our wide range of commonly used reports.</Text>
                            <Text type='paragraph' >Curated by top marketers in the industry.</Text>
                            </div>
                            <div className={`query-search flex flex-col`}>
                                <Input
                                onChange={handleSearchChange}
                                value={searchVal}
                                className={'fa-global-search--input btn-total-round'}
                                placeholder='Search all templates'
                                prefix={<SearchOutlined style={{ width: '1rem' }} color='#0E2647' />}
                                />
                            </div>
                        </div>
                    {/* </Header> */}
                    </div>
                    {templates.data.length>0 &&(
                    <div className={'mx-32 mt-10 justify-center grid grid-cols-3 gap-12'}>
                            {renderTemplateCards()}
                            {/* </div> */}
                        {/* </Row> */}
                    </div>)
                    }
                </ErrorBoundary>
            </div>
        );
}
export default SelectTemplates;