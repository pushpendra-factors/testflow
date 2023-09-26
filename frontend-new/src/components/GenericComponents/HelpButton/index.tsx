import { SVG } from 'Components/factorsComponents';
import { Button } from 'antd';
import React from 'react';

const HelpButton = ({ helpMessage }: HelpButtonProps) => {
  const handleHelpClick = () => {
    if (window?.Intercom) {
      window.Intercom('showNewMessage', helpMessage ? helpMessage : '');
    }
  };
  return (
    <Button
      icon={<SVG name='Headset' size='16' color='#8C8C8C' />}
      onClick={handleHelpClick}
    >
      Need help?
    </Button>
  );
};

interface HelpButtonProps {
  helpMessage?: string;
}

export default HelpButton;
