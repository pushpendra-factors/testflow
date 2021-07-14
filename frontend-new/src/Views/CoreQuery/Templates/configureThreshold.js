const ConfigureThreshold = ({configMatrix}) => {
    const [isOpen, setIsOpen] = useState(false);
    const [form] = Form.useForm();
    const [errorInfo, seterrorInfo] = useState(null);
    const [dataLoading, setDataLoading] = useState(false);

    const onChange = () => {
      seterrorInfo(null);
    };

    const submitData = (data) => {
      if(data) console.log('submitData', data);
      form.validateFields().then((value) => {
        console.log("submit value",value)
      })
    }
 
      const onFinish = values => {
        console.log('Received values of form:', values);
      }; 
    return (<div>

      
        <Row>
          <Col span={12}>
            <Form
              form={form}
              name="configure"
              validateTrigger
              initialValues={{ remember: false }}
              onFinish={submitData}
              onChange={onChange}
            >

{isOpen && <>
{configMatrix?.map((item, index) => {
  return(
    <div className={'mb-2 flex items-center'}> 
                  <Text type={'title'} level={7} style={{width: '100px'}} extraClass={'m-0 capitalize'}>{item.display_name}</Text> 
                  <Form.Item label={null}
                    name={`${item.metric}_vc`}
                    // rules={[{ required: true }]}
                    className={'ml-4'}
                  >
                    <InputNumber className={'fa-input w-full ml-4'}  style={{width: '75px'}} min={0} defaultValue={10} disabled={dataLoading} /> 
                    </Form.Item> 
                  <Form.Item label={null}
                    name={`${item.metric}_av`}
                    // rules={[{ required: true }]}
                    className={'ml-2'}
                  > 
                    <InputNumber className={'fa-input w-full ml-2'} style={{width: '75px'}} min={0} defaultValue={0} disabled={dataLoading} />
                  </Form.Item>  
      </div>
  )
})}
<Text type={'title'} level={8} color={'grey'} extraClass={'m-0 capitalize'}>By default al the % change values are set to 10% and Absolute change are set to 0. You can change/overide the values by adding it above.</Text>
</>}
              
              {isOpen ? <div className={'flex items-center'}>
        <Button onClick={() => setIsOpen(false)}> Collapse </Button>
        <Button type={'primary'}  htmlType="submit" > Update </Button>
      </div> : <Button onClick={() => setIsOpen(true)}> Configure </Button>}


            </Form>
          </Col>
        </Row> 

    </div>)
  }



  export default ConfigureThreshold