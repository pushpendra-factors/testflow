import React, {
    useState,
    useEffect,
    useCallback,
    useRef,
    useMemo
  } from 'react';
  import { Text, SVG } from 'factorsComponents';
  import { Tag } from 'antd';
import { getMsgPayloadMapping, dummyPayloadValue, getMsgPayloadMappingWebhook} from './../utils';
import ReactJson from 'react-json-view'

export const PreviewCardSlack = ({
    alertName,
    alertMessage,
    groupBy,
    selectedMentions,
    matchEventName
}) =>{ 
    let payloadProps = groupBy?.length>0 ? getMsgPayloadMapping(groupBy) : {};
    return (
        <div>
        {/* slack card starts here*/}
        <div className='background-color--mono-color-1 border--thin-2 p-4' style={{'width': '400px', 'border-radius': '8px', 'height': '320px', 'overflowY':'auto'}}>

        <div className='flex flex-col justify-start items-start'> 

        <div className='flex items-center'>
        <SVG name={'brand'} background='transparent' showBorder={false} size={32} />
        <Text type='title' level={7} weight={'bold'} extraClass='m-0 ml-2' >Factors.ai</Text>
        <span className='ml-2 mr-2'><Tag className='fa-tag--small-grey'>App</Tag></span>
        <Text type='title' level={8} weight={'thin'} extraClass='m-0' >3:00 PM</Text>
        </div>

        <div className='pl-10 w-full'>


        <div className='flex flex-col items-start mt-2'>
        <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{alertName ? alertName : 'Alert name'}</Text>
        <Text type='title' level={8} extraClass='m-0' >{alertMessage ? alertMessage : 'Alert message to be displayed'}</Text>
        <Text type='title' level={8}  color={'blue'} extraClass='m-0' >
            {selectedMentions?.map((item)=>{
                return `@${item} `
            })}
            </Text>
        <Text type='title' level={8} weight={'bold'} color={'blue'} extraClass='m-0 mt-2' >{'See activity in Factors app'}</Text>
        </div>

        <div className='mt-4 mb-2 border-top--thin-2'>
        <div className='flex flex-wrap mt-4 mb-2 border-left--blue-color-thick pl-2'>
         

 
   {groupBy?.length>0 ? payloadProps && Object?.entries(payloadProps).map(([key, value]) => { 
          return (<div className='px-2 py-2 w-1/2'>
              <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{matchEventName(key)}</Text>
              <Text type='title' level={8} color={'grey'} extraClass='m-0' >{dummyPayloadValue[key] ? dummyPayloadValue[key] : value}</Text>
          </div>)
        }) :
        [1,2,3,4].map((item) => { 
            return (<div className='px-2 py-2 w-1/2'>
                <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{`Property ${item}`}</Text>
                <Text type='title' level={8} color={'grey'} extraClass='m-0' >{'${Property Value}'}</Text>
            </div>)
          })
    }


        </div>
        </div>

        </div>

        </div> 

        </div> 
        {/* slack card ends here*/}
        <Text type='title' level={8} color={'grey'} extraClass='m-0 mt-2' >This is a preview of how your alert will look in Slack, using dummy data.</Text>
        </div>

    )
} 



