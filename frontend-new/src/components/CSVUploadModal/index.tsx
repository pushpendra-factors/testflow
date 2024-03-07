import Papa from 'papaparse';
import { UploadOutlined } from '@ant-design/icons';
import AppModal from 'Components/AppModal';
import { SVG, Text } from 'Components/factorsComponents';
import logger from 'Utils/logger';
import { Button, Col, List, Row, Upload } from 'antd';
import React, { useState } from 'react';
import csvTableImage from '../../assets/images/csvTable.svg';
import style from './index.module.scss';

const sampleCSVFileURLs =
  'https://s3.amazonaws.com/www.factors.ai/assets/files/Sample_file_for_page_URL_rules.csv';
const sampleCSVFileFilters =
  'https://s3.amazonaws.com/www.factors.ai/assets/files/sample_company_domains.csv';

type Props = {
  uploadModalOpen: boolean;
  setUploadModalOpen(data: boolean): void;
  handleOkClick(data: string[], name: string): void;
  uploadType: string;
};

function CSVUploadModal({
  uploadModalOpen,
  setUploadModalOpen,
  handleOkClick,
  uploadType = 'filters'
}: Props) {
  const [uploadFileName, setUploadFileName] = useState('');
  const [uploadFileArray, setUploadFileArray] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [isCSVFile, setIsCSVFile] = useState(false);
  const [errorState, setErrorState] = useState<string>('');

  const readFile = (file: Blob | MediaSource) =>
    new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (event) => {
        if (event.target?.result) {
          resolve(event.target.result as string);
        } else {
          reject(new Error('Failed to read file'));
        }
      };
      reader.onerror = (event) => {
        reject(new Error('Failed to read file'));
      };
      reader.readAsText(file);
    });

  const parseFile = (text: string) => {
    const { data } = Papa.parse(text);
    return data.map((row: any) => row[0]);
  };

  const handleFileChange = async (info: {
    file: { originFileObj: Blob | MediaSource };
  }) => {
    if (info?.file?.originFileObj) {
      try {
        const text = await readFile(info.file.originFileObj);
        const dataArray = parseFile(text);

        if (uploadType === 'filters') {
          const reader = new FileReader();
          const fileByteArray: any = [];
          reader.readAsArrayBuffer(info?.file?.originFileObj);
          reader.onloadend = function (evt) {
            if (evt?.target?.readyState === FileReader.DONE) {
              const arrayBuffer = evt.target.result;
              const array = new Uint8Array(arrayBuffer);
              for (let i = 0; i < array.length; i++) {
                fileByteArray.push(array[i]);
              }
            }
          };
          setUploadFileArray(fileByteArray);
        } else {
          setUploadFileArray(dataArray);
        }

        setUploadFileName(info?.file?.name);

        if (info?.file?.type !== 'text/csv')
          throw 'Only .csv files are allowed';
        if (info?.file?.size > 1000000 && uploadType === 'filters')
          throw 'File is larger than 1 MB';
        if (dataArray.length > 50 && uploadType === 'urls')
          throw 'Can’t upload a sheet with more than 50 URLs';
        if (dataArray.length > 10000 && uploadType === 'filters')
          throw 'Can’t upload a sheet with more than 10,000 rows';

        setIsCSVFile(true);
        setErrorState('');
      } catch (error: any) {
        setIsCSVFile(false);
        setErrorState(error);
        logger.error(error);
      }
    }
  };

  const handleCancel = () => {
    setUploadModalOpen(false);
    setUploadFileName('');
    setUploadFileArray([]);
    setErrorState('');
    setIsCSVFile(false);
  };

  const handleOk = async () => {
    setLoading(true);
    try {
      if (uploadFileArray.length === 0) {
        logger.error('error: empty file');
        return;
      }
      await handleOkClick(uploadFileArray, uploadFileName);
      handleCancel();
    } catch (error) {
      logger.error(error);
    }
    setLoading(false);
  };

  const listDataURLs = [
    <span>
      Add values in the first column only with{' '}
      <span className='font-bold'>no header</span>
    </span>,
    <span>
      You can upload a maximum of <span className='font-bold'>50 URLs</span>
    </span>,
    <span>
      Ensure that the file has a <span className='font-bold'>.csv</span>{' '}
      extension only
    </span>,
    <span>
      Don’t include <span className='font-bold'>https://</span> in the URL
    </span>
  ];

  const listDataFilters = [
    <span>
      Add values in the first column only with{' '}
      <span className='font-bold'>no header</span>
    </span>,
    <span>
      Ensure your CSV has less than{' '}
      <span className='font-bold'>10,000 rows</span>
    </span>,
    <span>
      Ensure that the file has a <span className='font-bold'>.csv</span>{' '}
      extension only
    </span>,
    <span>
      Ensure that the file size is less than{' '}
      <span className='font-bold'>1MB</span>
    </span>
  ];

  return (
    <AppModal
      visible={uploadModalOpen}
      width={780}
      closable
      title={null}
      footer={null}
      className={style.container}
      onCancel={() => handleCancel()}
    >
      <div>
        <div>
          <Text type='title' level={4} weight='bold' extraClcalass='m-0'>
            Upload CSV
          </Text>
          <Text type='title' level={6} color='grey' extraClass='m-0 -mt-2'>
            {uploadType === 'filters'
              ? 'Import a list of accounts, domains, and page URLs to filter with.'
              : 'Import a CSV with list of page URLs to include or exclude'}
          </Text>
        </div>
        <div className='mt-4 mb-8'>
          <div className='flex justify-between'>
            <div>
              <Text type='title' level={6} extraClcalass='m-0'>
                Please note the following before uploading
              </Text>
              <List
                header={null}
                footer={null}
                split={false}
                dataSource={
                  uploadType === 'filters' ? listDataFilters : listDataURLs
                }
                renderItem={(item) => (
                  <List.Item>
                    <Text
                      type='title'
                      level={7}
                      color='grey'
                      extraClass='m-0 -mt-1 ml-2'
                    >
                      <span className={style.dot} />
                      {item}
                    </Text>
                  </List.Item>
                )}
              />
              <Text type='title' level={7} color='grey' extraClass='m-0'>
                In case of any doubts, here is a sample{' '}
                <a
                  href={
                    uploadType === 'filters'
                      ? sampleCSVFileFilters
                      : sampleCSVFileURLs
                  }
                  target='_blank'
                  rel='noreferrer'
                >
                  file
                </a>
              </Text>
            </div>
            <div>
              <div>
                <img src={csvTableImage} alt='csv table' />
                <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
                  We read values in the first column starting from A1
                </Text>
              </div>
            </div>
          </div>
        </div>
        <div className='border rounded mt-2 flex justify-center '>
          <Upload
            showUploadList={false}
            onChange={handleFileChange}
            accept='.csv'
            maxCount={1}
            className='text-center'
          >
            <div className='p-8'>
              {uploadFileName ? (
                <Button className='inline'>
                  {uploadFileName}
                  <SVG extraClass='ml-1' name='close' color='grey' />
                </Button>
              ) : (
                <Button icon={<UploadOutlined />}>Upload CSV</Button>
              )}
            </div>
          </Upload>
        </div>
        {errorState && (
          <div>
            <Text type='title' level={7} color='red' extraClass='m-0'>
              {errorState}
            </Text>
          </div>
        )}
        <Row className='mt-4'>
          <Col span={24}>
            <div className='flex justify-end'>
              <Button
                size='large'
                className='mr-2'
                onClick={() => handleCancel()}
              >
                Cancel
              </Button>
              <Button
                size='large'
                className='ml-2'
                type='primary'
                onClick={() => handleOk()}
                disabled={!uploadFileName || !isCSVFile}
                loading={loading}
              >
                Done
              </Button>
            </div>
          </Col>
        </Row>
      </div>
    </AppModal>
  );
}

export default CSVUploadModal;
