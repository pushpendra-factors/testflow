import { SVG, Text } from 'Components/factorsComponents';
import { downloadInvoice, getInvoices } from 'Reducers/plansConfig/services';
import { GetInvoicesAPIResponse, Invoice } from 'Reducers/plansConfig/types';
import { Button, Divider, Table } from 'antd';
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import type { ColumnsType } from 'antd/es/table';
import moment from 'moment';
import logger from 'Utils/logger';

const InvoiceTab = () => {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [downloadLoadingId, setDownloadLoadingId] = useState<string>('');
  const { active_project } = useSelector((state) => state.global);

  useEffect(() => {
    const fetchInvoices = async () => {
      setLoading(true);
      try {
        const res = (await getInvoices(
          active_project?.id
        )) as GetInvoicesAPIResponse;
        if (res?.data) {
          setInvoices(res.data);
        }
        setLoading(false);
      } catch (error) {
        logger.error('Error in fetching invoices', error);
        setLoading(false);
      }
    };
    if (active_project?.id) fetchInvoices();
  }, [active_project]);

  const columns: ColumnsType<Invoice> = [
    {
      title: 'Invoice',
      dataIndex: 'id',
      key: 'id',
      render: (text) => (
        <div className='flex items-center gap-2'>
          <SVG name='RoundedFile' size='25' />
          {text}
        </div>
      )
    },
    {
      title: 'Billing date',
      dataIndex: 'billing_date',
      key: 'billing_date',
      render: (text) => <>{text ? moment(text).format('MMMM DD YYYY') : ''}</>
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (_, record) => <>{record.AmountDue <= 0 ? 'Paid' : 'Pending'}</>
    },
    {
      title: 'Amount',
      key: 'amount',
      dataIndex: 'Amount',
      render: (_, { Amount }) => `USD ${Amount}`
    },
    {
      title: 'Plan',
      key: 'plan',
      render: (_, record) => <>{record.items.join(', ')}</>
    },
    {
      title: '',
      key: '',
      render: (_, record) => {
        const getInvoice = async () => {
          try {
            setDownloadLoadingId(record.id);
            const res = await downloadInvoice(active_project?.id, record.id);
            if (res?.data?.url) {
              window.open(res?.data?.url, '_blank');
            }
            setDownloadLoadingId('');
          } catch (error) {
            logger.error('Error in fetching invoices', error);
            setDownloadLoadingId('');
          }
        };
        return (
          <Button
            onClick={() => getInvoice()}
            loading={downloadLoadingId === record.id}
            type='text'
            icon={<SVG name='Download' size='16' />}
          ></Button>
        );
      }
    }
  ];

  return (
    <div className='py-4'>
      <div className='mb-6'>
        <Text
          type={'title'}
          level={4}
          weight={'bold'}
          extraClass={'m-0 mb-2'}
          color='character-primary'
        >
          Billing and Invoicing
        </Text>
        <Text
          type={'title'}
          level={7}
          extraClass={'m-0'}
          color='character-secondary'
        >
          All invoices that have been generated
        </Text>
        <Divider />
        <Table
          loading={loading}
          columns={columns}
          dataSource={invoices}
          bordered
          pagination={false}
        />
      </div>
    </div>
  );
};

export default InvoiceTab;
