/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

import './style.css';
import './app.css';
import {main} from "../wailsjs/go/models";
import {GetPlatform} from "../wailsjs/go/main/App";
import {WindowIsFullscreen} from "../wailsjs/runtime";

document.querySelector('#app').innerHTML = '';

window.withHeader = false;

const platform = await GetPlatform();
if (platform === main.Platform.MacOs) {
    window.withHeader = true;
    document.querySelector('#app').innerHTML += `
        <div id="header" style="--wails-draggable:drag"></div>
    `;
}

window.isFullscreen = async () => {
    return await WindowIsFullscreen();
}

const isFirstLaunch = () => {
    return localStorage.getItem('launched') <= 0;
}

if (isFirstLaunch()) {
    import('./welcome')
} else {
    import('./terminal')
}
