/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

import './style.css';
import {GetWebsocketUrl, SetPtySize} from '../wailsjs/go/main/App';
import {Terminal} from "@xterm/xterm";
import {FitAddon} from '@xterm/addon-fit';

import {Environment} from '../wailsjs/runtime'

window.wsUrl = await GetWebsocketUrl();

import '../node_modules/@xterm/xterm/css/xterm.css';

import './app.css';

let environment = await Environment()

document.querySelector('#app').innerHTML = '';

if (environment.platform === 'darwin') {
    document.querySelector('#app').innerHTML += `
        <div id="header" style="--wails-draggable:drag"></div>
    `;
}

document.querySelector('#app').innerHTML += `
    <div id="terminal-container">
        <div id="terminal"></div>
    </div>
`;

const term = new Terminal({cursorBlink: true});
const fitAddon = new FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));
fitAddon.fit();

let xtermSize = fitAddon.proposeDimensions()
SetPtySize(xtermSize.rows, xtermSize.cols)

setInterval(() => {
    fitAddon.fit();
    let xtermSize = fitAddon.proposeDimensions()
    SetPtySize(xtermSize.rows, xtermSize.cols)
}, 2000)

const textDecoder = new TextDecoder();
const socket = new WebSocket(window.wsUrl);
socket.binaryType = 'arraybuffer';

socket.onopen = function () {
    term.focus()
    console.log('websocket opened');
};

socket.onmessage = function(event) {
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

term.onData(function(data) {
    socket.send(data);
});
