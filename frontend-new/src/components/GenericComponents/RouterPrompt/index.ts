import { useCallback, useEffect, useState } from 'react';
import { useHistory } from 'react-router';

import { Modal } from 'antd';

const { confirm } = Modal;

type RouterPromptProps = {
  when: boolean;
  okText: string;
  cancelText: string;
  title: string;
  content: string;
  onOK: () => boolean;
  onCancel: () => boolean;
};
const RouterPrompt = ({
  when,
  onOK,
  onCancel,
  title,
  content,
  okText,
  cancelText
}: RouterPromptProps) => {
  const history = useHistory();

  const [showPrompt, setShowPrompt] = useState(false);
  const [currentPath, setCurrentPath] = useState('');

  useEffect(() => {
    if (when) {
      history.block((prompt) => {
        setCurrentPath(prompt.pathname);
        setShowPrompt(true);
        return 'true';
      });
    } else {
      history.block(() => {});
    }

    return () => {
      history.block(() => {});
    };
  }, [history, when]);

  const handleOK = useCallback(async () => {
    if (onOK) {
      const canRoute = await Promise.resolve(onOK());
      if (canRoute) {
        history.block(() => {});
        history.push(currentPath);
      }
    }
  }, [currentPath, history, onOK]);

  const handleCancel = useCallback(async () => {
    if (onCancel) {
      const canRoute = await Promise.resolve(onCancel());
      if (canRoute) {
        history.block(() => {});
        history.push(currentPath);
      }
    }
    setShowPrompt(false);
  }, [currentPath, history, onCancel]);

  useEffect(() => {
    if (showPrompt) {
      confirm({
        title,
        content,
        okText,
        cancelText,
        onOk() {
          handleOK();
        },
        onCancel() {
          handleCancel();
        }
      });
    }
  }, [showPrompt]);
  return null;
};

export default RouterPrompt;
