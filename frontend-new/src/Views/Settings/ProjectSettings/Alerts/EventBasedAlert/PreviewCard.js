import React, {
    useState,
    useEffect,
    useCallback,
    useRef,
    useMemo
  } from 'react';
  import { Text, SVG } from 'factorsComponents';
  import { Tag } from 'antd';
import { getMsgPayloadMapping} from './../utils';

export const PreviewCardSlack = ({
    alertName,
    alertMessage,
    groupBy,
    selectedMentions
}) =>{

    let payloadProps = groupBy ? getMsgPayloadMapping(groupBy) : {}
    return (
        <div>
        {/* slack card starts here*/}
        <div className='background-color--mono-color-1 border--thin-2 p-4' style={{'width': '400px', 'border-radius': '8px'}}>  
        
        <div className='flex flex-col justify-start items-start'>
        
        <div className='flex items-center'>
        <SVG name={'brand'} background='transparent' showBorder={false} size={32} />
        <Text type='title' level={7} weight={'bold'} extraClass='m-0 ml-2' >Factors.ai</Text>
        <span className='ml-2 mr-2'><Tag className='fa-tag--small-grey'>App</Tag></span>
        <Text type='title' level={8} weight={'thin'} extraClass='m-0' >3:00 PM</Text>
        </div>

        <div className='ml-10'>


        <div className='flex flex-col items-start mt-2'>
        <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{alertName}</Text>
        <Text type='title' level={8} extraClass='m-0' >{alertMessage}</Text>
        <Text type='title' level={8}  color={'blue'} extraClass='m-0' >
            {selectedMentions?.map((item)=>{
                return `@${item} `
            })}
            </Text>
        </div>

        <div className='mt-4 mb-2 border-top--thin-2'>
        <div className='flex flex-wrap mt-4 mb-2 border-left--blue-color-thick pl-2'>
         

 
   { 
    Object.entries(payloadProps).map(([key, value]) => { 
          return (<div className='px-2 py-2'>
              <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{key}</Text>
              <Text type='title' level={8} color={'grey'} extraClass='m-0' >{value}</Text>
          </div>)
        })
    }


        </div>
        </div>

        </div>

        </div>
        
        </div> 
        {/* slack card ends here*/}
        <Text type='title' level={8} color={'grey'} extraClass='m-0 mt-2' >This is a preview of how your alert will look in slack</Text>
        </div>

    )
} 



export const PreviewCardTeams = ({
    alertName,
    alertMessage,
    groupBy,
}) =>{

    let payloadProps = groupBy ? getMsgPayloadMapping(groupBy) : {}
    return (
        <div>

        {/* teams card starts here*/}
        <div className='background-color--mono-color-1 border--thin-2 ' style={{'width': '400px', 'border-radius': '8px'}}>  

        <div className='flex items-center justify-between border-bottom--thin-2 py-2 px-4'>
            <div className='flex items-center'>
            <SVG name={'brand'} background='transparent' showBorder={false} size={32} />
            <Text type='title' level={7} weight={'bold'} extraClass='m-0 ml-2' >Factors.ai</Text>
            </div>
        <Text type='title' level={8} weight={'thin'} extraClass='m-0 ml-2' >3:00 PM</Text>
        </div>

        <div className='flex flex-col justify-start items-start p-4'>
        
        <div className='ml-4'>


        <div className='flex flex-col items-start mt-2'>
        <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{alertName}</Text>
        <Text type='title' level={7} extraClass='m-0' >{alertMessage}</Text>
        </div>

        <div className='mt-4 mb-2 '>
        <div className='flex flex-wrap mt-4 mb-2 border-left--blue-color-thick pl-2'>
         

 
   { 
    payloadProps && Object?.entries(payloadProps).map(([key, value]) => { 
          return (<div className='px-2 py-2'>
              <Text type='title' level={7} weight={'bold'} extraClass='m-0' >{key}</Text>
              <Text type='title' level={8} color={'grey'} extraClass='m-0' >{value}</Text>
          </div>)
        })
    }


        </div>
        </div>

        </div>

        </div>
        
        </div> 
        {/* teams card ends here*/}
            <Text type='title' level={8} color={'grey'} extraClass='m-0 mt-2' >This is a preview of how your alert will look in slack</Text>
        </div>
    )
}



export const PreviewCardWebhook = ({
    alertName,
    alertMessage,
    groupBy,
}) =>{

    let payloadProps = groupBy ? getMsgPayloadMapping(groupBy) : {};
    payloadProps['title']= alertName;
    payloadProps['message']= alertMessage;

    return (
        <div>

        {/* card starts here*/}
        <div className='background-color--mono-color-1 border--thin-2 ' style={{'width': '400px', 'border-radius': '8px'}}>  
        <pre>
        <code className='fa-code-code-block'>
          {JSON.stringify(payloadProps)}
        </code>
      </pre>
         
        </div> 
        {/*  card ends here*/}
            <Text type='title' level={8} color={'grey'} extraClass='m-0 mt-2' >This is a preview of how your alert will look in slack</Text>
        </div>
    )
}