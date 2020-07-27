import React, { useState, useEffect } from 'react';
import classnames from 'classnames';

import styles from './DarkMode.module.css';
import { IoIosMoon } from 'react-icons/io';
import { isUndefined } from 'lodash';

interface Props {
  darkVersion?: boolean;
}

const DEFAULT_THEME = 'theme';

const DarkMode = (props: Props) => {
  const [active, setActive] = useState<boolean>(false);
  const [activeTheme, setActiveTheme] = useState<string>(DEFAULT_THEME);

  const switchTheme = (newTheme: string) => {
    const dom = document.getElementById(`atifactHubStyle-${activeTheme}`);
    if (dom) {
      dom.remove();
    }

    const themeAssetId = `atifactHubStyle-${newTheme}`;
    if (!document.getElementById(themeAssetId)) {
      const style = document.createElement('link');
      style.type = 'text/css';
      style.rel = 'stylesheet';
      style.id = `atifactHubStyle-${newTheme}`;
      style.href = `../../themes/${newTheme}.scss`;
      document.body.append(style);
      document.body.setAttribute('data-theme', newTheme);
      setActiveTheme(newTheme);
    }
  };

  useEffect(() => {
    let theme = active ? 'darkTheme' : DEFAULT_THEME;
    document.documentElement.setAttribute('data-theme', theme);

    // import(`../../themes/${theme}.scss`).then(() => {
    //   console.log('entro import', theme);
    //   return;
    // });
    // switchTheme(theme);
  }, [active]);

  // import(`../styles/darkTheme.scss`).then(() => {
  //   return;
  // });

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
