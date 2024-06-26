import { ReactNode } from 'react';

export interface EditableTitleProps {
  title: string;
  editable?: boolean;
  handleEdit: (title: string) => void;
  editIcon?: ReactNode;
  enterIcon?: ReactNode;
  inputClassName?: string;
  titleClassName?: string;
}
