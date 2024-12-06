/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

import './style.css';
import {GetWebsocketUrl, SetPtySize, GetPlatform, GetTerminalTheme, GetTerminalFontConfig} from '../wailsjs/go/main/App';
import {main} from '../wailsjs/go/models';
import {Terminal} from "@xterm/xterm";
import {FitAddon} from '@xterm/addon-fit';
import {WebLinksAddon} from '@xterm/addon-web-links';

import {BrowserOpenURL} from '../wailsjs/runtime'

const wsUrl = await GetWebsocketUrl();

import '../node_modules/@xterm/xterm/css/xterm.css';
import './app.css';

let fontName = null;
const fontConfig = await GetTerminalFontConfig()

switch (fontConfig.Family) {
    default:
    case main.TerminalFontFamily.FiraCode:
        fontName = 'Fira Code';
        break;
}

document.querySelector('#app').innerHTML = '';

const platform = await GetPlatform();

if (platform === main.Platform.MacOs) {
    document.querySelector('#app').innerHTML += `
        <div id="header" style="--wails-draggable:drag"></div>
    `;
}

document.querySelector('#app').innerHTML += `
    <div id="terminal-container">
        <div id="terminal"></div>
    </div>
`;

import {OneHalfDark, OneHalfLight} from './themes';

let lightTheme = null
let darkTheme = null

const selectedTheme = await GetTerminalTheme()

switch (selectedTheme) {
    default:
    case main.TerminalTheme.OneHalf:
        lightTheme = OneHalfLight;
        darkTheme = OneHalfDark;
        break;
}

function getCurrentTheme(event) {
    const newColorScheme = event.matches ? "dark" : "light";

    if (newColorScheme === 'dark') {
        return darkTheme
    } else {
        return lightTheme
    }
}

function applyTheme(event) {
    term.options.theme = getCurrentTheme(event)
}

let startTheme = getCurrentTheme(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)'))

const term = new Terminal({
    cursorBlink: true,
    theme: startTheme,
    fontFamily: `"${fontName}", monospace`,
    fontWeight: fontConfig.Weight,
    fontWeightBold: fontConfig.WeightBold,
    fontSize: fontConfig.Size,
});
const fitAddon = new FitAddon();
term.loadAddon(fitAddon);
term.loadAddon(new WebLinksAddon((event, uri) => {
    let openLinkTargetKey = Platform === Platforms.MacOS ? event.metaKey : event.ctrlKey;

    if (
        event.type === 'mouseup' &&
        event.button === 0 &&
        openLinkTargetKey
    ) {
        BrowserOpenURL(uri);
    }
}))
term.open(document.getElementById('terminal'));
fitAddon.fit();

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme);

let xtermSize = fitAddon.proposeDimensions()
SetPtySize(xtermSize.rows, xtermSize.cols)

setInterval(() => {
    fitAddon.fit();
    let xtermSize = fitAddon.proposeDimensions()
    SetPtySize(xtermSize.rows, xtermSize.cols)
}, 2000)

const textDecoder = new TextDecoder();
const socket = new WebSocket(wsUrl);
socket.binaryType = 'arraybuffer';

socket.onopen = function () {
    term.focus()
    console.log('websocket opened');
};

socket.onmessage = function (event) {
    if (typeof event.data === 'string') {
        term.write(event.data);
    } else {
        const text = textDecoder.decode(new Uint8Array(event.data));
        term.write(text);
    }
};

socket.onerror = function (event) {
    console.error('WebSocket error:', event);
};

socket.onclose = function () {
    console.log('websocket closed');
};

term.onData(function (data) {
    socket.send(data);
});
