import React from 'react';
import { useHistory } from 'react-router-dom';

interface Props {
  name: string;
  label: string;
  checked: boolean;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const CheckBox = (props: Props) => {
  const history = useHistory();
  const handleOnChange = () => {
    history.push({
      pathname: history.location.pathname,
    });
  };

  return (
    <div className="custom-control custom-checkbox mr-sm-2 mt-2">
      <input
        type="checkbox"
        className="custom-control-input"
        id={props.name}
        onChange={handleOnChange}
        checked={props.checked}
        disabled
      />
      <label className="custom-control-label" htmlFor={props.name}>{props.label}</label>
    </div>
  );
}

export default CheckBox;
