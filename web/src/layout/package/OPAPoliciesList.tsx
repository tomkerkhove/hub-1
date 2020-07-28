import React, { useState } from 'react';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { tomorrowNightBright } from 'react-syntax-highlighter/dist/cjs/styles/hljs';

import { OPAPolicies } from '../../types';
import Accordion from '../common/Accordion';

interface Props {
  policies: OPAPolicies;
}

const OPAPoliciesList = (props: Props) => {
  const [isOpen, setOpen] = useState<string | null>(null);
  return (
    <>
      {Object.keys(props.policies).map((policy: string, index: number) => {
        return (
          <Accordion key={`policy_${index}`} title={policy} isOpen={policy === isOpen} setOpenStatus={setOpen}>
            <SyntaxHighlighter language="rego" style={tomorrowNightBright} customStyle={{ padding: '1.5rem' }}>
              {props.policies[policy]}
            </SyntaxHighlighter>
          </Accordion>
        );
      })}
    </>
  );
};

export default OPAPoliciesList;
