import { SVG, Text } from 'Components/factorsComponents';
import React from 'react';

const password_checks = [
  {
    name: 'Atleast 8 characters',
    validateFunction: (value: string) => value?.length >= 8
  },
  {
    name: 'Include one number',
    validateFunction: (value: string) => value.match(/(?=.*?[0-9])/)
  },
  {
    name: 'Both lower and uppercase letters',
    validateFunction: (value: string) => value.match(/(?=.*[a-z])(?=.*[A-Z])/)
  },
  {
    name: 'Special character',
    validateFunction: (value: string) => value.match(/(?=.*?[#?!@$%^&*-])/)
  }
];

const PasswordChecks = ({ password = '' }: { password: string }) => {
  return (
    <div className='flex flex-col gap-2 mt-6'>
      {password_checks.map((passwordCheck) => {
        const isValid = passwordCheck.validateFunction(password);
        return (
          <div className='flex gap-2 items-center'>
            <SVG
              name={isValid ? 'CheckCircle' : 'CloseCircle'}
              size='16'
              color={isValid ? '#52C41A' : '#F5222D'}
            />
            <Text
              type={'paragraph'}
              mini
              extraClass={'m-0'}
              color='character-primary'
            >
              {passwordCheck.name}
            </Text>
          </div>
        );
      })}
    </div>
  );
};

export default PasswordChecks;
