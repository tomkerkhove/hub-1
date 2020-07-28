import React, { useState, useEffect, useContext } from 'react';
import classnames from 'classnames';

import { AppCtx, updateTheme } from '../../context/AppCtx';
import styles from './DarkMode.module.css';
import { IoIosMoon } from 'react-icons/io';
import { isUndefined } from 'lodash';

interface Props {
  darkVersion?: boolean;
}

const DEFAULT_THEME = 'theme';

const DarkMode = (props: Props) => {
  const { ctx, dispatch } = useContext(AppCtx);
  const activeTheme = ctx.prefs.theme || DEFAULT_THEME;
  const [active, setActive] = useState<boolean>(activeTheme === 'darkTheme');

  useEffect(() => {
    let theme = active ? 'darkTheme' : DEFAULT_THEME; // TODO
    document.documentElement.setAttribute('data-theme', theme);
    dispatch(updateTheme(theme));
  }, [active]);

  return (
    <div>
      <div
        className={classnames('custom-control custom-switch', styles.switch, {
          [styles.darkSwitch]: !isUndefined(props.darkVersion) && props.darkVersion,
        })}
      >
        <input
          id="darkMode"
          type="checkbox"
          className={`custom-control-input ${styles.checkbox}`}
          onChange={() => setActive(!active)}
          checked={active}
        />
        <label className={`custom-control-label ${styles.label}`} htmlFor="darkMode">
          <IoIosMoon />
        </label>
      </div>
    </div>
  );
};

export default DarkMode;