export const PreviewCardTeams = ({
    alertName,
    alertMessage,
    groupBy,
    matchEventName
}) =>{

    let payloadProps = groupBy?.length>0 ? getMsgPayloadMapping(groupBy) : {};
    return (
        <div>

        {/* teams card starts here*/}
        <div className='background-color--mono-color-1 border--thin-2 ' style={{'width': '400px', 'border-radius': '8px', 'height': '320px', 'overflowY':'auto'}}>

 
        <div className='flex items-center justify-between border-bottom--thin-2 py-2 px-4'>
            <div className='flex items-center'>
            <SVG name={'brand'} background='transparent' showBorder={false} size={32} />
            <Text type='title' level={7} weight={'bold'} extraClass='m-0 ml-2' >Factors.ai</Text>
            </div>
        <Text type='title' level={8} weight={'thin'} extraClass='m-0 ml-2' >3:00 PM</Text>
        </div>

        <div className='flex flex-col justify-start items-start px-4'>
        
        <div className='w-full p-4'>


        <div className='flex flex-col items-start mt-2'>
        <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{alertName ? alertName : 'Alert name'}</Text>
        <Text type='title' level={7} extraClass='m-0' >{alertMessage ? alertMessage : 'Alert message to be displayed'}</Text>
        </div>

        <div className='mt-4 mb-2'>
        <div className='flex flex-col flex-wrap mt-4 mb-2'>
   
   {groupBy?.length>0 ? payloadProps && Object?.entries(payloadProps).map(([key, value]) => { 
          return (<div className='flex items-center w-full justify-between flex-wrap mt-2'>
              <Text type='title' level={7} color={'grey'}  extraClass='m-0' >{matchEventName(key)}</Text>
              <Text type='title' level={7} extraClass='m-0' >{dummyPayloadValue[key] ? dummyPayloadValue[key] : value}</Text>
          </div>)
        }) :
        [1,2,3,4].map((item) => { 
            return (<div className='flex items-center w-full justify-between'>
              <Text type='title' level={7} color={'grey'}  extraClass='m-0 mt-1' >{`Property ${item}`}</Text>
              <Text type='title' level={7} extraClass='m-0' >{'${Property Value}'}</Text>
          </div>)
          })
    }

        </div>
        </div>

        </div>

        </div> 

        </div> 
        {/* teams card ends here*/}
            <Text type='title' level={8} color={'grey'} extraClass='m-0 mt-2' >This is a preview of how your alert will look in Teams, using dummy data.</Text>
        </div>
    )
}



export const PreviewCardWebhook = ({
    alertName,
    alertMessage,
    groupBy,
    selectedEvent,
    matchEventName,
    factorsURLinWebhook,
    activeGrpBtn
}) =>{

    let payloadProps = {};
    payloadProps['Title']= alertName ? alertName : 'Alert name';
    payloadProps['Message']= alertMessage ? alertMessage : 'Alert message to be displayed';
    payloadProps['Event']= selectedEvent ? selectedEvent : ''; 

    payloadProps['MessageProperty'] = groupBy?.length>0 ? getMsgPayloadMappingWebhook(groupBy, matchEventName, dummyPayloadValue) : [];

    if(factorsURLinWebhook){
        let url = ""
        if(activeGrpBtn == 'events' || activeGrpBtn == 'account'){
            url = "https://app.factors.ai/profiles/accounts/{account-id}";
        }
        else{
            url = "https://app.factors.ai/profiles/people";
        }
        let obj = {
            "DisplayName": "Factors Activity URL",
            "PropValue": url
        }
        payloadProps['MessageProperty'].push(obj);
    }
    
    return (
        <div>

        {/* card starts here*/}
        <div className='background-color--mono-color-1 border--thin-2 p-2' style={{'width': '400px', 'border-radius': '8px', 'height': '320px', 'overflowY':'auto'}}>

        { !groupBy?.length>0 ? 
        <div className='flex flex-col items-center justify-center' style={{'minHeight': '250px'}}>  
            <Text type='title' level={7} color={'grey'} weight={'thin'} extraClass='m-0' >Add properties to preview</Text>
        </div>
        
        : <ReactJson src={payloadProps}
                className={'fa-custom--reactjson'}
                style={{'minHeight':'300px'}}
                // theme={'summerfruit:inverted'}
                displayDataTypes={false}
                displayObjectSize={false}
                enableClipboard={false} 
                quotesOnKeys={false}

            /> } 

        </div> 
        {/*  card ends here*/}
            <Text type='title' level={8} color={'grey'} extraClass='m-0 mt-2' >This is a preview of your alert response</Text>
        </div>
    )
}