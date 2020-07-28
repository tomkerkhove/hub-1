import classnames from 'classnames';
import React, { Dispatch, SetStateAction } from 'react';
import { FaCaretDown, FaCaretRight } from 'react-icons/fa';

import styles from './Accordion.module.css';

interface Props {
  isOpen: boolean;
  setOpenStatus: Dispatch<SetStateAction<string | null>>;
  title: string;
  children: JSX.Element;
}

const Accordion = (props: Props) => {
  return (
    <div className="mt-3">
      <div
        className={classnames('btn btn-outline-secondary btn-block rounded-pill', {
          [`btn-secondary text-light ${styles.open}`]: props.isOpen,
        })}
        onClick={() => props.setOpenStatus(props.isOpen ? null : props.title)}
      >
        <div className="d-flex flex-row justify-content-between align-items-center px-2">
          <div className="text-left">{props.title}</div>
          <div className="ml-3">{props.isOpen ? <FaCaretDown /> : <FaCaretRight />}</div>
        </div>
      </div>
      <div className={classnames('overflow-hidden', styles.accordionItem, { [styles.collapsed]: !props.isOpen })}>
        <div className="pt-3 px-3">{props.children}</div>
      </div>
    </div>
  );
};

export default Accordion;
